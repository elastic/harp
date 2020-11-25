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

package vault

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/hashicorp/vault/api"

	"github.com/elastic/harp/pkg/server/storage"
	"github.com/elastic/harp/pkg/vault/kv"
)

type engine struct {
	u        *url.URL
	basePath string

	client  *api.Client
	service kv.SecretReader
}

func build(u *url.URL) (storage.Engine, error) {
	// Initialize Vault connection
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, fmt.Errorf("vault: unable to initialize connection: %w", err)
	}

	// Retrieve a secret reader
	reader, err := kv.New(client, u.Path)
	if err != nil {
		return nil, err
	}

	// Build engine instance
	return &engine{
		u:        u,
		basePath: u.Path,
		client:   client,
		service:  reader,
	}, nil
}

func init() {
	// Register to storage factory
	storage.MustRegister("vault", build)
}

// -----------------------------------------------------------------------------

func (e *engine) Get(ctx context.Context, id string) ([]byte, error) {
	// Read from Vault
	secret, err := e.service.Read(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("vault: unable to read secret from vault server: %w", err)
	}
	if secret == nil {
		return []byte{}, storage.ErrSecretNotFound
	}

	// Encode secret as json
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(secret); err != nil {
		return nil, fmt.Errorf("vault: unable to encode secret: %w", err)
	}

	// Return secret
	return buf.Bytes(), nil
}
