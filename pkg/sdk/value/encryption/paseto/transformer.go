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

package paseto

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/elastic/harp/pkg/sdk/security"
	"github.com/elastic/harp/pkg/sdk/value"
	"github.com/elastic/harp/pkg/sdk/value/encryption"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/chacha20"
)

func init() {
	encryption.Register("paseto", Transformer)
}

const (
	keyLength     = 32
	v4LocalPrefix = "v4.local."
)

func Transformer(key string) (value.Transformer, error) {
	// Remove the prefix
	key = strings.TrimPrefix(key, "paseto:")

	// Decode key
	k, err := base64.URLEncoding.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("paseto: unable to decode key: %w", err)
	}
	if l := len(k); l != keyLength {
		return nil, fmt.Errorf("paseto: invalid secret key length (%d)", l)
	}

	// Copy secret key
	var secretKey [keyLength]byte
	copy(secretKey[:], k)

	return &pasetoTransformer{
		key: secretKey,
	}, nil
}

// -----------------------------------------------------------------------------

type pasetoTransformer struct {
	key [keyLength]byte
}

func (d *pasetoTransformer) From(_ context.Context, input []byte) ([]byte, error) {
	// Check token header
	if !strings.HasPrefix(string(input), v4LocalPrefix) {
		return nil, errors.New("paseto: invalid token")
	}

	var (
		// Token footer
		f = ""
		// Implicit assertions
		i = ""
	)

	// Decode token
	raw, err := base64.RawURLEncoding.DecodeString(string(input[9:]))
	if err != nil {
		return nil, fmt.Errorf("paseto: invalid token body: %w", err)
	}

	// Extract components
	n := raw[:32]
	t := raw[len(raw)-32:]
	c := raw[32 : len(raw)-32]

	// Derive keys from seed and secret key
	ek, n2, ak, err := kdf(d.key[:], n)
	if err != nil {
		return nil, fmt.Errorf("paseto: unable to derive keys from seed: %w", err)
	}

	// Compute MAC
	t2, err := mac(ak, v4LocalPrefix, n, c, f, i)
	if err != nil {
		return nil, fmt.Errorf("paseto: unable to compute MAC: %w", err)
	}

	// Time-constant compare MAC
	if !security.SecureCompare(t, t2) {
		return nil, errors.New("paseto: invalid pre-authentication header")
	}

	// Prepare XChaCha20 stream cipher
	ciph, err := chacha20.NewUnauthenticatedCipher(ek, n2)
	if err != nil {
		return nil, fmt.Errorf("paseto: unable to initialize XChaCha20 cipher: %w", err)
	}

	// Encrypt the payload
	m := make([]byte, len(input))
	ciph.XORKeyStream(m, c)

	// No error
	return m, nil
}

func (d *pasetoTransformer) To(_ context.Context, input []byte) ([]byte, error) {
	var (
		// Token footer
		f = ""
		// Implicit assertions
		i = ""
	)

	// Create random seed
	var n [32]byte
	if _, err := rand.Read(n[:]); err != nil {
		return nil, fmt.Errorf("paseto: unable to generate random seed: %w", err)
	}

	// Derive keys from seed and secret key
	ek, n2, ak, err := kdf(d.key[:], n[:])
	if err != nil {
		return nil, fmt.Errorf("paseto: unable to derive keys from seed: %w", err)
	}

	// Prepare XChaCha20 stream cipher
	ciph, err := chacha20.NewUnauthenticatedCipher(ek, n2)
	if err != nil {
		return nil, fmt.Errorf("paseto: unable to initialize XChaCha20 cipher: %w", err)
	}

	// Encrypt the payload
	c := make([]byte, len(input))
	ciph.XORKeyStream(c, input)

	// Compute MAC
	t, err := mac(ak, v4LocalPrefix, n[:], c, f, i)
	if err != nil {
		return nil, fmt.Errorf("paseto: unable to compute MAC: %w", err)
	}

	// Serialize final token
	// h || base64url(n || c || t)
	body := append(n[:], c...)
	body = append(body, t...)

	// No error
	return append([]byte("v4.local."), []byte(base64.RawURLEncoding.EncodeToString(body))...), nil
}

// -----------------------------------------------------------------------------

func kdf(key, n []byte) (ek, n2, ak []byte, err error) {
	// Derive encryption key
	encKDF, err := blake2b.New(56, key)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unable to initialize encryption kdf: %w", err)
	}

	encKDF.Write([]byte("paseto-encryption-key"))
	encKDF.Write(n)
	tmp := encKDF.Sum(nil)

	// Split encryption key (Ek) and nonce (n2)
	ek = tmp[:32]
	n2 = tmp[32:]

	// Derive authentication key
	authKDF, err := blake2b.New(32, key)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unable to initialize authentication kdf: %w", err)
	}

	authKDF.Write([]byte("paseto-auth-key-for-aead"))
	authKDF.Write(n)
	ak = authKDF.Sum(nil)

	// No error
	return ek, n2, ak, nil
}

func preAuthenticationEncoding(pieces ...[]byte) ([]byte, error) {
	output := &bytes.Buffer{}

	// Encode piece count
	count := len(pieces)
	if err := binary.Write(output, binary.LittleEndian, uint64(count)); err != nil {
		return nil, err
	}

	// For each element
	for i := range pieces {
		// Encode size
		if err := binary.Write(output, binary.LittleEndian, uint64(len(pieces[i]))); err != nil {
			return nil, err
		}

		// Encode data
		if _, err := output.Write(pieces[i]); err != nil {
			return nil, err
		}
	}

	// No error
	return output.Bytes(), nil
}

func mac(ak []byte, h string, n, c []byte, f, i string) ([]byte, error) {
	// Compute pre-authentication message
	preAuth, err := preAuthenticationEncoding([]byte(h), n, c, []byte(f), []byte(i))
	if err != nil {
		return nil, fmt.Errorf("unable to compute pre-authentication content: %w", err)
	}

	// Compute MAC
	mac, err := blake2b.New(32, ak)
	if err != nil {
		return nil, fmt.Errorf("unable to in initialize MAC kdf: %w", err)
	}

	// Hash pre-authentication content
	mac.Write(preAuth)

	// No error
	return mac.Sum(nil), nil
}
