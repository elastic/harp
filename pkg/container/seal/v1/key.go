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
package v1

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/awnumar/memguard"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/nacl/box"

	"github.com/elastic/harp/pkg/container/seal"
	"github.com/elastic/harp/pkg/sdk/security/crypto/extra25519"
)

const (
	PublicKeyPrefix  = "v1.pk."
	PrivateKeyPrefix = "v1.sk."
)

// -----------------------------------------------------------------------------

// CenerateKey create an X25519 key pair used as container identifier.
func (a *adapter) GenerateKey(fopts ...seal.GenerateOption) (publicKey, privateKey string, err error) {
	// Prepare defaults
	opts := &seal.GenerateOptions{
		DCKDMasterKey: nil,
		DCKDTarget:    "",
		RandomSource:  rand.Reader,
	}

	// Apply optional parameters
	for _, f := range fopts {
		f(opts)
	}

	// Master key derivation
	if opts.DCKDMasterKey != nil {
		// Argon2ID(masterKey, Blake2B-512('harp deterministic salt v1', Target), 1, 64Mb, 4, 64)
		// Don't clean bytes, already done by memguard.
		masterKey := opts.DCKDMasterKey.Bytes()
		if len(masterKey) < 32 {
			return "", "", fmt.Errorf("the master key must be 32 bytes long at least")
		}

		// Generate deterministic salt
		h, err := blake2b.New512([]byte("harp deterministic salt v1"))
		if err != nil {
			return "", "", fmt.Errorf("unable to initialize salt derivation: %w", err)
		}
		h.Write([]byte(opts.DCKDTarget))
		salt := h.Sum(nil)
		defer memguard.WipeBytes(salt)

		// Derive deterministic container key using Argon2id
		dk := argon2.IDKey(masterKey[:32], salt, 1, 64*1024, 4, 64)
		defer memguard.WipeBytes(dk)

		// Assign to seed
		opts.RandomSource = bytes.NewBuffer(dk)
	}

	// Generate x25519 container key pair
	pub, priv, errGen := box.GenerateKey(opts.RandomSource)
	if errGen != nil {
		return "", "", fmt.Errorf("unable to generate container key: %w", errGen)
	}

	// Encode keys
	encodedPub := append([]byte(PublicKeyPrefix), base64.RawURLEncoding.EncodeToString(pub[:])...)
	encodedPriv := append([]byte(PrivateKeyPrefix), base64.RawURLEncoding.EncodeToString(priv[:])...)

	// No error
	return string(encodedPub), string(encodedPriv), nil
}

// PublicKeys return the appropriate key format used by the sealing strategy.
func (a *adapter) publicKeys(keys ...string) ([]*[32]byte, error) {
	// v1.pk.[data]
	res := []*[publicKeySize]byte{}

	for _, key := range keys {
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
		if extra25519.IsEdLowOrder(keyRaw) {
			return nil, fmt.Errorf("low order public key usage is forbidden for key '%s, try to generate a new one to fix the issue", key)
		}

		// Copy the public key
		var pk [publicKeySize]byte
		copy(pk[:], keyRaw[:publicKeySize])

		// Append it to sealing keys
		res = append(res, &pk)
	}

	// No error
	return res, nil
}
