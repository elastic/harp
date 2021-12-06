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

package raw

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"hash"
	"math/big"

	"github.com/elastic/harp/pkg/sdk/security/crypto/rfc6979"
	"github.com/elastic/harp/pkg/sdk/types"
	"github.com/elastic/harp/pkg/sdk/value/signature"
)

type rawTransformer struct {
	key interface{}
}

// -----------------------------------------------------------------------------

func (d *rawTransformer) To(ctx context.Context, input []byte) ([]byte, error) {
	if types.IsNil(d.key) {
		return nil, fmt.Errorf("paseto: signer key must not be nil")
	}

	var (
		sig []byte
		err error
	)

	switch sk := d.key.(type) {
	case ed25519.PrivateKey:
		// Ed25519 doesn't support pre-hash as input to prevent collision.
		sig = ed25519.Sign(sk, input)

	case *ecdsa.PrivateKey:
		var digest []byte

		// Get hash function for curve
		hf, errHash := d.hashFromCurve(sk.Curve)
		if errHash != nil {
			return nil, fmt.Errorf("raw: unable to retrieve hash function for curve: %w", errHash)
		}

		// Build a hash function instance
		h := hf()

		// Input is already a hash?
		if signature.IsInputPreHashed(ctx) {
			if len(input) == h.Size() {
				digest = input
			} else {
				return nil, fmt.Errorf("invalid pre-hash length, expected %d bytes", h.Size())
			}
		} else {
			// Hash the decoded content
			h.Write(input)

			// Set hash value
			digest = h.Sum(nil)
		}

		var (
			errSig error
			r, s   *big.Int
		)
		if signature.IsDeterministic(ctx) {
			// Deterministic signature
			r, s = rfc6979.SignECDSA(sk, digest, hf)
			if r == nil {
				errSig = errors.New("unable to apply determistic signature")
			}
		} else {
			// Sign
			r, s, errSig = ecdsa.Sign(rand.Reader, sk, digest)
		}
		if errSig != nil {
			return nil, fmt.Errorf("raw: unable to sign the content: %w", errSig)
		}

		// Pack the signature
		sig = append(r.Bytes(), s.Bytes()...)
	default:
		return nil, errors.New("raw: key is not supported")
	}
	if err != nil {
		return nil, fmt.Errorf("raw: unable so sign input: %w", err)
	}

	// Detached signature requested?
	if signature.IsDetached(ctx) {
		return sig, nil
	}

	// No error
	return append(sig, input...), nil
}

func (d *rawTransformer) From(ctx context.Context, input []byte) ([]byte, error) {
	switch sk := d.key.(type) {
	case ed25519.PublicKey:
		// Ed25519 doesn't support pre-hash as input to prevent collision.
		if ed25519.Verify(sk, input[ed25519.SignatureSize:], input[:ed25519.SignatureSize]) {
			return input[ed25519.SignatureSize:], nil
		}

	case *ecdsa.PublicKey:
		// Extract signature
		sigLen, err := d.signatureSizeFromCurve(sk)
		if err != nil {
			return nil, fmt.Errorf("raw: unable to retrieve signature length: %w", err)
		}

		// Get hash function for curve
		hf, err := d.hashFromCurve(sk.Curve)
		if err != nil {
			return nil, fmt.Errorf("raw: unable to retrieve hash function for curve: %w", err)
		}

		// Build a hash function instance
		h := hf()

		var digest []byte

		// Input is already a hash?
		if signature.IsInputPreHashed(ctx) {
			if len(input) == h.Size() {
				digest = input
			} else {
				return nil, fmt.Errorf("invalid pre-hash length, expected %d bytes", h.Size())
			}
		} else {
			// Hash the decoded content
			h.Write(input)

			// Set hash value
			digest = h.Sum(nil)
		}

		// Unpack signature
		var (
			r = new(big.Int).SetBytes(input[sigLen/2:])
			s = new(big.Int).SetBytes(input[:sigLen/2])
		)

		// Validate signature
		if ecdsa.Verify(sk, digest, r, s) {
			return input[sigLen:], nil
		}
	default:
		return nil, errors.New("raw: key is not supported")
	}

	// Default to error
	return nil, errors.New("raw: unable to valid input signature")
}

// -----------------------------------------------------------------------------

func (d *rawTransformer) hashFromCurve(curve elliptic.Curve) (func() hash.Hash, error) {
	switch curve {
	case elliptic.P256():
		return sha256.New, nil
	case elliptic.P384():
		return sha512.New384, nil
	case elliptic.P521():
		return sha512.New, nil
	default:
	}

	// Default to error
	return nil, errors.New("current curve is not supported")
}

func (d *rawTransformer) signatureSizeFromCurve(curve elliptic.Curve) (int, error) {
	switch curve {
	case elliptic.P256():
		return 64, nil
	case elliptic.P384():
		return 96, nil
	case elliptic.P521():
		return 132, nil
	default:
	}

	return 0, errors.New("current curve is not supported")
}
