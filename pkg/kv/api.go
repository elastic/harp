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
package kv

import (
	"context"
	"errors"
)

var (
	// ErrKeyNotFound is raised when the given key could not be found in the store.
	ErrKeyNotFound = errors.New("key not found")
)

// Store describes the key/value store contract.
type Store interface {
	// Get the value stored at the given key.
	Get(ctx context.Context, key string) (*Pair, error)
	// Put the given value at the given key.
	Put(ctx context.Context, key string, value []byte) error
	// List subkeys at a given path
	List(ctx context.Context, path string) ([]*Pair, error)
}

// -----------------------------------------------------------------------------

type Pair struct {
	Key     string
	Value   []byte
	Version uint64
}
