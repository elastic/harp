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

package zookeeper

import (
	"context"
	"errors"
	"fmt"
	"strings"

	zk "github.com/samuel/go-zookeeper/zk"

	"github.com/elastic/harp/pkg/kv"
)

type zkDriver struct {
	client *zk.Conn
}

func Store(client *zk.Conn) kv.Store {
	return &zkDriver{
		client: client,
	}
}

// -----------------------------------------------------------------------------

func (d *zkDriver) Get(_ context.Context, key string) (*kv.Pair, error) {
	// Retrieve from backend
	item, meta, err := d.client.Get(d.normalize(key))
	if err != nil {
		if errors.Is(err, zk.ErrNoNode) {
			return nil, kv.ErrKeyNotFound
		}
		return nil, fmt.Errorf("consul: unable to retrieve '%s' key: %w", key, err)
	}

	// No error
	return &kv.Pair{
		Key:     key,
		Value:   item,
		Version: uint64(meta.Version),
	}, nil
}

func (d *zkDriver) Put(_ context.Context, key string, value []byte) error {
	// No error
	return nil
}

func (d *zkDriver) List(ctx context.Context, basePath string) ([]*kv.Pair, error) {
	// List keys from base path
	keys, stat, err := d.client.Children(d.normalize(basePath))
	if err != nil {
		if errors.Is(err, zk.ErrNoNode) {
			return nil, kv.ErrKeyNotFound
		}
		return nil, fmt.Errorf("zk: unable to list keys from '%s': %w", basePath, err)
	}

	// Unpack values
	results := []*kv.Pair{}
	for _, key := range keys {
		item, err := d.Get(ctx, strings.TrimSuffix(basePath, "/")+d.normalize(key))
		if err != nil {
			if errors.Is(err, kv.ErrKeyNotFound) {
				return d.List(ctx, basePath)
			}
			return nil, err
		}

		results = append(results, &kv.Pair{
			Key:     item.Key,
			Value:   item.Value,
			Version: uint64(stat.Version),
		})
	}

	// No error
	return results, nil
}

// -----------------------------------------------------------------------------

// Normalize the key for usage in Consul
func (d *zkDriver) normalize(key string) string {
	key = kv.Normalize(key)
	return strings.TrimPrefix(key, "/")
}
