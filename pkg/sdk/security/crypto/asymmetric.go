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

package crypto

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"fmt"

	"github.com/pkg/errors"
	"golang.org/x/crypto/nacl/box"

	"github.com/elastic/harp/build/fips"
)

// Keypair generates crypto keys according to given key type.
func Keypair(keyType string) (interface{}, error) {
	// Generate crypto materials
	pub, priv, err := generateKeyPair(keyType)
	if err != nil {
		return nil, fmt.Errorf("unable to generate a '%s' key pair: %w", keyType, err)
	}

	// No error
	return struct {
		Private interface{}
		Public  interface{}
	}{
		Private: priv,
		Public:  pub,
	}, nil
}

// -----------------------------------------------------------------------------

//nolint:gocyclo // To refactor
func generateKeyPair(keyType string) (publicKey, privateKey interface{}, err error) {
	switch keyType {
	case "rsa", "rsa:normal", "rsa:2048":
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to generate rsa-2048 key: %w", err)
		}
		pub := key.Public()
		return pub, key, nil
	case "rsa:strong", "rsa:4096":
		key, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to generate rsa-4096 key: %w", err)
		}
		pub := key.Public()
		return pub, key, nil
	case "ec", "ec:normal", "ec:p256":
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to generate ec-p256 key: %w", err)
		}
		pub := key.Public()
		return pub, key, nil
	case "ec:high", "ec:p384":
		key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to generate ec-p384 key: %w", err)
		}
		pub := key.Public()
		return pub, key, nil
	case "ec:strong", "ec:p521":
		key, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to generate ec-p521 key: %w", err)
		}
		pub := key.Public()
		return pub, key, nil
	case "ssh", "ed25519":
		if fips.Enabled() {
			return nil, nil, errors.New("ed25519 key processing is disabled in FIPS Mode")
		}
		pub, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to generate ed25519 key: %w", err)
		}
		return pub, priv, nil
	case "naclbox", "x25519":
		if fips.Enabled() {
			return nil, nil, errors.New("x25519 key processing is disabled in FIPS Mode")
		}
		pub, priv, err := box.GenerateKey(rand.Reader)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to generate naclbox key: %w", err)
		}
		return pub, priv, nil
	default:
		return nil, nil, fmt.Errorf("invalid keytype (%s) [(rsa, rsa:normal, rsa:2048), (rsa:strong, rsa:4096), (ec, ec:normal, ec:p256), (ec:high, ec:p384), (ec:strong, ec:p521), (ssh, ed25519), (naclbox, x25519)]", keyType)
	}
}
