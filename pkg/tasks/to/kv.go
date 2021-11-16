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

package to

import (
	"context"
	"encoding/json"
	"fmt"
	"path"

	"github.com/elastic/harp/pkg/bundle"
	"github.com/elastic/harp/pkg/kv"
	"github.com/elastic/harp/pkg/tasks"
)

type PublishKVTask struct {
	_               struct{}
	ContainerReader tasks.ReaderProvider
	Store           kv.Store
	SecretAsKey     bool
	Prefix          string
}

func (t *PublishKVTask) Run(ctx context.Context) error {
	// Create the reader
	reader, err := t.ContainerReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to open input bundle reader: %w", err)
	}

	// Extract bundle from container
	b, err := bundle.FromContainerReader(reader)
	if err != nil {
		return fmt.Errorf("unable to load bundle: %w", err)
	}

	// Convert as map
	bundleMap, err := bundle.AsMap(b)
	if err != nil {
		return fmt.Errorf("unable to transform the bundle as a map: %w", err)
	}

	// Foreach element in the bundle map.
	for key, value := range bundleMap {
		if t.Prefix != "" {
			key = path.Join(path.Clean(t.Prefix), key)
		}
		if !t.SecretAsKey {
			// Encode as json
			payload, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("unable to encode value as JSON for '%s': %w", key, err)
			}

			// Insert in KV store.
			if err := t.Store.Put(ctx, key, payload); err != nil {
				return fmt.Errorf("unable to publish '%s' secret in store: %w", key, err)
			}
		} else {
			// Range over secrets
			secrets, ok := value.(bundle.KV)
			if !ok {
				continue
			}

			// Publish each secret as a leaf.
			for secKey, secValue := range secrets {
				// Insert in KV store.
				if err := t.Store.Put(ctx, path.Join(key, secKey), []byte(fmt.Sprintf("%v", secValue))); err != nil {
					return fmt.Errorf("unable to publish '%s' secret in store: %w", key, err)
				}
			}
		}
	}

	// No error
	return nil
}
