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
	"time"

	"github.com/elastic/harp/pkg/kv/etcd3"
	"github.com/elastic/harp/test/integration/resource"
	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// -----------------------------------------------------------------------------

func TestWithEtcd(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create zk instance
	kvURI := resource.Etcd(ctx, t)

	// Create zk client
	client, errClient := clientv3.New(clientv3.Config{
		Endpoints:   []string{kvURI},
		DialTimeout: 5 * time.Second,
	})
	assert.NoError(t, errClient)
	assert.NotNil(t, client)

	// Initialize KV Store
	s := etcd3.Store(client)

	// Run test suite
	t.Run("store", testSuite(ctx, s))
}
