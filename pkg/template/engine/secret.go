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

package engine

import (
	"fmt"
)

// SecretReaderFunc is a function to retrieve a secret from a given path.
type SecretReaderFunc func(path string) (map[string]interface{}, error)

// SecretReaders uses given secret reader funcs to resolve secret path.
func SecretReaders(secretReaders []SecretReaderFunc) func(string) (map[string]interface{}, error) {
	return func(secretPath string) (map[string]interface{}, error) {
		// For all secret readers
		for _, sr := range secretReaders {
			value, err := sr(secretPath)
			if err != nil {
				// Check next secret reader
				continue
			}

			// No error
			return value, nil
		}

		// Return error
		return nil, fmt.Errorf("no value found for '%s', check secret path or secret reader settings", secretPath)
	}
}
