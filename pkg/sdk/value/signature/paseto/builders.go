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

package paseto

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gopkg.in/square/go-jose.v2"

	"github.com/elastic/harp/pkg/sdk/value"
	"github.com/elastic/harp/pkg/sdk/value/signature"
)

func init() {
	signature.Register("paseto", FromKey)
}

// FromKey returns an encryption transformer instance according to the given key format.
func FromKey(key string) (value.Transformer, error) {
	// Remove the prefix
	key = strings.TrimPrefix(key, "paseto:")

	// Split components
	parts := strings.SplitN(key, ":", 2)
	if len(parts) != 2 {
		return nil, errors.New("paseto: invalid key format")
	}

	// Delegate to builder
	return Transformer(parts[0], parts[1])
}

// Transformer returns a JWS signature value transformer instance.
func Transformer(algorithm, key string) (value.Transformer, error) {
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

	// Select appropriate strategy
	switch algorithm {
	case "v3":
		return &pasetoTransformer{
			key: jwk.Key,
		}, err
	case "v4":
		return &pasetoTransformer{
			key: jwk.Key,
		}, err
	default:
	}

	// Unsupported encryption scheme.
	return nil, fmt.Errorf("unsupported jws algorithm '%s'", algorithm)
}
