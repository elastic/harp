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

package dae

import (
	"context"
	"crypto/cipher"
	"errors"
	"fmt"

	"github.com/elastic/harp/pkg/sdk/value/encryption"
)

// -----------------------------------------------------------------------------

type daeTransformer struct {
	aead             cipher.AEAD
	nonceDeriverFunc NonceDeriverFunc
}

func (t *daeTransformer) To(ctx context.Context, input []byte) ([]byte, error) {
	// Check input size
	if len(input) > 64*1024*1024 {
		return nil, errors.New("value too large")
	}

	// Derive nonce
	nonce, err := t.nonceDeriverFunc(input, t.aead)
	if err != nil {
		return nil, fmt.Errorf("dae: unable to derive nonce: %w", err)
	}
	if len(nonce) != t.aead.NonceSize() {
		return nil, errors.New("dae: derived nonce is too short")
	}

	// Retrieve additional data from context
	aad, _ := encryption.AdditionalData(ctx)

	// Seal the cleartext with deterministic nonce
	cipherText := t.aead.Seal(nil, nonce, input, aad)

	// Return encrypted value
	return append(nonce, cipherText...), nil
}

func (t *daeTransformer) From(ctx context.Context, input []byte) ([]byte, error) {
	// Check input size
	if len(input) < t.aead.NonceSize() {
		return nil, errors.New("dae: ciphered text too short")
	}

	nonce := input[:t.aead.NonceSize()]
	text := input[t.aead.NonceSize():]
	aad, _ := encryption.AdditionalData(ctx)

	clearText, err := t.aead.Open(nil, nonce, text, aad)
	if err != nil {
		return nil, errors.New("failed to decrypt given message")
	}

	// No error
	return clearText, nil
}
