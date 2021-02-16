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

package aead

import (
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

const (
	keyLength = 32
)

func encrypt(plaintext []byte, ciph cipher.AEAD) ([]byte, error) {
	if len(plaintext) > 64*1024*1024 {
		return nil, errors.New("value too large")
	}
	nonce := make([]byte, ciph.NonceSize(), ciph.NonceSize()+ciph.Overhead()+len(plaintext))
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("unable to generate nonce: %w", err)
	}

	cipherText := ciph.Seal(nil, nonce, plaintext, nil)

	return append(nonce, cipherText...), nil
}

func decrypt(ciphertext []byte, ciph cipher.AEAD) ([]byte, error) {
	if len(ciphertext) < ciph.NonceSize() {
		return nil, errors.New("ciphered text too short")
	}

	nonce := ciphertext[:ciph.NonceSize()]
	text := ciphertext[ciph.NonceSize():]

	clearText, err := ciph.Open(nil, nonce, text, nil)
	if err != nil {
		return nil, errors.New("failed to decrypt given message")
	}

	return clearText, nil
}
