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

package jwe

import (
	"encoding/base64"
	"fmt"
	"strings"

	"gopkg.in/square/go-jose.v2"

	"github.com/elastic/harp/pkg/sdk/value"
	"github.com/elastic/harp/pkg/sdk/value/encryption"
)

func init() {
	encryption.Register("jwe", Transformer)
}

// Transformer returns an encryption transformer instance according to the given key format.
func Transformer(key string) (value.Transformer, error) {
	// Remove the prefix
	key = strings.TrimPrefix(key, "jwe:")

	switch {
	case strings.HasPrefix(key, "a128kw:"):
		k, err := base64.URLEncoding.DecodeString(strings.TrimPrefix(key, "a128kw:"))
		if err != nil {
			return nil, fmt.Errorf("jwe: unable to decode key: %w", err)
		}
		if len(k) < 16 {
			return nil, fmt.Errorf("jwe: key too short: %w", err)
		}
		return transformer(k, jose.A128KW, jose.A128GCM)
	case strings.HasPrefix(key, "a192kw:"):
		k, err := base64.URLEncoding.DecodeString(strings.TrimPrefix(key, "a192kw:"))
		if err != nil {
			return nil, fmt.Errorf("jwe: unable to decode key: %w", err)
		}
		if len(k) < 24 {
			return nil, fmt.Errorf("jwe: key too short: %w", err)
		}
		return transformer(k, jose.A192KW, jose.A192GCM)
	case strings.HasPrefix(key, "a256kw:"):
		k, err := base64.URLEncoding.DecodeString(strings.TrimPrefix(key, "a256kw:"))
		if err != nil {
			return nil, fmt.Errorf("jwe: unable to decode key: %w", err)
		}
		if len(k) < 32 {
			return nil, fmt.Errorf("jwe: key too short: %w", err)
		}
		return transformer(k, jose.A256KW, jose.A256GCM)
	case strings.HasPrefix(key, "pbes2-a128kw:"):
		return transformer(strings.TrimPrefix(key, "pbes2-a128kw:"), jose.PBES2_HS256_A128KW, jose.A128GCM)
	case strings.HasPrefix(key, "pbes2-a192kw:"):
		return transformer(strings.TrimPrefix(key, "pbes2-a192kw:"), jose.PBES2_HS384_A192KW, jose.A192GCM)
	case strings.HasPrefix(key, "pbes2-a256kw:"):
		return transformer(strings.TrimPrefix(key, "pbes2-256kw:"), jose.PBES2_HS512_A256KW, jose.A256GCM)
	default:
	}

	// Unsupported encryption scheme.
	return nil, fmt.Errorf("unsupported jwe algorithm for key '%s'", key)
}
