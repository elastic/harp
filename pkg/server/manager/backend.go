// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package manager

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/gosimple/slug"

	"github.com/elastic/harp/pkg/sdk/value/encryption"
	"github.com/elastic/harp/pkg/server/storage"
	valueDecorator "github.com/elastic/harp/pkg/server/storage/decorators/value"
)

// Backend declares backend manager contract.
type Backend interface {
	GetSecret(context.Context, string, string) ([]byte, error)
	Register(context.Context, string, string) error
	GetNameSpace(context.Context, string) (storage.Engine, error)
}

var (
	// ErrNamespaceNotFound is raised when namespace is not registered
	ErrNamespaceNotFound = errors.New("manager: namespace not found")
	// ErrNamespaceAlreadyRegistered is raised when namespace is already registered
	ErrNamespaceAlreadyRegistered = errors.New("manager: namespace already registered")
)

// Default returns the default backend manager instance
func Default() Backend {
	return &backendManager{
		backends: map[string]storage.Engine{},
	}
}

// -----------------------------------------------------------------------------

type backendManager struct {
	sync.RWMutex
	backends map[string]storage.Engine
}

func (bm *backendManager) GetSecret(ctx context.Context, namespace, identifier string) ([]byte, error) {
	// Check backend registration
	engine, err := bm.GetNameSpace(ctx, namespace)
	if err != nil {
		return nil, err
	}

	// Delegate to engine
	return engine.Get(ctx, identifier)
}

func (bm *backendManager) Register(ctx context.Context, namespace, uri string) error {
	// Check backend registration
	_, err := bm.GetNameSpace(ctx, namespace)
	if err == nil {
		return ErrNamespaceAlreadyRegistered
	}

	// Load backend settings
	engine, err := storage.Build(uri)
	if err != nil {
		return fmt.Errorf("unable to build secret backend (%s:%s): %w", namespace, uri, err)
	}

	// Add encryption backend
	engine, err = wrapEncryptionEngine(uri, engine)
	if err != nil {
		return fmt.Errorf("unable to wrap encryption engine with secret backend (%s:%s): %w", namespace, uri, err)
	}

	// Add to backend map
	bm.Lock()
	bm.backends[clean(namespace)] = engine
	bm.Unlock()

	// Return no error
	return nil
}

func (bm *backendManager) GetNameSpace(ctx context.Context, namespace string) (storage.Engine, error) {
	// Lock read
	bm.RLock()
	defer bm.RUnlock()

	// Check backend registration
	engine, ok := bm.backends[clean(namespace)]
	if !ok {
		return nil, ErrNamespaceNotFound
	}

	// No error
	return engine, nil
}

// -----------------------------------------------------------------------------

func clean(ns string) string {
	// Remove any starting "/"
	ns = strings.TrimPrefix(ns, "/")
	// Slugify
	return slug.Make(ns)
}

func wrapEncryptionEngine(uri string, engine storage.Engine) (storage.Engine, error) {
	// Parse URL first
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("backend: unable to parse backend url: %w", err)
	}

	// Parse parameters to wrap engine with at-rest / in-transit encryption
	var (
		q       = u.Query()
		keyRaw  = q.Get("key")
		decrypt = q.Get("enc_revert")
	)

	// key is defined
	if keyRaw != "" {
		transformer, err := encryption.FromKey(keyRaw)
		if err != nil {
			return nil, fmt.Errorf("unable to initialize secret encryption: %w", err)
		}

		// Wrap engine using transformer decorator
		engine = valueDecorator.Transformer(transformer, decrypt == "true")(engine)
	}

	// No error
	return engine, nil
}
