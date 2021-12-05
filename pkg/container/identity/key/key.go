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

package key

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/elastic/harp/pkg/sdk/security/crypto/extra25519"
)

const (
	V1IdentityPublicKeyPrefix = "v1.ipk."
	V2IdentityPublicKeyPrefix = "v2.ipk."
)

// -----------------------------------------------------------------------------

type Key struct {
	version  uint32
	key      interface{}
	identity bool
	public   bool
}

func FromString(input string) (*Key, error) {
	switch {
	// Ed25519 public key
	case strings.HasPrefix(input, V1IdentityPublicKeyPrefix):
		// Decode public key
		pk, err := base64.RawURLEncoding.DecodeString(input[7:])
		if err != nil {
			return nil, fmt.Errorf("unable to decode public key: %w", err)
		}
		if len(pk) != ed25519.PublicKeySize {
			return nil, errors.New("invalid public key size")
		}

		// Return wrapped key
		return &Key{
			version:  1,
			key:      ed25519.PublicKey(pk),
			identity: true,
			public:   true,
		}, nil

	// EC P-384 public key
	case strings.HasPrefix(input, V2IdentityPublicKeyPrefix):
		// Decode public key
		pkRaw, err := base64.RawURLEncoding.DecodeString(input[7:])
		if err != nil {
			return nil, fmt.Errorf("unable to decode public key: %w", err)
		}
		x, y := elliptic.UnmarshalCompressed(elliptic.P384(), pkRaw)
		if x == nil || y == nil {
			return nil, errors.New("unable to unmarshal the public key")
		}

		// Rebuild the public key
		var pk ecdsa.PublicKey
		pk.Curve = elliptic.P384()
		pk.X = x
		pk.Y = y

		// Return wrapped key
		return &Key{
			version:  2,
			key:      &pk,
			identity: true,
			public:   true,
		}, nil

	// Unrecognized
	default:
	}

	// Default to error
	return nil, fmt.Errorf("unrecognized key '%s'", input)
}

// -----------------------------------------------------------------------------

func (k *Key) Verify(message, signature []byte) bool {
	switch keyRaw := k.key.(type) {
	case *ecdsa.PublicKey:
		// Unpack signature
		r := new(big.Int).SetBytes(signature[:48])
		s := new(big.Int).SetBytes(signature[48:])

		// Compute digest
		digest := sha512.Sum384(message)

		// Verify signature
		return ecdsa.Verify(keyRaw, digest[:], r, s)
	case ed25519.PublicKey:
		// Verify the signature
		return ed25519.Verify(keyRaw, message, signature)
	default:
	}

	return false
}

func (k *Key) String() string {
	var payload []byte
	switch keyRaw := k.key.(type) {
	case *ecdsa.PublicKey:
		payload = elliptic.MarshalCompressed(keyRaw.Curve, keyRaw.X, keyRaw.Y)
	case ed25519.PublicKey:
		payload = keyRaw
	default:
		return ""
	}

	return fmt.Sprintf("v%d.ipk.%s", k.version, base64.RawURLEncoding.EncodeToString(payload))
}

func (k *Key) SealingKey() string {
	var payload []byte
	switch keyRaw := k.key.(type) {
	case *ecdsa.PublicKey:
		payload = elliptic.MarshalCompressed(keyRaw.Curve, keyRaw.X, keyRaw.Y)
	case ed25519.PublicKey:
		// Convert Ed25519 to X25519
		var pkRaw [32]byte
		if !extra25519.PublicKeyToCurve25519(&pkRaw, keyRaw) {
			return ""
		}
		payload = pkRaw[:]
	default:
		return ""
	}

	return fmt.Sprintf("v%d.sk.%s", k.version, base64.RawURLEncoding.EncodeToString(payload))
}
