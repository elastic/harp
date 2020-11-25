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

package aes

import (
	"crypto/cipher"
	"crypto/rand"
	"fmt"
)

func encrypt(data []byte, block cipher.Block) ([]byte, error) {
	// Initialize AEAD cipher chain
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("aes: unable to initialize aead chain : %w", err)
	}

	// Check nonce size
	nonceSize := aead.NonceSize()

	// Calculate allocation limit (lgtm)
	bufSize := nonceSize + aead.Overhead() + len(data)
	if bufSize > 100*1024*1024 { // Limit to 100MB
		return nil, fmt.Errorf("data too large")
	}

	// Allocate data buffer
	result := make([]byte, bufSize)
	n, err := rand.Read(result[:nonceSize])
	if err != nil {
		return nil, fmt.Errorf("aes: unable to generate nonce : %w", err)
	}
	if n != nonceSize {
		return nil, fmt.Errorf("aes: unable to read sufficient random bytes")
	}

	// Encrypt and seal
	cipherText := aead.Seal(result[nonceSize:nonceSize], result[:nonceSize], data, nil)

	// No error
	return result[:nonceSize+len(cipherText)], nil
}

func decrypt(ciphertext []byte, block cipher.Block) ([]byte, error) {
	// Initialize AEAD cipher chain
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("aes: unable to initialize aead chain : %w", err)
	}

	// Check nonce size
	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("aes: the stored data was shorter than the required size")
	}

	// Try to decrypt data
	out, err := aead.Open(nil, ciphertext[:nonceSize], ciphertext[nonceSize:], nil)
	if err != nil {
		return nil, fmt.Errorf("aes: unable to decrypt data: %w", err)
	}

	// No error
	return out, nil
}
