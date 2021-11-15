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

//go:build integration

package kv

import (
	"context"
	"testing"

	"github.com/elastic/harp/pkg/kv"
	"github.com/stretchr/testify/assert"
)

func testSuite(ctx context.Context, s kv.Store) func(t *testing.T) {
	return func(t *testing.T) {
		assert.NotNil(t, s)

		// Check if empty
		pairs, err := s.List(ctx, "app")
		assert.Error(t, err)
		assert.ErrorIs(t, err, kv.ErrKeyNotFound)
		assert.Nil(t, pairs)

		// Create keys
		err = s.Put(ctx, "app/production/customer1/ece/v1.0.0/adminconsole/database/usage_credentials/host", []byte("InNhbXBsZS1pbnN0YW5jZS5hYmMyZGVmZ2hpamUudXMtd2VzdC0yLnJkcy5hbWF6b25hd3MuY29tIg=="))
		assert.NoError(t, err)

		// Retrieve the key
		pair, err := s.Get(ctx, "app/production/customer1/ece/v1.0.0/adminconsole/database/usage_credentials/host")
		assert.NoError(t, err)
		assert.NotNil(t, pair)
		assert.Equal(t, []byte("InNhbXBsZS1pbnN0YW5jZS5hYmMyZGVmZ2hpamUudXMtd2VzdC0yLnJkcy5hbWF6b25hd3MuY29tIg=="), pair.Value)
		assert.Equal(t, "app/production/customer1/ece/v1.0.0/adminconsole/database/usage_credentials/host", pair.Key)

		// List elements
		pairs, err = s.List(ctx, "app")
		assert.NoError(t, err)
		assert.NotNil(t, pairs)
		assert.Len(t, pairs, 1)

		// Create another keys
		err = s.Put(ctx, "platform/production/customer1/us-east-1/zookeeper/accounts/admin_credentials", []byte("zkadmin-h8HB5AKi"))
		assert.NoError(t, err)

		// List elements
		pairs, err = s.List(ctx, "app")
		assert.NoError(t, err)
		assert.NotNil(t, pairs)
		assert.Len(t, pairs, 1)

		// List elements
		pairs, err = s.List(ctx, "platform")
		assert.NoError(t, err)
		assert.NotNil(t, pairs)
		assert.Len(t, pairs, 1)

		// Check existence
		exists, err := s.Exists(ctx, "non-existent")
		assert.NoError(t, err)
		assert.False(t, exists)

		exists, err = s.Exists(ctx, "platform/production/customer1/us-east-1/zookeeper/accounts/admin_credentials")
		assert.NoError(t, err)
		assert.True(t, exists)

		// Delete
		err = s.Delete(ctx, "platform/production/customer1/us-east-1/zookeeper/accounts/admin_credentials")
		assert.NoError(t, err)

		exists, err = s.Exists(ctx, "platform/production/customer1/us-east-1/zookeeper/accounts/admin_credentials")
		assert.NoError(t, err)
		assert.False(t, exists)
	}
}
