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
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"

	"github.com/elastic/harp/pkg/sdk/security/crypto/extra25519"
)

// JSONWebKey holds internal container key attributes.
type JSONWebKey struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	X   string `json:"x,omitempty"`
	Y   string `json:"y,omitempty"`
	D   string `json:"d,omitempty"`
}

func (k *JSONWebKey) Sign(message []byte) (string, error) {
	var sig []byte

	// Decode private key
	d, err := base64.RawURLEncoding.DecodeString(k.D)
	if err != nil {
		return "", fmt.Errorf("unable to decode private key: %w", err)
	}

	switch k.Crv {
	case "Ed25519":
		if len(d) != ed25519.PrivateKeySize {
			return "", errors.New("invalid private key size")
		}

		// Sign the message
		sig = ed25519.Sign(ed25519.PrivateKey(d), message)
	case "P-384":
		if len(d) != 48 {
			return "", errors.New("invalid private key size")
		}

		// Rebuild the private key
		var sk ecdsa.PrivateKey
		sk.Curve = elliptic.P384()
		sk.D = new(big.Int).SetBytes(d)

		digest := sha512.Sum384(message)
		r, s, err := ecdsa.Sign(rand.Reader, &sk, digest[:])
		if err != nil {
			return "", fmt.Errorf("unable to sign the identity: %w", err)
		}

		// Assemble the signature
		sig = append(r.Bytes(), s.Bytes()...)
	}

	// Encode the signature
	return base64.RawURLEncoding.EncodeToString(sig), nil
}

// RecoveryKey returns the private encryption key from the private identity key.
func (k *JSONWebKey) RecoveryKey() (string, error) {
	// Decode private key
	privKeyRaw, err := base64.RawURLEncoding.DecodeString(k.D)
	if err != nil {
		return "", errors.New("invalid identity, private key is invalid")
	}

	switch k.Crv {
	case "X25519": // Legacy keys
		return base64.RawURLEncoding.EncodeToString(privKeyRaw), nil
	case "Ed25519":
		// Convert Ed25519 private key to x25519 key.
		var sk [32]byte
		extra25519.PrivateKeyToCurve25519(&sk, privKeyRaw)
		return fmt.Sprintf("v1.ck.%s", base64.RawURLEncoding.EncodeToString(sk[:])), nil
	case "P-384":
		// FIPS compliant sealing process use ECDSA P-384 key.
		return fmt.Sprintf("v2.ck.%s", base64.RawURLEncoding.EncodeToString(privKeyRaw)), nil
	default:
	}

	// Unhandled key
	return "", fmt.Errorf("unhandled private key format '%s'", k.Crv)
}
