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

package secretbox

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/elastic/harp/pkg/sdk/value"
)

// Transformer returns a Nacl SecretBox encryption value transformer
func Transformer(key string) (value.Transformer, error) {
	// Decode key
	k, err := base64.URLEncoding.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("secretbox: unable to decode key: %w", err)
	}
	if l := len(k); l != keyLength {
		return nil, fmt.Errorf("secretbox: invalid secret key length (%d)", l)
	}

	// Copy secret key
	secretKey := new([keyLength]byte)
	copy(secretKey[:], k)

	// Return transformer
	return &secretboxTransformer{
		key: secretKey,
	}, nil
}

// -----------------------------------------------------------------------------

type secretboxTransformer struct {
	key *[keyLength]byte
}

func (d *secretboxTransformer) From(_ context.Context, input []byte) ([]byte, error) {
	// Check output
	if l := len(input); l < nonceLength {
		return nil, fmt.Errorf("secretbox: invalid secret length (%d), check encryption status", l)
	}

	// Decrypt value
	out, err := decrypt(input, *d.key)
	if err != nil {
		return nil, fmt.Errorf("secretbox: unable to transform value: %w", err)
	}

	// No error
	return out, nil
}

func (d *secretboxTransformer) To(_ context.Context, input []byte) ([]byte, error) {
	// Encrypt value
	out, err := encrypt(input, *d.key)
	if err != nil {
		return nil, fmt.Errorf("secretbox: unable to transform value: %w", err)
	}

	// No error
	return out, nil
}
