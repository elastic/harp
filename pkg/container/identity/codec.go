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
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"

	"github.com/elastic/harp/pkg/sdk/security/crypto/extra25519"
	"github.com/elastic/harp/pkg/sdk/types"
)

const (
	apiVersion = "harp.elastic.co/v1"
	kind       = "ContainerIdentity"
)

// -----------------------------------------------------------------------------

type PrivateKeyGeneratorFunc func(io.Reader) (*JSONWebKey, string, error)

// New identity from description.
func New(random io.Reader, description string, generator PrivateKeyGeneratorFunc) (*Identity, []byte, error) {
	// Check arguments
	if err := validation.Validate(description, validation.Required, is.ASCII); err != nil {
		return nil, nil, fmt.Errorf("unable to create identity with invalid description: %w", err)
	}

	// Delegate to generator
	jwk, encodedPub, err := generator(random)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate identity private key: %w", err)
	}

	// Encode JWK as json
	payload, err := json.Marshal(jwk)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to serialize identity keypair: %w", err)
	}

	// Prepae identity object
	id := &Identity{
		APIVersion:  apiVersion,
		Kind:        kind,
		Timestamp:   time.Now().UTC(),
		Description: description,
		Public:      encodedPub,
	}

	// Encode to json for signature
	protected, err := json.Marshal(id)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to serialize identity for signature: %w", err)
	}

	// Sign the protected data
	sig, err := jwk.Sign(protected)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to sign protected data: %w", err)
	}

	// Auto-assign the signature
	id.Signature = sig

	// Return unsealed identity
	return id, payload, nil
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

	// Check component
	if input.Private == nil {
		return nil, fmt.Errorf("invalid identity: missing private component")
	}

	// Validate self signature
	if errVerify := input.Verify(); errVerify != nil {
		return nil, fmt.Errorf("unable to verify identity: %w", errVerify)
	}

	// Return no error
	return &input, nil
}

// RecoveryKey returns the private encryption key from the private identity key.
func RecoveryKey(key *JSONWebKey) (string, error) {
	// Check arguments
	if key == nil {
		return "", errors.New("unable to get container key from a nil identity")
	}

	// Decode private key
	privKeyRaw, err := base64.RawURLEncoding.DecodeString(key.D)
	if err != nil {
		return "", errors.New("invalid identity, private key is invalid")
	}

	switch key.Crv {
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
	return "", fmt.Errorf("unhandled private key format '%s'", key.Crv)
}

func SealingPublicKey(ipk string) (string, error) {
	// Check arguments
	if ipk == "" {
		return "", errors.New("unable to get container sealing key from a blank public key")
	}

	switch {
	case strings.HasPrefix(ipk, "v1.ipk."):
		// Decode public key
		pk, err := base64.RawURLEncoding.DecodeString(ipk[7:])
		if err != nil {
			return "", fmt.Errorf("unable to decode public key: %w", err)
		}
		if len(pk) != ed25519.PublicKeySize {
			return "", errors.New("invalid public key size")
		}

		// Convert Ed25519 to X25519
		var pkRaw [32]byte
		if !extra25519.PublicKeyToCurve25519(&pkRaw, pk) {
			return "", errors.New("unable to convert identity public key to container sealing key")
		}

		return fmt.Sprintf("v1.sk.%s", base64.RawURLEncoding.EncodeToString(pkRaw[:])), nil
	case strings.HasPrefix(ipk, "v2.ipk."):
		// Decode public key
		pkRaw, err := base64.RawURLEncoding.DecodeString(ipk[7:])
		if err != nil {
			return "", fmt.Errorf("unable to decode public key: %w", err)
		}
		x, y := elliptic.UnmarshalCompressed(elliptic.P384(), pkRaw)
		if x == nil || y == nil {
			return "", errors.New("unable to unmarshal the public key")
		}

		return fmt.Sprintf("v2.sk.%s", base64.RawURLEncoding.EncodeToString(pkRaw)), nil
	default:
	}

	return "", errors.New("identiy public key is not supported")
}
