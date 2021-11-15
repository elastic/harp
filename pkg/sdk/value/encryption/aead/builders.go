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
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"strings"

	miscreant "github.com/miscreant/miscreant.go"
	"golang.org/x/crypto/chacha20poly1305"

	"github.com/elastic/harp/pkg/sdk/value"
	"github.com/elastic/harp/pkg/sdk/value/encryption"
)

var (
	aesgcmPrefix     = "aes-gcm"
	aespmacsivPrefix = "aes-pmac-siv"
	aessivPrefix     = "aes-siv"
	chachaPrefix     = "chacha"
	xchachaPrefix    = "xchacha"
)

func init() {
	encryption.Register(aesgcmPrefix, AESGCM)
	encryption.Register(aespmacsivPrefix, AESPMACSIV)
	encryption.Register(aessivPrefix, AESSIV)
	encryption.Register(chachaPrefix, Chacha20Poly1305)
	encryption.Register(xchachaPrefix, XChacha20Poly1305)
}

// AESGCM returns an AES-GCM value transformer instance.
func AESGCM(key string) (value.Transformer, error) {
	// Remove the prefix
	key = strings.TrimPrefix(key, "aes-gcm:")

	// Decode key
	k, err := base64.URLEncoding.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("aes: unable to decode key: %w", err)
	}

	// Create AES block cipher
	block, err := aes.NewCipher(k)
	if err != nil {
		return nil, fmt.Errorf("aes: unable to initialize block cipher: %w", err)
	}

	// Initialize AEAD cipher chain
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("aes: unable to initialize aead chain : %w", err)
	}

	// Return transformer
	return &aeadTransformer{
		aead: aead,
	}, nil
}

// AESSIV returns an AES-SIV/AES-CMAC-SIV value transformer instance.
func AESSIV(key string) (value.Transformer, error) {
	// Remove the prefix
	key = strings.TrimPrefix(key, "aes-siv:")

	// Decode key
	k, err := base64.URLEncoding.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("aes: unable to decode key: %w", err)
	}
	if l := len(k); l != 64 {
		return nil, fmt.Errorf("aes: invalid secret key length (%d)", l)
	}

	// Initialize AEAD
	aead, err := miscreant.NewAEAD("AES-SIV", k, 16)
	if err != nil {
		return nil, fmt.Errorf("aes: unable to initialize aes-pmac-siv: %w", err)
	}

	// Return transformer
	return &aeadTransformer{
		aead: aead,
	}, nil
}

// AESPMACSIV returns an AES-PMAC-SIV value transformer instance.
func AESPMACSIV(key string) (value.Transformer, error) {
	// Remove the prefix
	key = strings.TrimPrefix(key, "aes-pmac-siv:")

	// Decode key
	k, err := base64.URLEncoding.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("aes: unable to decode key: %w", err)
	}
	if l := len(k); l != 64 {
		return nil, fmt.Errorf("aes: invalid secret key length (%d)", l)
	}

	// Initialize AEAD
	aead, err := miscreant.NewAEAD("AES-PMAC-SIV", k, 16)
	if err != nil {
		return nil, fmt.Errorf("aes: unable to initialize aes-pmac-siv: %w", err)
	}

	// Return transformer
	return &aeadTransformer{
		aead: aead,
	}, nil
}

// Chacha20Poly1305 returns an ChaCha20Poly1305 value transformer instance.
func Chacha20Poly1305(key string) (value.Transformer, error) {
	// Remove the prefix
	key = strings.TrimPrefix(key, "chacha:")

	// Decode key
	k, err := base64.URLEncoding.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("chacha: unable to decode key: %w", err)
	}
	if l := len(k); l != keyLength {
		return nil, fmt.Errorf("chacha: invalid secret key length (%d)", l)
	}

	// Create Chacha20-Poly1305 aead cipher
	aead, err := chacha20poly1305.New(k)
	if err != nil {
		return nil, fmt.Errorf("chacha: unable to initialize chacha cipher: %w", err)
	}

	// Return transformer
	return &aeadTransformer{
		aead: aead,
	}, nil
}

// XChacha20Poly1305 returns an XChaCha20Poly1305 value transformer instance.
func XChacha20Poly1305(key string) (value.Transformer, error) {
	// Remove the prefix
	key = strings.TrimPrefix(key, "xchacha:")

	// Decode key
	k, err := base64.URLEncoding.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("xchacha: unable to decode key: %w", err)
	}
	if l := len(k); l != keyLength {
		return nil, fmt.Errorf("xchacha: invalid secret key length (%d)", l)
	}

	// Create Chacha20-Poly1305 aead cipher
	aead, err := chacha20poly1305.NewX(k)
	if err != nil {
		return nil, fmt.Errorf("xchacha: unable to initialize chacha cipher: %w", err)
	}

	// Return transformer
	return &aeadTransformer{
		aead: aead,
	}, nil
}
