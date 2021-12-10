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
			digest = input
			if len(digest) != h.Size() {
				return nil, fmt.Errorf("raw: invalid pre-hash length, expected %d bytes, got %d", h.Size(), len(input))
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

		// Calculate optimized buffer size
		curveBits := sk.Curve.Params().BitSize
		keyBytes := curveBits / 8
		if curveBits%8 > 0 {
			keyBytes++
		}

		// We serialize the outputs (r and s) into big-endian byte arrays and pad
		// them with zeros on the left to make sure the sizes work out. Both arrays
		// must be keyBytes long, and the output must be 2*keyBytes long.
		rBytes := r.Bytes()
		rBytesPadded := make([]byte, keyBytes)
		copy(rBytesPadded[keyBytes-len(rBytes):], rBytes)

		sBytes := s.Bytes()
		sBytesPadded := make([]byte, keyBytes)
		copy(sBytesPadded[keyBytes-len(sBytes):], sBytes)

		// Assemble the signature
		sig = rBytesPadded
		sig = append(sig, sBytesPadded...)
	default:
		return nil, errors.New("raw: key is not supported")
	}
	if err != nil {
		return nil, fmt.Errorf("raw: unable to sign input: %w", err)
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
		pkLen, err := d.privateKeySizeFromCurve(sk.Curve)
		if err != nil {
			return nil, fmt.Errorf("raw: unable to retrieve private key length: %w", err)
		}

		// Check minimal input length
		if len(input) < 2*pkLen {
			return nil, errors.New("raw: too short signature")
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
			digest = input[2*pkLen:]
			if len(digest) != h.Size() {
				return nil, fmt.Errorf("invalid pre-hash length, expected %d bytes, got %d bytes", h.Size(), len(digest))
			}
		} else {
			// Hash the decoded content
			h.Write(input[2*pkLen:])

			// Set hash value
			digest = h.Sum(nil)
		}

		// Extract sig
		sig := input[:2*pkLen]

		// Unpack signature
		var (
			r = new(big.Int).SetBytes(sig[:pkLen])
			s = new(big.Int).SetBytes(sig[pkLen:])
		)

		// Validate signature
		if ecdsa.Verify(sk, digest, r, s) {
			return input[2*pkLen:], nil
		}
	default:
		return nil, errors.New("raw: key is not supported")
	}

	// Default to error
	return nil, errors.New("raw: unable to validate input signature")
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

func (d *rawTransformer) privateKeySizeFromCurve(curve elliptic.Curve) (int, error) {
	switch curve {
	case elliptic.P256():
		return 32, nil
	case elliptic.P384():
		return 48, nil
	case elliptic.P521():
		return 66, nil
	default:
	}

	return 0, errors.New("current curve is not supported")
}
