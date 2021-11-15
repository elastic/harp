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

	zk "github.com/go-zookeeper/zk"

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
		return nil, fmt.Errorf("zk: unable to retrieve '%s' key: %w", key, err)
	}

	// No error
	return &kv.Pair{
		Key:     key,
		Value:   item,
		Version: uint64(meta.Version),
	}, nil
}

func (d *zkDriver) Put(ctx context.Context, key string, value []byte) error {
	// Check if key exists
	exists, err := d.Exists(ctx, key)
	if err != nil {
		return err
	}
	if !exists {
		// Create full hierarchy if the key doesn't exists
		d.createFullPath(kv.SplitKey(strings.TrimSuffix(key, "/")))
	}

	// Set the value (last version)
	_, err = d.client.Set(d.normalize(key), value, -1)
	if err != nil {
		return fmt.Errorf("zk: unable to set '%s' value: %w", key, err)
	}

	// No error
	return nil
}

func (d *zkDriver) Delete(_ context.Context, key string) error {
	// Try to delete from store.
	err := d.client.Delete(d.normalize(key), -1)
	if err != nil {
		if errors.Is(err, zk.ErrNoNode) {
			return kv.ErrKeyNotFound
		}
		return fmt.Errorf("zk: unable to delete '%s': %w", key, err)
	}

	// No error
	return nil
}

func (d *zkDriver) Exists(_ context.Context, key string) (bool, error) {
	key = d.normalize(key)

	exists, _, err := d.client.Exists(key)
	if err != nil {
		return false, fmt.Errorf("zk: unable to check key '%s' existence: %w", key, err)
	}

	// No error
	return exists, nil
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

func (d *zkDriver) Close() error {
	// Skip if client instance is nil
	if d.client == nil {
		return nil
	}

	// Close the client connection.
	d.client.Close()

	// No error
	return nil
}

// -----------------------------------------------------------------------------

// Normalize the key for usage in Consul
func (d *zkDriver) normalize(key string) string {
	key = kv.Normalize(key)
	return strings.TrimSuffix(key, "/")
}

// -----------------------------------------------------------------------------

// createFullPath creates the entire path for a directory
// that does not exist
func (d *zkDriver) createFullPath(path []string) error {
	for i := 1; i <= len(path); i++ {
		newpath := "/" + strings.Join(path[:i], "/")
		_, err := d.client.Create(newpath, []byte{}, 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			// Skip if node already exists
			if err != zk.ErrNodeExists {
				return err
			}
		}
	}
	return nil
}
