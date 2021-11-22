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
	"errors"
	"fmt"
	"strings"

	api "github.com/hashicorp/consul/api"

	"github.com/elastic/harp/pkg/kv"
	"github.com/elastic/harp/pkg/sdk/types"
)

type consulDriver struct {
	client Client
}

func Store(client Client) kv.Store {
	return &consulDriver{
		client: client,
	}
}

// -----------------------------------------------------------------------------

func (d *consulDriver) Get(_ context.Context, key string) (*kv.Pair, error) {
	// Check arguments
	if types.IsNil(d.client) {
		return nil, errors.New("consul: unable to query with nil client")
	}

	// Retrieve from backend
	item, meta, err := d.client.Get(d.normalize(key), &api.QueryOptions{
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
	// Check arguments
	if types.IsNil(d.client) {
		return errors.New("consul: unable to query with nil client")
	}

	// Prepare the item to put
	item := &api.KVPair{
		Key:   d.normalize(key),
		Value: value,
	}

	// Delegate to client
	if _, err := d.client.Put(item, nil); err != nil {
		return fmt.Errorf("consul: unable to put '%s' value: %w", key, err)
	}

	// No error
	return nil
}

func (d *consulDriver) Delete(ctx context.Context, key string) error {
	// Check arguments
	if types.IsNil(d.client) {
		return errors.New("consul: unable to query with nil client")
	}

	// Retrieve from store
	found, err := d.Exists(ctx, key)
	if err != nil {
		return fmt.Errorf("consul: unable to retrieve '%s' for deletion: %w", key, err)
	}
	if !found {
		return kv.ErrKeyNotFound
	}

	// Delete the value
	if _, err := d.client.Delete(d.normalize(key), nil); err != nil {
		return fmt.Errorf("consul: unable to delete '%s': %w", key, err)
	}

	// No error
	return nil
}

func (d *consulDriver) Exists(ctx context.Context, key string) (bool, error) {
	// Retrieve from stroe
	_, err := d.Get(ctx, key)
	if err != nil {
		if errors.Is(err, kv.ErrKeyNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("consul: unable to check key '%s' existence: %w", key, err)
	}

	// No error
	return true, nil
}

func (d *consulDriver) List(_ context.Context, basePath string) ([]*kv.Pair, error) {
	// Check arguments
	if types.IsNil(d.client) {
		return nil, errors.New("consul: unable to query with nil client")
	}

	// List keys from base path
	items, _, err := d.client.List(d.normalize(basePath), nil)
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

func (d *consulDriver) Close() error {
	// No error
	return nil
}

// -----------------------------------------------------------------------------

// Normalize the key for usage in Consul
func (d *consulDriver) normalize(key string) string {
	key = kv.Normalize(key)
	return strings.TrimPrefix(key, "/")
}
