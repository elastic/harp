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

package identity

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/elastic/harp/pkg/sdk/types"
	"github.com/elastic/harp/pkg/sdk/value"
)

// Identity object to hold container sealer identity information.
type Identity struct {
	APIVersion  string      `json:"@apiVersion"`
	Kind        string      `json:"@kind"`
	Timestamp   time.Time   `json:"@timestamp"`
	Description string      `json:"@description"`
	Public      string      `json:"public"`
	Private     *PrivateKey `json:"private"`
	Signature   string      `json:"signature"`
}

// HasPrivateKey returns true if identity as a wrapped private.
func (i *Identity) HasPrivateKey() bool {
	return i.Private != nil
}

// Decrypt private key with given transformer.
func (i *Identity) Decrypt(ctx context.Context, t value.Transformer) (*JSONWebKey, error) {
	// Check arguments
	if types.IsNil(t) {
		return nil, fmt.Errorf("can't process with nil transformer")
	}
	if !i.HasPrivateKey() {
		return nil, fmt.Errorf("trying to decrypt a nil private key")
	}

	// Decode payload
	payload, err := base64.RawURLEncoding.DecodeString(i.Private.Content)
	if err != nil {
		return nil, fmt.Errorf("unable to decode private key: %w", err)
	}

	// Apply transformation
	clearText, err := t.From(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("unable to decrypt identity payload: %w", err)
	}

	// Decode key
	var key JSONWebKey
	if err = json.NewDecoder(bytes.NewReader(clearText)).Decode(&key); err != nil {
		return nil, fmt.Errorf("unable to decode payload as JSON: %w", err)
	}

	// Return result
	return &key, nil
}

// Verify the identity signature using its own public key.
func (i *Identity) Verify() error {
	// Clear the signature
	id := &Identity{}
	*id = *i

	// Clean protected
	id.Signature = ""
	id.Private = nil

	// Prepare protected
	protected, err := json.Marshal(id)
	if err != nil {
		return fmt.Errorf("unable to serialize identity for signature: %w", err)
	}

	// Decode the signature
	sig, err := base64.RawURLEncoding.DecodeString(i.Signature)
	if err != nil {
		return fmt.Errorf("unable to decode the signature: %w", err)
	}

	// Verify the signature
	switch {
	case strings.HasPrefix(i.Public, "v1.ipk."):
		// Decode public key
		pk, err := base64.RawURLEncoding.DecodeString(i.Public[7:])
		if err != nil {
			return fmt.Errorf("unable to decode public key: %w", err)
		}
		if len(pk) != ed25519.PublicKeySize {
			return errors.New("invalid public key size")
		}

		// Verify the signature
		if ed25519.Verify(ed25519.PublicKey(pk), protected, sig) {
			return nil
		}
	case strings.HasPrefix(i.Public, "v2.ipk"):
		// Decode public key
		pkRaw, err := base64.RawURLEncoding.DecodeString(i.Public[7:])
		if err != nil {
			return fmt.Errorf("unable to decode public key: %w", err)
		}
		x, y := elliptic.UnmarshalCompressed(elliptic.P384(), pkRaw)
		if x == nil || y == nil {
			return errors.New("unable to unmarshal the public key")
		}

		// Rebuild the public key
		var pk ecdsa.PublicKey
		pk.Curve = elliptic.P384()
		pk.X = x
		pk.Y = y

		r := new(big.Int).SetBytes(sig[:48])
		s := new(big.Int).SetBytes(sig[48:])

		digest := sha512.Sum384(protected)
		if ecdsa.Verify(&pk, digest[:], r, s) {
			return nil
		}
	}

	return errors.New("unable to validate identity signature")
}

// PrivateKey wraps encoded private and related informations.
type PrivateKey struct {
	Encoding string `json:"encoding,omitempty"`
	Content  string `json:"content"`
}

// -----------------------------------------------------------------------------

// JSONWebKey holds internal container key attributes.
type JSONWebKey struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	X   string `json:"x,omitempty"`
	Y   string `json:"y,omitempty"`
	D   string `json:"d,omitempty"`
}

func (jwk *JSONWebKey) Sign(message []byte) (string, error) {
	var sig []byte

	switch jwk.Crv {
	case "Ed25519":
		// Decode private key
		d, err := base64.RawURLEncoding.DecodeString(jwk.D)
		if err != nil {
			return "", fmt.Errorf("unable to decode private key: %w", err)
		}
		if len(d) != ed25519.PrivateKeySize {
			return "", errors.New("invalid private key size")
		}

		// Sign the message
		sig = ed25519.Sign(ed25519.PrivateKey(d), message)
	case "P-384":
		// Decode private key
		d, err := base64.RawURLEncoding.DecodeString(jwk.D)
		if err != nil {
			return "", fmt.Errorf("unable to decode private key: %w", err)
		}
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
