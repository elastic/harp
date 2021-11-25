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
	"encoding/base64"
	"fmt"
	"strings"

	pasetov4 "github.com/elastic/harp/pkg/sdk/security/crypto/paseto/v4"
	"github.com/elastic/harp/pkg/sdk/value"
	"github.com/elastic/harp/pkg/sdk/value/encryption"
)

func init() {
	encryption.Register("paseto", Transformer)
}

func Transformer(key string) (value.Transformer, error) {
	// Remove the prefix
	key = strings.TrimPrefix(key, "paseto:")

	// Decode key
	k, err := base64.URLEncoding.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("paseto: unable to decode key: %w", err)
	}
	if l := len(k); l != pasetov4.KeyLength {
		return nil, fmt.Errorf("paseto: invalid secret key length (%d)", l)
	}

	// Copy secret key
	var secretKey [pasetov4.KeyLength]byte
	copy(secretKey[:], k)

	return &pasetoTransformer{
		key: secretKey,
	}, nil
}

// -----------------------------------------------------------------------------

type pasetoTransformer struct {
	key [pasetov4.KeyLength]byte
}

func (d *pasetoTransformer) From(_ context.Context, input []byte) ([]byte, error) {
	return pasetov4.Decrypt(d.key[:], input, "", "")
}

func (d *pasetoTransformer) To(_ context.Context, input []byte) ([]byte, error) {
	// Encrypt with paseto v4.local
	return pasetov4.Encrypt(d.key[:], input, "", "")
}
