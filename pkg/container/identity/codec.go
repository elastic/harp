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
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gosimple/slug"

	"github.com/elastic/harp/pkg/sdk/security/crypto/bech32"
	"github.com/elastic/harp/pkg/sdk/security/crypto/extra25519"
	"github.com/elastic/harp/pkg/sdk/types"
)

const (
	apiVersion = "harp.elastic.co/v1"
	kind       = "ContainerIdentity"
)

// -----------------------------------------------------------------------------

// New identity from description.
func New(random io.Reader, description string) (*Identity, []byte, error) {
	// Check arguments
	if err := validation.Validate(description, validation.Required, is.ASCII); err != nil {
		return nil, nil, fmt.Errorf("unable to create identity with invalid description: %w", err)
	}

	// Generate ed25519 keys as identity
	pub, priv, err := ed25519.GenerateKey(random)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate identity keypair: %w", err)
	}

	// Wrap as JWK
	jwk := JSONWebKey{
		Kty: "OKP",
		Crv: "Ed25519",
		X:   base64.RawURLEncoding.EncodeToString(pub[:]),
		D:   base64.RawURLEncoding.EncodeToString(priv[:]),
	}

	// Encode JWK as json
	payload, err := json.Marshal(jwk)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to serialize identity keypair: %w", err)
	}

	// Encode public key using Bech32 with description as hrp
	encoded, err := bech32.Encode(slug.Make(description), pub[:])
	if err != nil {
		return nil, nil, fmt.Errorf("unable to encode public key: %w", err)
	}

	// Return unsealed identity
	return &Identity{
		APIVersion:  apiVersion,
		Kind:        kind,
		Timestamp:   time.Now().UTC(),
		Description: description,
		Public:      encoded,
	}, payload, nil
}

// FromReader extract identity instance from reader.
func FromReader(r io.Reader) (*Identity, error) {
	// Check arguments
	if types.IsNil(r) {
		return nil, fmt.Errorf("unable to read nil reader")
	}

	// Convert input as a map
	var input Identity
	if err := json.NewDecoder(r).Decode(&input); err != nil {
		return nil, fmt.Errorf("unable to decode input JSON: %w", err)
	}

	// Check public key encoding
	_, _, err := bech32.Decode(input.Public)
	if err != nil {
		return nil, fmt.Errorf("invalid public key encoding")
	}

	// Check component
	if input.Private == nil {
		return nil, fmt.Errorf("invalid identity: missing private component")
	}

	// Return no error
	return &input, nil
}

// RecoveryKey returns the x25519 private encryption key from the private
// identity key.
func RecoveryKey(key *JSONWebKey) (*[32]byte, error) {
	// Check arguments
	if key == nil {
		return nil, errors.New("unable to get container key from a nil identity")
	}

	// Decode ed25519 private key
	privKeyRaw, err := base64.RawURLEncoding.DecodeString(key.D)
	if err != nil {
		return nil, errors.New("invalid identity, private key is invalid")
	}

	var recoveryPrivateKey [32]byte

	switch key.Crv {
	case "X25519": // Legacy keys
		copy(recoveryPrivateKey[:], privKeyRaw)
	case "Ed25519":
		// Convert Ed25519 private key to x25519 key.
		extra25519.PrivateKeyToCurve25519(&recoveryPrivateKey, privKeyRaw)
	default:
		return nil, fmt.Errorf("unhandled private key format '%s'", key.Crv)
	}

	// No error
	return &recoveryPrivateKey, nil
}
