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
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/chacha20"

	"github.com/elastic/harp/pkg/sdk/security"
)

func encrypt(key, n, m []byte, f, i string) ([]byte, error) {
	// Derive keys from seed and secret key
	ek, n2, ak, err := kdf(key, n)
	if err != nil {
		return nil, fmt.Errorf("paseto: unable to derive keys from seed: %w", err)
	}

	// Prepare XChaCha20 stream cipher
	ciph, err := chacha20.NewUnauthenticatedCipher(ek, n2)
	if err != nil {
		return nil, fmt.Errorf("paseto: unable to initialize XChaCha20 cipher: %w", err)
	}

	// Encrypt the payload
	c := make([]byte, len(m))
	ciph.XORKeyStream(c, m)

	// Compute MAC
	t, err := mac(ak, v4LocalPrefix, n, c, f, i)
	if err != nil {
		return nil, fmt.Errorf("paseto: unable to compute MAC: %w", err)
	}

	// Serialize final token
	// h || base64url(n || c || t)
	body := append([]byte{}, n...)
	body = append(body, c...)
	body = append(body, t...)

	// Assemble final token
	final := append([]byte("v4.local."), []byte(base64.RawURLEncoding.EncodeToString(body))...)
	if f != "" {
		final = append(final, append([]byte("."), []byte(base64.RawURLEncoding.EncodeToString([]byte(f)))...)...)
	}

	// No error
	return final, nil
}

func decrypt(key, input []byte, f, i string) ([]byte, error) {
	// Check token header
	if !strings.HasPrefix(string(input), v4LocalPrefix) {
		return nil, errors.New("paseto: invalid token")
	}

	// Trim prefix
	input = input[9:]

	// Check footer usage
	if f != "" {
		// Split the footer and the body
		parts := strings.SplitN(string(input), ".", 2)
		if len(parts) != 2 {
			return nil, errors.New("paseto: invalid token, footer is missing but expected")
		}

		// Decode footer
		footer, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			return nil, fmt.Errorf("paseto: invalid token, footer has invalid encoding: %w", err)
		}

		// Compare footer
		if !security.SecureCompareString(f, string(footer)) {
			return nil, errors.New("paseto: invalid token, footer mismatch")
		}

		// Continue without footer
		input = []byte(parts[0])
	}

	// Decode token
	raw, err := base64.RawURLEncoding.DecodeString(string(input))
	if err != nil {
		return nil, fmt.Errorf("paseto: invalid token body: %w", err)
	}

	// Extract components
	n := raw[:32]
	t := raw[len(raw)-32:]
	c := raw[32 : len(raw)-32]

	// Derive keys from seed and secret key
	ek, n2, ak, err := kdf(key, n)
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
	m := make([]byte, len(c))
	ciph.XORKeyStream(m, c)

	// No error
	return m, nil
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
