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
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/nacl/secretbox"
)

const (
	keyLength   = 32
	nonceLength = 24
)

func generateNonce() ([nonceLength]byte, error) {
	var nonce [nonceLength]byte
	_, err := io.ReadFull(rand.Reader, nonce[:])
	return nonce, err
}

func encrypt(plaintext []byte, key [keyLength]byte) ([]byte, error) {
	nonce, err := generateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce")
	}
	return secretbox.Seal(nonce[:], plaintext, &nonce, &key), nil
}

func decrypt(ciphertext []byte, key [keyLength]byte) ([]byte, error) {
	var nonce [nonceLength]byte
	copy(nonce[:], ciphertext[:nonceLength])
	decrypted, ok := secretbox.Open(nil, ciphertext[nonceLength:], &nonce, &key)
	if !ok {
		return nil, errors.New("failed to decrypt given message")
	}
	return decrypted, nil
}
