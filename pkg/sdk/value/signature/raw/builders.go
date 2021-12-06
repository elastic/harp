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

package raw

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/square/go-jose.v2"

	"github.com/elastic/harp/pkg/sdk/value"
	"github.com/elastic/harp/pkg/sdk/value/signature"
)

func init() {
	signature.Register("raw", FromKey)
}

// FromKey returns an encryption transformer instance according to the given key format.
func FromKey(key string) (value.Transformer, error) {
	// Remove the prefix
	key = strings.TrimPrefix(key, "raw:")

	// Delegate to builder
	return Transformer(key)
}

// Transformer returns a JWS signature value transformer instance.
func Transformer(key string) (value.Transformer, error) {
	// Decode key
	keyRaw, err := base64.RawURLEncoding.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("unable to decode transformer key: %w", err)
	}

	// Check JWK encoding
	var jwk jose.JSONWebKey
	if errJSON := json.Unmarshal(keyRaw, &jwk); errJSON != nil {
		return nil, fmt.Errorf("unable to decode the transformer key: %w", errJSON)
	}

	// Delegate to transformer
	return &rawTransformer{
		key: jwk.Key,
	}, err
}
