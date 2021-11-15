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

package encryption

import (
	"fmt"

	"github.com/elastic/harp/pkg/sdk/value"
)

// TransformerFactoryFunc is used for transformer building for encryption.
type TransformerFactoryFunc func(string) (value.Transformer, error)

var (
	registry map[string]TransformerFactoryFunc
)

// Register a transformer with the given prefix.
func Register(prefix string, factory TransformerFactoryFunc) {
	// Lazy initialization
	if registry == nil {
		registry = map[string]TransformerFactoryFunc{}
	}

	// Check if not already registered
	if _, ok := registry[prefix]; ok {
		panic(fmt.Errorf("encryption transformer already registered fro '%s' prefix", prefix))
	}

	// Register the transformer
	registry[prefix] = factory
}
