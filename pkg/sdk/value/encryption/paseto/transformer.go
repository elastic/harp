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
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/elastic/harp/pkg/sdk/value"
	"github.com/elastic/harp/pkg/sdk/value/encryption"
)

func init() {
	encryption.Register("paseto", Transformer)
}

const (
	keyLength     = 32
	v4LocalPrefix = "v4.local."
)

func Transformer(key string) (value.Transformer, error) {
	// Remove the prefix
	key = strings.TrimPrefix(key, "paseto:")

	// Decode key
	k, err := base64.URLEncoding.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("paseto: unable to decode key: %w", err)
	}
	if l := len(k); l != keyLength {
		return nil, fmt.Errorf("paseto: invalid secret key length (%d)", l)
	}

	// Copy secret key
	var secretKey [keyLength]byte
	copy(secretKey[:], k)

	return &pasetoTransformer{
		key: secretKey,
	}, nil
}

// -----------------------------------------------------------------------------

type pasetoTransformer struct {
	key [keyLength]byte
}

func (d *pasetoTransformer) From(_ context.Context, input []byte) ([]byte, error) {
	return decrypt(d.key[:], input, "", "")
}

func (d *pasetoTransformer) To(_ context.Context, input []byte) ([]byte, error) {
	// Create random seed
	var n [32]byte
	if _, err := rand.Read(n[:]); err != nil {
		return nil, fmt.Errorf("paseto: unable to generate random seed: %w", err)
	}

	// Encrypt with paseto v4.local
	return encrypt(d.key[:], n[:], input, "", "")
}
