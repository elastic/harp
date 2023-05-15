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

package v4

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/elastic/harp/pkg/sdk/security"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/chacha20"
)

const (
	// KeyLength is the requested encryption key size.
	KeyLength               = 32
	nonceLength             = 32
	macLength               = 32
	encryptionKDFLength     = 56
	authenticationKeyLength = 32
	v4LocalPrefix           = "v4.local."
	v4PublicPrefix          = "v4.public."
)

// PASETO v4 symmetric encryption primitive.
// https://github.com/paseto-standard/paseto-spec/blob/master/docs/01-Protocol-Versions/Version4.md#encrypt
func Encrypt(r io.Reader, key, m []byte, f, i string) ([]byte, error) {
	// Create random seed
	var n [nonceLength]byte
	if _, err := io.ReadFull(r, n[:]); err != nil {
		return nil, fmt.Errorf("paseto: unable to generate random seed: %w", err)
	}

	// Delegate to primitive
	return encrypt(key, n[:], m, f, i)
}

// PASETO v4 symmetric decryption primitive
// https://github.com/paseto-standard/paseto-spec/blob/master/docs/01-Protocol-Versions/Version4.md#decrypt
func Decrypt(key, input []byte, f, i string) ([]byte, error) {
	// Check arguments
	if key == nil {
		return nil, errors.New("paseto: key is nil")
	}
	if len(key) != KeyLength {
		return nil, fmt.Errorf("paseto: invalid key length, it must be %d bytes long", KeyLength)
	}
	if input == nil {
		return nil, errors.New("paseto: input is nil")
	}

	// Check token header
	if !bytes.HasPrefix(input, []byte(v4LocalPrefix)) {
		return nil, errors.New("paseto: invalid token")
	}

	// Trim prefix
	input = input[len(v4LocalPrefix):]

	// Check footer usage
	if f != "" {
		// Split the footer and the body
		parts := bytes.SplitN(input, []byte("."), 2)
		if len(parts) != 2 {
			return nil, errors.New("paseto: invalid token, footer is missing but expected")
		}

		// Decode footer
		footer := make([]byte, base64.RawURLEncoding.DecodedLen(len(parts[1])))
		if _, err := base64.RawURLEncoding.Decode(footer, parts[1]); err != nil {
			return nil, fmt.Errorf("paseto: invalid token, footer has invalid encoding: %w", err)
		}

		// Compare footer
		if !security.SecureCompare([]byte(f), footer) {
			return nil, errors.New("paseto: invalid token, footer mismatch")
		}

		// Continue without footer
		input = parts[0]
	}

	// Decode token
	raw := make([]byte, base64.RawURLEncoding.DecodedLen(len(input)))
	if _, err := base64.RawURLEncoding.Decode(raw, input); err != nil {
		return nil, fmt.Errorf("paseto: invalid token body: %w", err)
	}

	// Extract components
	n := raw[:nonceLength]
	t := raw[len(raw)-macLength:]
	c := raw[macLength : len(raw)-macLength]

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

// PASETO v4 public signature primitive.
// https://github.com/paseto-standard/paseto-spec/blob/master/docs/01-Protocol-Versions/Version4.md#sign
func Sign(m []byte, sk ed25519.PrivateKey, f, i string) ([]byte, error) {
	// Compute protected content
	m2, err := pae([]byte(v4PublicPrefix), m, []byte(f), []byte(i))
	if err != nil {
		return nil, fmt.Errorf("unable to prepare protected content: %w", err)
	}

	// Sign protected content
	sig := ed25519.Sign(sk, m2)

	// Prepare content
	body := append([]byte{}, m...)
	body = append(body, sig...)

	// Encode body as RawURLBase64
	encodedBody := make([]byte, base64.RawURLEncoding.EncodedLen(len(body)))
	base64.RawURLEncoding.Encode(encodedBody, body)

	// Assemble final token
	final := append([]byte(v4PublicPrefix), encodedBody...)
	if f != "" {
		// Encode footer as RawURLBase64
		encodedFooter := make([]byte, base64.RawURLEncoding.EncodedLen(len(f)))
		base64.RawURLEncoding.Encode(encodedFooter, []byte(f))

		// Assemble body and footer
		final = append(final, append([]byte("."), encodedFooter...)...)
	}

	// No error
	return final, nil
}

// PASETO v4 signature verification primitive.
// https://github.com/paseto-standard/paseto-spec/blob/master/docs/01-Protocol-Versions/Version4.md#verify
func Verify(sm []byte, pk ed25519.PublicKey, f, i string) ([]byte, error) {
	// Check token header
	if !bytes.HasPrefix(sm, []byte(v4PublicPrefix)) {
		return nil, errors.New("paseto: invalid token")
	}

	// Trim prefix
	sm = sm[len(v4PublicPrefix):]

	// Check footer usage
	if f != "" {
		// Split the footer and the body
		parts := bytes.SplitN(sm, []byte("."), 2)
		if len(parts) != 2 {
			return nil, errors.New("paseto: invalid token, footer is missing but expected")
		}

		// Decode footer
		footer := make([]byte, base64.RawURLEncoding.DecodedLen(len(parts[1])))
		if _, err := base64.RawURLEncoding.Decode(footer, parts[1]); err != nil {
			return nil, fmt.Errorf("paseto: invalid token, footer has invalid encoding: %w", err)
		}

		// Compare footer
		if !security.SecureCompare([]byte(f), footer) {
			return nil, errors.New("paseto: invalid token, footer mismatch")
		}

		// Continue without footer
		sm = parts[0]
	}

	// Decode token
	raw := make([]byte, base64.RawURLEncoding.DecodedLen(len(sm)))
	if _, err := base64.RawURLEncoding.Decode(raw, sm); err != nil {
		return nil, fmt.Errorf("paseto: invalid token body: %w", err)
	}

	// Extract components
	m := raw[:len(raw)-ed25519.SignatureSize]
	s := raw[len(raw)-ed25519.SignatureSize:]

	// Compute protected content
	m2, err := pae([]byte(v4PublicPrefix), m, []byte(f), []byte(i))
	if err != nil {
		return nil, fmt.Errorf("unable to prepare protected content: %w", err)
	}

	// Check signature
	if !ed25519.Verify(pk, m2, s) {
		return nil, errors.New("paseto: invalid token signature")
	}

	// No error
	return m, nil
}

// -----------------------------------------------------------------------------

func encrypt(key, n, m []byte, f, i string) ([]byte, error) {
	// Check arguments
	if len(key) != KeyLength {
		return nil, fmt.Errorf("paseto: invalid key length, it must be %d bytes long", KeyLength)
	}
	if len(n) != nonceLength {
		return nil, fmt.Errorf("paseto: invalid nonce length, it must be %d bytes long", nonceLength)
	}

	// Derive keys from seed and secret key
	ek, n2, ak, err := kdf(key, n)
	if err != nil {
		return nil, fmt.Errorf("paseto: unable to derive keys from seed: %w", err)
	}

	// Prepare XChaCha20 stream cipher (nonce > 24bytes => XChacha)
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

	// Encode body as RawURLBase64
	encodedBody := make([]byte, base64.RawURLEncoding.EncodedLen(len(body)))
	base64.RawURLEncoding.Encode(encodedBody, body)

	// Assemble final token
	final := append([]byte(v4LocalPrefix), encodedBody...)
	if f != "" {
		// Encode footer as RawURLBase64
		encodedFooter := make([]byte, base64.RawURLEncoding.EncodedLen(len(f)))
		base64.RawURLEncoding.Encode(encodedFooter, []byte(f))

		// Assemble body and footer
		final = append(final, append([]byte("."), encodedFooter...)...)
	}

	// No error
	return final, nil
}

func kdf(key, n []byte) (ek, n2, ak []byte, err error) {
	// Derive encryption key
	encKDF, err := blake2b.New(encryptionKDFLength, key)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unable to initialize encryption kdf: %w", err)
	}

	// Domain separation (we use the same seed for 2 different purposes)
	encKDF.Write([]byte("paseto-encryption-key"))
	encKDF.Write(n)
	tmp := encKDF.Sum(nil)

	// Split encryption key (Ek) and nonce (n2)
	ek = tmp[:KeyLength]
	n2 = tmp[KeyLength:]

	// Derive authentication key
	authKDF, err := blake2b.New(authenticationKeyLength, key)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unable to initialize authentication kdf: %w", err)
	}

	// Domain separation (we use the same seed for 2 different purposes)
	authKDF.Write([]byte("paseto-auth-key-for-aead"))
	authKDF.Write(n)
	ak = authKDF.Sum(nil)

	// No error
	return ek, n2, ak, nil
}

func mac(ak []byte, h string, n, c []byte, f, i string) ([]byte, error) {
	// Compute pre-authentication message
	preAuth, err := pae([]byte(h), n, c, []byte(f), []byte(i))
	if err != nil {
		return nil, fmt.Errorf("unable to compute pre-authentication content: %w", err)
	}

	// Compute MAC
	mac, err := blake2b.New(macLength, ak)
	if err != nil {
		return nil, fmt.Errorf("unable to in initialize MAC kdf: %w", err)
	}

	// Hash pre-authentication content
	mac.Write(preAuth)

	// No error
	return mac.Sum(nil), nil
}

// https://github.com/paseto-standard/paseto-spec/blob/master/docs/01-Protocol-Versions/Common.md#authentication-padding
func pae(pieces ...[]byte) ([]byte, error) {
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
