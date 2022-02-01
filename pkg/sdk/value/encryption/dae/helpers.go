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
	"bytes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"hash"
	"io"
	"strings"

	"golang.org/x/crypto/hkdf"
)

// decode the input key
// <key>(:<salt>)?
func decodeKey(key string) (k, salt []byte, err error) {
	// Check arguments
	if key == "" {
		return nil, nil, errors.New("input key must not be blank")
	}

	// Try to split input key
	parts := strings.Split(key, ":")
	if len(parts) > 2 {
		return nil, nil, errors.New("unable to decode input key")
	}

	// Decode key
	k, err = base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, nil, fmt.Errorf("unable to decode key: %w", err)
	}

	// Are we using salt?
	if len(parts) > 1 {
		// Decode salt
		salt, err = base64.URLEncoding.DecodeString(parts[1])
		if err != nil {
			return nil, nil, fmt.Errorf("unable to decode salt: %w", err)
		}
	}

	return k, salt, err
}

//nolint:unparam // to refactor
func deriveKey(secret, salt, info []byte, dkLen int) ([]byte, error) {
	// Prepare HKDF-SHA256
	reader := hkdf.New(sha256.New, secret, salt, info)

	// Prepare output buffer
	out := bytes.NewBuffer(nil)
	out.Grow(dkLen)
	limReader := &io.LimitedReader{
		R: reader,
		N: int64(dkLen),
	}

	// Read all data from buffer
	n, err := out.ReadFrom(limReader)
	if err != nil {
		return nil, fmt.Errorf("unable to derive key: %w", err)
	}
	if n != int64(dkLen) {
		return nil, errors.New("invalid derived key length")
	}

	// No error
	return out.Bytes(), nil
}

// -----------------------------------------------------------------------------

type NonceDeriverFunc func([]byte, cipher.AEAD) ([]byte, error)

func HMAC(h func() hash.Hash, key []byte) NonceDeriverFunc {
	return func(input []byte, ciph cipher.AEAD) ([]byte, error) {
		hm := hmac.New(h, key)
		hm.Write(input)
		nonceSum := hm.Sum(nil)
		nonce := nonceSum[:ciph.NonceSize()]
		return nonce, nil
	}
}

func Keyed(key []byte, khf func([]byte) (hash.Hash, error)) NonceDeriverFunc {
	return func(input []byte, ciph cipher.AEAD) ([]byte, error) {
		hm, err := khf(key)
		if err != nil {
			return nil, fmt.Errorf("dae: unable to initialize blake2b nonce deriver: %w", err)
		}
		hm.Write(input)
		nonceSum := hm.Sum(nil)
		nonce := nonceSum[:ciph.NonceSize()]
		return nonce, nil
	}
}
