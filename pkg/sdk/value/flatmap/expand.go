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

package flatmap

import (
	"strings"

	"github.com/elastic/harp/pkg/bundle"
)

// Expand takes a map and a key (prefix) and expands that value into
// a more complex structure. This is the reverse of the Flatten operation.
func Expand(m bundle.KV, key string) interface{} {
	// If the key is exactly a key in the map, just return it
	if v, ok := m[key]; ok {
		if v == "true" {
			return true
		} else if v == "false" {
			return false
		}

		return v
	}

	// Check if this is a prefix in the map
	prefix := key
	if key != "" {
		prefix = key + "/"
	}
	for k := range m {
		if strings.HasPrefix(k, prefix) {
			return expandMap(m, prefix)
		}
	}

	return nil
}

func expandMap(m bundle.KV, prefix string) bundle.KV {
	result := make(bundle.KV)
	for k := range m {
		if !strings.HasPrefix(k, prefix) {
			// Prefix not found
			continue
		}

		// Remove the prefix
		key := k[len(prefix):]
		idx := strings.Index(key, "/")
		if idx != -1 {
			key = key[:idx]
		}
		if _, ok := result[key]; ok {
			continue
		}

		// Recursive call to handle subtree
		result[key] = Expand(m, k[:len(prefix)+len(key)])
	}

	return result
}
