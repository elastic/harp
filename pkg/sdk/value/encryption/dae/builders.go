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
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"fmt"
	"strings"

	miscreant "github.com/miscreant/miscreant.go"

	"github.com/elastic/harp/build/fips"
	"github.com/elastic/harp/pkg/sdk/value"
	"github.com/elastic/harp/pkg/sdk/value/encryption"

	"golang.org/x/crypto/chacha20poly1305"
)

var (
	aesgcmPrefix     = "dae-aes-gcm"
	aespmacsivPrefix = "dae-aes-pmac-siv"
	aessivPrefix     = "dae-aes-siv"
	chachaPrefix     = "dae-chacha"
	xchachaPrefix    = "dae-xchacha"
)

func init() {
	encryption.Register(aesgcmPrefix, AESGCM)

	if !fips.Enabled() {
		encryption.Register(aespmacsivPrefix, AESPMACSIV)
		encryption.Register(aessivPrefix, AESSIV)
		encryption.Register(chachaPrefix, Chacha20Poly1305)
		encryption.Register(xchachaPrefix, XChacha20Poly1305)
	}
}

// AESGCM returns an AES-GCM value transformer instance.
func AESGCM(key string) (value.Transformer, error) {
	// Remove the prefix
	key = strings.TrimPrefix(key, "dae-aes-gcm:")

	k, salt, err := decodeKey(key)
	if err != nil {
		return nil, fmt.Errorf("aes: unable to decode transformer key: %w", err)
	}
	switch len(k) {
	case 16, 24, 32:
	default:
		return nil, fmt.Errorf("aes: invalid key length, use 16 bytes (AES128) or 24 bytes (AES192) or 32 bytes (AES256)")
	}

	// Derive keys from input key
	dk, err := deriveKey(k, salt, nil, len(k)+32)
	if err != nil {
		return nil, fmt.Errorf("aes: unable te derive required keys: %w", err)
	}

	// Create AES block cipher
	block, err := aes.NewCipher(dk[:len(k)])
	if err != nil {
		return nil, fmt.Errorf("aes: unable to initialize block cipher: %w", err)
	}

	// Initialize AEAD cipher chain
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("aes: unable to initialize aead chain : %w", err)
	}

	// Return transformer
	return &daeTransformer{
		aead:             aead,
		nonceDeriverFunc: HMAC(sha256.New, dk[len(k):]),
	}, nil
}

// AESSIV returns an AES-SIV/AES-CMAC-SIV value transformer instance.
func AESSIV(key string) (value.Transformer, error) {
	// Remove the prefix
	key = strings.TrimPrefix(key, "dae-aes-siv:")

	// Decode key
	k, salt, err := decodeKey(key)
	if err != nil {
		return nil, fmt.Errorf("aes: unable to decode transformer key: %w", err)
	}
	if l := len(k); l != 64 {
		return nil, fmt.Errorf("aes: invalid secret key length (%d)", l)
	}

	// Derive keys from input key
	dk, err := deriveKey(k, salt, nil, len(k)+32)
	if err != nil {
		return nil, fmt.Errorf("aes: unable te derive required keys: %w", err)
	}

	// Initialize AEAD
	aead, err := miscreant.NewAEAD("AES-SIV", dk[:len(k)], 32)
	if err != nil {
		return nil, fmt.Errorf("aes: unable to initialize aes-pmac-siv: %w", err)
	}

	// Return transformer
	return &daeTransformer{
		aead:             aead,
		nonceDeriverFunc: HMAC(sha256.New, dk[len(k):]),
	}, nil
}

// AESPMACSIV returns an AES-PMAC-SIV value transformer instance.
func AESPMACSIV(key string) (value.Transformer, error) {
	// Remove the prefix
	key = strings.TrimPrefix(key, "dae-aes-pmac-siv:")

	// Decode key
	k, salt, err := decodeKey(key)
	if err != nil {
		return nil, fmt.Errorf("aes: unable to decode transformer key: %w", err)
	}
	if l := len(k); l != 64 {
		return nil, fmt.Errorf("aes: invalid secret key length (%d)", l)
	}

	// Derive keys from input key
	dk, err := deriveKey(k, salt, nil, len(k)+32)
	if err != nil {
		return nil, fmt.Errorf("aes: unable te derive required keys: %w", err)
	}

	// Initialize AEAD
	aead, err := miscreant.NewAEAD("AES-PMAC-SIV", dk[:len(k)], 32)
	if err != nil {
		return nil, fmt.Errorf("aes: unable to initialize aes-pmac-siv: %w", err)
	}

	// Return transformer
	return &daeTransformer{
		aead:             aead,
		nonceDeriverFunc: HMAC(sha256.New, dk[len(k):]),
	}, nil
}

// Chacha20Poly1305 returns an ChaCha20Poly1305 value transformer instance.
func Chacha20Poly1305(key string) (value.Transformer, error) {
	// Remove the prefix
	key = strings.TrimPrefix(key, "dae-chacha:")

	// Decode key
	k, salt, err := decodeKey(key)
	if err != nil {
		return nil, fmt.Errorf("chacha: unable to decode transformer key: %w", err)
	}
	if l := len(k); l != 32 {
		return nil, fmt.Errorf("chacha: invalid secret key length (%d)", l)
	}

	// Derive keys from input key
	dk, err := deriveKey(k, salt, nil, len(k)+32)
	if err != nil {
		return nil, fmt.Errorf("chacha: unable te derive required keys: %w", err)
	}

	// Create Chacha20-Poly1305 aead cipher
	aead, err := chacha20poly1305.New(dk[:len(k)])
	if err != nil {
		return nil, fmt.Errorf("chacha: unable to initialize chacha cipher: %w", err)
	}

	// Return transformer
	return &daeTransformer{
		aead:             aead,
		nonceDeriverFunc: HMAC(sha256.New, dk[len(k):]),
	}, nil
}

// XChacha20Poly1305 returns an XChaCha20Poly1305 value transformer instance.
func XChacha20Poly1305(key string) (value.Transformer, error) {
	// Remove the prefix
	key = strings.TrimPrefix(key, "dae-xchacha:")

	// Decode key
	k, salt, err := decodeKey(key)
	if err != nil {
		return nil, fmt.Errorf("xchacha: unable to decode transformer key: %w", err)
	}
	if l := len(k); l != 32 {
		return nil, fmt.Errorf("xchacha: invalid secret key length (%d)", l)
	}

	// Derive keys from input key
	dk, err := deriveKey(k, salt, nil, len(k)+32)
	if err != nil {
		return nil, fmt.Errorf("xchacha: unable te derive required keys: %w", err)
	}

	// Create Chacha20-Poly1305 aead cipher
	aead, err := chacha20poly1305.NewX(dk[:len(k)])
	if err != nil {
		return nil, fmt.Errorf("xchacha: unable to initialize chacha cipher: %w", err)
	}

	// Return transformer
	return &daeTransformer{
		aead:             aead,
		nonceDeriverFunc: HMAC(sha256.New, dk[len(k):]),
	}, nil
}
