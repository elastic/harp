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
package v2

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/awnumar/memguard"
	"golang.org/x/crypto/pbkdf2"
)

const (
	PublicKeyPrefix  = "v2.pk."
	PrivateKeyPrefix = "v2.sk."
)

// GenerateOptions represents container key generation options.
type generateOptions struct {
	dckdMasterKey *memguard.LockedBuffer
	dckdTarget    string
	randomSource  io.Reader
}

// GenerateOption represents functional pattern builder for optional parameters.
type GenerateOption func(o *generateOptions)

// WithDeterministicKey enables deterministic container key generation.
func WithDeterministicKey(masterKey *memguard.LockedBuffer, target string) GenerateOption {
	return func(o *generateOptions) {
		o.dckdMasterKey = masterKey
		o.dckdTarget = target
	}
}

// WithRandom provides the random source for key generation.
func WithRandom(random io.Reader) GenerateOption {
	return func(o *generateOptions) {
		o.randomSource = random
	}
}

// -----------------------------------------------------------------------------

// CenerateKey create an ECDSA P-384 key pair used as container identifier.
func GenerateKey(fopts ...GenerateOption) (publicKey, privateKey string, err error) {
	// Prepare defaults
	opts := &generateOptions{
		dckdMasterKey: nil,
		dckdTarget:    "",
		randomSource:  rand.Reader,
	}

	// Apply optional parameters
	for _, f := range fopts {
		f(opts)
	}

	// Master key derivation
	if opts.dckdMasterKey != nil {
		// PBKDF2-SHA512(masterKey, HMAC-SHA-512('harp deterministic salt v2', Target), 250000, 64)
		// Don't clean bytes, already done by memguard.
		masterKey := opts.dckdMasterKey.Bytes()
		if len(masterKey) < 32 {
			return "", "", fmt.Errorf("the master key must be 32 bytes long at least")
		}

		// Generate deterministic salt
		h := hmac.New(sha512.New, []byte("harp deterministic salt v2"))
		h.Write([]byte(opts.dckdTarget))
		salt := h.Sum(nil)
		defer memguard.WipeBytes(salt)

		// Derive deterministic container key using PBKDF2-SHA512
		dk := pbkdf2.Key(masterKey[:32], salt, 250000, 64, sha512.New)
		defer memguard.WipeBytes(dk)

		// Assign to seed
		opts.randomSource = bytes.NewBuffer(dk)
	}

	// Generate ECDSA P-384 container key pair
	priv, errGen := ecdsa.GenerateKey(elliptic.P384(), opts.randomSource)
	if errGen != nil {
		return "", "", fmt.Errorf("unable to generate container key: %w", errGen)
	}

	// Encode keys
	encodedPub := append([]byte(PublicKeyPrefix), base64.RawURLEncoding.EncodeToString(elliptic.MarshalCompressed(priv.Curve, priv.PublicKey.X, priv.PublicKey.Y))...)
	encodedPriv := append([]byte(PrivateKeyPrefix), base64.RawURLEncoding.EncodeToString(priv.D.Bytes())...)

	// No error
	return string(encodedPub), string(encodedPriv), nil
}

// PublicKeys return the appropriate key format used by the sealing strategy.
func (a *adapter) PublicKeys(keys ...string) ([]interface{}, error) {
	// v1.pk.[data]
	res := []interface{}{}

	for _, key := range keys {
		// Check prefix
		if !strings.HasPrefix(key, PublicKeyPrefix) {
			return nil, fmt.Errorf("iunexpected public key '%s' for v2 sealing process", key)
		}

		// Remove prefix if exists
		key = strings.TrimPrefix(key, PublicKeyPrefix)

		// Decode key
		keyRaw, err := base64.RawURLEncoding.DecodeString(key)
		if err != nil {
			return nil, fmt.Errorf("unable to decode public key '%s': %w", key, err)
		}

		// Public key sanity checks
		if len(keyRaw) != publicKeySize {
			return nil, fmt.Errorf("invalid public key length for key '%s'", key)
		}

		// Decode the compressed point
		x, y := elliptic.UnmarshalCompressed(elliptic.P384(), keyRaw)
		if x == nil {
			return nil, fmt.Errorf("invalid public key '%s'", key)
		}

		// Reassemble the public key
		pub := ecdsa.PublicKey{
			Curve: elliptic.P384(),
			X:     x,
			Y:     y,
		}

		// Append it to sealing keys
		res = append(res, &pub)
	}

	// No error
	return res, nil
}
