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
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/nacl/box"
)

func Legacy(random io.Reader) (*JSONWebKey, string, error) {
	// Generate X25519 keys as identity
	pub, priv, err := box.GenerateKey(random)
	if err != nil {
		return nil, "", fmt.Errorf("unable to generate identity keypair: %w", err)
	}

	// Wrap as JWK
	return &JSONWebKey{
		Kty: "OKP",
		Crv: "X25519",
		X:   base64.RawURLEncoding.EncodeToString(pub[:]),
		D:   base64.RawURLEncoding.EncodeToString(priv[:]),
	}, base64.RawURLEncoding.EncodeToString(pub[:]), err
}

func Ed25519(random io.Reader) (*JSONWebKey, string, error) {
	// Generate ed25519 keys as identity
	pub, priv, err := ed25519.GenerateKey(random)
	if err != nil {
		return nil, "", fmt.Errorf("unable to generate identity keypair: %w", err)
	}

	// Wrap as JWK
	return &JSONWebKey{
		Kty: "OKP",
		Crv: "Ed25519",
		X:   base64.RawURLEncoding.EncodeToString(pub[:]),
		D:   base64.RawURLEncoding.EncodeToString(priv[:]),
	}, fmt.Sprintf("v1.ipk.%s", base64.RawURLEncoding.EncodeToString(pub[:])), err
}

func P384(random io.Reader) (*JSONWebKey, string, error) {
	// Generate ecdsa P-384 keys as identity
	priv, err := ecdsa.GenerateKey(elliptic.P384(), random)
	if err != nil {
		return nil, "", fmt.Errorf("unable to generate identity keypair: %w", err)
	}

	// Marshall as compressed point
	pub := elliptic.MarshalCompressed(priv.Curve, priv.PublicKey.X, priv.PublicKey.Y)

	// P-384 key components are 48 bytes (384 bits)
	// big.Int.Bytes() strips leading zeros, so we need to pad to fixed size
	const p384KeySize = 48
	dBytes := priv.D.Bytes()
	dPadded := make([]byte, p384KeySize)
	copy(dPadded[p384KeySize-len(dBytes):], dBytes)

	xBytes := priv.PublicKey.X.Bytes()
	xPadded := make([]byte, p384KeySize)
	copy(xPadded[p384KeySize-len(xBytes):], xBytes)

	yBytes := priv.PublicKey.Y.Bytes()
	yPadded := make([]byte, p384KeySize)
	copy(yPadded[p384KeySize-len(yBytes):], yBytes)

	// Wrap as JWK
	return &JSONWebKey{
		Kty: "EC",
		Crv: "P-384",
		X:   base64.RawURLEncoding.EncodeToString(xPadded),
		Y:   base64.RawURLEncoding.EncodeToString(yPadded),
		D:   base64.RawURLEncoding.EncodeToString(dPadded),
	}, fmt.Sprintf("v2.ipk.%s", base64.RawURLEncoding.EncodeToString(pub)), err
}
