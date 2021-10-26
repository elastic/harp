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

package consul

import (
	"context"
	"fmt"
	"strings"

	api "github.com/hashicorp/consul/api"

	"github.com/elastic/harp/pkg/kv"
)

type consulDriver struct {
	client *api.Client
}

func Store(client *api.Client) kv.Store {
	return &consulDriver{
		client: client,
	}
}

// -----------------------------------------------------------------------------

func (d *consulDriver) Get(_ context.Context, key string) (*kv.Pair, error) {
	// Retrieve from backend
	item, meta, err := d.client.KV().Get(d.normalize(key), &api.QueryOptions{
		AllowStale:        false,
		RequireConsistent: true,
	})
	if err != nil {
		return nil, fmt.Errorf("consul: unable to retrieve '%s' key: %w", key, err)
	}
	if item == nil {
		return nil, kv.ErrKeyNotFound
	}

	// No error
	return &kv.Pair{
		Key:     item.Key,
		Value:   item.Value,
		Version: meta.LastIndex,
	}, nil
}

func (d *consulDriver) Put(_ context.Context, key string, value []byte) error {
	// Prepare the item to put
	item := &api.KVPair{
		Key:   d.normalize(key),
		Value: value,
	}

	// Delegate to client
	if _, err := d.client.KV().Put(item, nil); err != nil {
		return fmt.Errorf("consul: unable to put '%s' value: %w", key, err)
	}

	// No error
	return nil
}

func (d *consulDriver) List(_ context.Context, basePath string) ([]*kv.Pair, error) {
	// List keys from base path
	items, _, err := d.client.KV().List(d.normalize(basePath), nil)
	if err != nil {
		return nil, fmt.Errorf("consul: unable to list keys from '%s': %w", basePath, err)
	}
	if len(items) == 0 {
		return nil, kv.ErrKeyNotFound
	}

	// Unpack values
	results := []*kv.Pair{}
	for _, item := range items {
		// Skip first item as base path
		if item.Key == basePath {
			continue
		}
		results = append(results, &kv.Pair{
			Key:     item.Key,
			Value:   item.Value,
			Version: item.ModifyIndex,
		})
	}

	// No error
	return results, nil
}

// -----------------------------------------------------------------------------

// Normalize the key for usage in Consul
func (d *consulDriver) normalize(key string) string {
	key = kv.Normalize(key)
	return strings.TrimPrefix(key, "/")
}
