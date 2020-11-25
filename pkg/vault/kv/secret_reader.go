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
	"fmt"

	"github.com/hashicorp/vault/api"
)

// SecretGetter pull a secret from Vault using given path.
//
// To be used of template function.
func SecretGetter(client *api.Client) func(string) (map[string]interface{}, error) {
	return func(path string) (map[string]interface{}, error) {
		// Create dedicated service reader
		service, err := New(client, path)
		if err != nil {
			return nil, fmt.Errorf("unable to prepare vault reader for path '%s': %w", path, err)
		}

		// Delegate to reader
		return service.Read(context.Background(), path)
	}
}
