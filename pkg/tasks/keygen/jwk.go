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

package keygen

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"gopkg.in/square/go-jose.v2"

	"github.com/elastic/harp/pkg/tasks"
)

// BundleDumpTask implements secret-container creation from a Bundle Dump.
type JWKTask struct {
	SignatureAlgorithm string
	KeySize            int
	KeyID              string
	OutputWriter       tasks.WriterProvider
}

// Run the task.
func (t *JWKTask) Run(ctx context.Context) error {
	var (
		writer io.Writer
		err    error
	)

	// Generate key
	_, sk, err := t.keygenSig(jose.SignatureAlgorithm(t.SignatureAlgorithm), t.KeySize)
	if err != nil {
		return fmt.Errorf("unable to generate key pair: %w", err)
	}

	// Wrap as JWK
	priv := jose.JSONWebKey{
		Key:       sk,
		KeyID:     t.KeyID,
		Algorithm: t.SignatureAlgorithm,
	}

	// Create output writer
	writer, err = t.OutputWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to open output writer: %w", err)
	}

	// Encode as JSON
	if err := json.NewEncoder(writer).Encode(priv); err != nil {
		return fmt.Errorf("unable to encode JWK: %w", err)
	}

	// No error
	return nil
}

// -----------------------------------------------------------------------------

// KeygenSig generates keypair for corresponding SignatureAlgorithm.
//
//nolint:gocyclo // to refactor
func (t *JWKTask) keygenSig(alg jose.SignatureAlgorithm, bits int) (crypto.PublicKey, crypto.PrivateKey, error) {
	switch alg {
	case jose.ES256, jose.ES384, jose.ES512, jose.EdDSA:
		keylen := map[jose.SignatureAlgorithm]int{
			jose.ES256: 256,
			jose.ES384: 384,
			jose.ES512: 521, // sic!
			jose.EdDSA: 256,
		}
		if bits != 0 && bits != keylen[alg] {
			return nil, nil, errors.New("this `alg` does not support arbitrary key length")
		}
	case jose.RS256, jose.RS384, jose.RS512, jose.PS256, jose.PS384, jose.PS512:
		if bits == 0 {
			bits = 2048
		}
		if bits < 2048 {
			return nil, nil, errors.New("too short key for RSA `alg`, 2048+ is required")
		}
	case jose.HS256, jose.HS384, jose.HS512:
		return nil, nil, fmt.Errorf("can't generate crypto keys for '%s' signature", alg)
	}
	switch alg {
	case jose.ES256:
		// The cryptographic operations are implemented using constant-time algorithms.
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, nil, err
		}
		pub := key.Public()
		return pub, key, err
	case jose.ES384:
		// NB: The cryptographic operations do not use constant-time algorithms.
		key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
		if err != nil {
			return nil, nil, err
		}
		pub := key.Public()
		return pub, key, err
	case jose.ES512:
		// NB: The cryptographic operations do not use constant-time algorithms.
		key, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
		if err != nil {
			return nil, nil, err
		}
		pub := key.Public()
		return pub, key, err
	case jose.EdDSA:
		pub, key, err := ed25519.GenerateKey(rand.Reader)
		return pub, key, err
	case jose.RS256, jose.RS384, jose.RS512, jose.PS256, jose.PS384, jose.PS512:
		key, err := rsa.GenerateKey(rand.Reader, bits)
		if err != nil {
			return nil, nil, err
		}
		pub := key.Public()
		return pub, key, err
	case jose.HS256, jose.HS384, jose.HS512:
		return nil, nil, fmt.Errorf("can't generate crypto keys for '%s' signature", alg)
	default:
		return nil, nil, errors.New("unknown signature algorithm provided")
	}
}
