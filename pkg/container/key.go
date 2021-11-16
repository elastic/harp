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
package container

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/awnumar/memguard"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/nacl/box"

	"github.com/elastic/harp/pkg/sdk/security/crypto/bech32"
	"github.com/elastic/harp/pkg/sdk/security/crypto/extra25519"
	"github.com/elastic/harp/pkg/sdk/types"
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

// CenerateKey create an X25519 key pair used as container identifier.
func GenerateKey(fopts ...GenerateOption) (publicKey, privateKey *[32]byte, err error) {
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
		// Argon2ID(masterKey, Blake2B-512(Target), 1, 64Mb, 4, 64)
		// Don't clean bytes, already done by memguard.
		masterKey := opts.dckdMasterKey.Bytes()
		if len(masterKey) < 32 {
			return nil, nil, fmt.Errorf("the master key must be 32 bytes long at least")
		}

		// Generate deterministic salt
		salt := blake2b.Sum512([]byte(opts.dckdTarget))
		defer memguard.WipeBytes(salt[:])

		// Derive deterministic container key using Argon2id
		dk := argon2.IDKey(masterKey[:32], salt[:], 1, 64*1024, 4, 64)
		defer memguard.WipeBytes(dk)

		// Assign to seed
		opts.randomSource = bytes.NewBuffer(dk)
	}

	// Generate x25519 container key pair
	pub, priv, errGen := box.GenerateKey(opts.randomSource)
	if errGen != nil {
		return nil, nil, fmt.Errorf("unable to generate container key: %w", errGen)
	}

	// No error
	return pub, priv, nil
}

// PublicKeysFromIdentities convert ed25519 public key to x25519 public container key.
func PublicKeysFromIdentities(identities ...string) ([]*[32]byte, error) {
	// If using sealing seed
	peerPublicKeys := []*[32]byte{}

	// Given identities
	if len(identities) == 0 {
		return nil, fmt.Errorf("at least one identity must be provided")
	}

	// Filter identities
	var filteredIdentities types.StringArray

	// Process identities
	for _, id := range identities {
		// Check if identity is already added
		if !filteredIdentities.AddIfNotContains(id) {
			continue
		}

		// Check encoding
		_, publicKeyRaw, errDecode := bech32.Decode(id)
		if errDecode != nil {
			return nil, fmt.Errorf("invalid '%s' as public identity: %w", id, errDecode)
		}

		// Convert ed25519 public to x25519 key
		var publicKey [32]byte
		if !extra25519.PublicKeyToCurve25519(&publicKey, publicKeyRaw) {
			return nil, fmt.Errorf("unable to convert identity '%s' to container key", id)
		}

		// Append to identity
		peerPublicKeys = append(peerPublicKeys, &publicKey)
	}

	// No error
	return peerPublicKeys, nil
}
