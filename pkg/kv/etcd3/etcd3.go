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

package etcd3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/kv"
	"github.com/elastic/harp/pkg/sdk/log"
)

const (
	// ListBatchSize defines the pagination page size.
	ListBatchSize = 50
)

type etcd3Driver struct {
	client *clientv3.Client
}

func Store(client *clientv3.Client) kv.Store {
	return &etcd3Driver{
		client: client,
	}
}

// -----------------------------------------------------------------------------

func (d *etcd3Driver) Get(ctx context.Context, key string) (*kv.Pair, error) {
	// Retrieve key value
	resp, err := d.client.KV.Get(ctx, d.normalize(key), clientv3.WithLimit(1))
	if err != nil {
		return nil, fmt.Errorf("etcd3: unable to retrieve '%s' key: %w", key, err)
	}
	if resp == nil {
		return nil, fmt.Errorf("etcd3: got nil response for '%s'", key)
	}

	// Unpack result
	if len(resp.Kvs) == 0 {
		return nil, kv.ErrKeyNotFound
	}
	if len(resp.Kvs) > 1 {
		return nil, fmt.Errorf("etcd3: '%s' key returned multiple result where only one is expected", key)
	}

	// No error
	return &kv.Pair{
		Key:     string(resp.Kvs[0].Key),
		Value:   resp.Kvs[0].Value,
		Version: uint64(resp.Kvs[0].Version),
	}, nil
}

func (d *etcd3Driver) Put(ctx context.Context, key string, value []byte) error {
	// Put a value
	_, err := d.client.KV.Put(ctx, d.normalize(key), string(value))
	if err != nil {
		return fmt.Errorf("etcd3: unable to put '%s' value: %w", key, err)
	}

	// No error
	return nil
}

func (d *etcd3Driver) Delete(ctx context.Context, key string) error {
	// Try to delete from store
	resp, err := d.client.Delete(ctx, d.normalize(key))
	if err != nil {
		return fmt.Errorf("etcd3: unable to delete '%s' key: %w", key, err)
	}
	if resp == nil {
		return fmt.Errorf("etcd3: got nil response for '%s'", key)
	}
	if resp.Deleted == 0 {
		return kv.ErrKeyNotFound
	}

	// No error
	return nil
}

func (d *etcd3Driver) Exists(ctx context.Context, key string) (bool, error) {
	_, err := d.Get(ctx, key)
	if err != nil {
		if errors.Is(err, kv.ErrKeyNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("etcd3: unable to check key '%s' existence: %w", key, err)
	}

	// No error
	return true, nil
}

func (d *etcd3Driver) List(ctx context.Context, basePath string) ([]*kv.Pair, error) {
	log.For(ctx).Debug("etcd3: Try to list keys", zap.String("prefix", basePath))

	var (
		results = []*kv.Pair{}
		lastKey string
	)
	for {
		// Check if operation is ended
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Prepare query options
		opts := []clientv3.OpOption{
			clientv3.WithPrefix(),
			clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
			clientv3.WithLimit(ListBatchSize),
		}

		// If lastkey is defined set the cursor
		if lastKey != "" {
			opts = append(opts, clientv3.WithFromKey())
			basePath = lastKey
		}

		log.For(ctx).Debug("etcd3: Get all keys", zap.String("key", basePath))

		// Retrieve key value
		resp, err := d.client.KV.Get(ctx, d.normalize(basePath), opts...)
		if err != nil {
			return nil, fmt.Errorf("etcd3: unable to retrieve '%s' from base path: %w", basePath, err)
		}
		if resp == nil {
			return nil, fmt.Errorf("etcd3: got nil response for '%s'", basePath)
		}

		// Exit on empty result
		if len(resp.Kvs) == 0 {
			log.For(ctx).Debug("etcd3: No more result, stop.")
			break
		}

		// Unpack values
		for _, item := range resp.Kvs {
			log.For(ctx).Debug("etcd3: Unpack result", zap.String("key", string(item.Key)))

			// Skip first if lastKey is defined
			if lastKey != "" && bytes.Equal(item.Key, []byte(lastKey)) {
				continue
			}
			results = append(results, &kv.Pair{
				Key:     string(item.Key),
				Value:   item.Value,
				Version: uint64(item.Version),
			})
		}

		// No need to paginate
		if len(resp.Kvs) < ListBatchSize {
			break
		}

		// Retrieve last key
		lastKey = string(resp.Kvs[len(resp.Kvs)-1].Key)
	}

	// Raise keynotfound if no result.
	if len(results) == 0 {
		return nil, kv.ErrKeyNotFound
	}

	// No error
	return results, nil
}

func (d *etcd3Driver) Close() error {
	// Skip if client instance is nil
	if d.client == nil {
		return nil
	}

	// Try to close client connection.
	if err := d.client.Close(); err != nil {
		return fmt.Errorf("etcd3: unable to close client connection: %w", err)
	}

	// No error
	return nil
}

// -----------------------------------------------------------------------------

// Normalize the key for usage in Consul
func (d *etcd3Driver) normalize(key string) string {
	key = kv.Normalize(key)
	return strings.TrimPrefix(key, "/")
}
