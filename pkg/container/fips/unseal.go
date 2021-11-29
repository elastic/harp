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

package fips

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/sha512"
	"errors"
	"fmt"
	"math/big"

	"github.com/awnumar/memguard"
	"github.com/davecgh/go-spew/spew"
	"google.golang.org/protobuf/proto"

	containerv1 "github.com/elastic/harp/api/gen/go/harp/container/v1"
	"github.com/elastic/harp/pkg/sdk/security"
	"github.com/elastic/harp/pkg/sdk/types"
)

// Unseal a sealed container with the given identity
//nolint:funlen,gocyclo // To refactor
func Unseal(container *containerv1.Container, identity *memguard.LockedBuffer) (*containerv1.Container, error) {
	// Check parameters
	if types.IsNil(container) {
		return nil, fmt.Errorf("unable to process nil container")
	}
	if types.IsNil(container.Headers) {
		return nil, fmt.Errorf("unable to process nil container headers")
	}
	if identity == nil {
		return nil, fmt.Errorf("unable to process without container key")
	}

	// Check headers
	if container.Headers.ContentType != containerSealedContentType {
		return nil, fmt.Errorf("unable to unseal container")
	}

	// Check ephemeral container public encryption key
	if len(container.Headers.EncryptionPublicKey) != publicKeySize {
		return nil, fmt.Errorf("invalid container public size")
	}

	// Decode public key
	var publicKey ecdsa.PublicKey
	publicKey.Curve = elliptic.P384()
	publicKey.X, publicKey.Y = elliptic.UnmarshalCompressed(elliptic.P384(), container.Headers.EncryptionPublicKey)
	if publicKey.X == nil {
		return nil, errors.New("invalid container signing public key")
	}

	// Check identity private encryption key
	privRaw := identity.Bytes()
	if len(privRaw) != privateKeySize {
		return nil, fmt.Errorf("invalid identity private key length")
	}
	var pk ecdsa.PrivateKey
	pk.PublicKey.Curve = elliptic.P384()
	pk.D = big.NewInt(0).SetBytes(privRaw)

	// Precompute identifier
	derivedKey, err := deriveSharedKeyFromRecipient(&publicKey, &pk)
	if err != nil {
		return nil, fmt.Errorf("unable to execute key agreement: %w", err)
	}

	// Try recipients
	payloadKey, err := tryRecipientKeys(derivedKey, container.Headers.Recipients)
	if err != nil {
		return nil, fmt.Errorf("unable to unseal container: error occurred during recipient key tests: %w", err)
	}

	// Check private key
	if len(payloadKey) != encryptionKeySize {
		return nil, fmt.Errorf("unable to unseal container: invalid encryption key size")
	}
	var encryptionKey [encryptionKeySize]byte
	copy(encryptionKey[:], payloadKey[:encryptionKeySize])

	// Create AES block cipher
	block, err := aes.NewCipher(payloadKey[:])
	if err != nil {
		return nil, fmt.Errorf("unable to initialize block cipher: %w", err)
	}

	// Initialize AEAD cipher chain
	aead, err := cipher.NewGCMWithNonceSize(block, nonceSize)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize aead chain: %w", err)
	}

	// Prepare sig nonce
	var pubSigNonce [nonceSize]byte
	copy(pubSigNonce[:], "harp_psigkey")

	// Decrypt signing public key
	containerSignKeyRaw, err := decrypt(container.Headers.ContainerBox, pubSigNonce[:], aead)
	if err != nil {
		return nil, fmt.Errorf("invalid container key")
	}
	if len(containerSignKeyRaw) != publicKeySize {
		return nil, fmt.Errorf("unable to unseal container: invalid signature key size")
	}

	// Compute headers hash
	headerHash, err := computeHeaderHash(container.Headers)
	if err != nil {
		return nil, fmt.Errorf("unable to compute header hash: %w", err)
	}

	// Extract payload nonce
	var payloadNonce [nonceSize]byte
	copy(payloadNonce[:], headerHash[:nonceSize])

	// Decrypt payload
	payloadRaw, err := decrypt(container.Raw, payloadNonce[:], aead)
	if err != nil || len(payloadRaw) < ed25519.SignatureSize {
		return nil, fmt.Errorf("invalid ciphered content")
	}

	// Prepare protected content
	protected := bytes.Buffer{}
	protected.Write([]byte("harp fips encrypted signature"))
	protected.WriteByte(0x00)
	protected.Write(headerHash)
	contentHash := sha512.Sum512(payloadRaw)
	protected.Write(contentHash[:])

	// Extract signature / content
	detachedSig := payloadRaw[:signatureSize]
	content := payloadRaw[signatureSize:]

	spew.Dump(payloadRaw)

	// Compute SHA-384 checksum
	digest := sha512.Sum384(protected.Bytes())

	var (
		r = big.NewInt(0).SetBytes(detachedSig[:48])
		s = big.NewInt(0).SetBytes(detachedSig[48:])
	)
	// Validate signature
	if ecdsa.Verify(&publicKey, digest[:], r, s) {
		return nil, fmt.Errorf("unable to verify protected content: %w", err)
	}

	// Unmarshal inner container
	out := &containerv1.Container{}
	if err := proto.Unmarshal(content, out); err != nil {
		return nil, fmt.Errorf("unable to unpack inner content: %w", err)
	}

	// No error
	return out, nil
}

// -----------------------------------------------------------------------------

func tryRecipientKeys(derivedKey *[32]byte, recipients []*containerv1.Recipient) ([]byte, error) {
	// Calculate recipient identifier
	identifier, err := keyIdentifierFromDerivedKey(derivedKey)
	if err != nil {
		return nil, fmt.Errorf("unable to generate identifier: %w", err)
	}

	// Create AES block cipher
	block, err := aes.NewCipher(derivedKey[:])
	if err != nil {
		return nil, fmt.Errorf("unable to initialize block cipher: %w", err)
	}

	// Initialize AEAD cipher chain
	aead, err := cipher.NewGCMWithNonceSize(block, nonceSize)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize aead chain: %w", err)
	}

	// Find matching recipient
	for _, r := range recipients {
		// Check recipient identifiers
		if !security.SecureCompare(identifier, r.Identifier) {
			continue
		}

		var nonce [nonceSize]byte
		copy(nonce[:], r.Key[:nonceSize])

		// Try to decrypt the secretbox with the derived key.
		payloadKey, err := decrypt(r.Key[nonceSize:], nonce[:], aead)
		if err != nil {
			return nil, fmt.Errorf("invalid recipient encryption key")
		}

		// Encryption key found, return no error.
		return payloadKey, nil
	}

	// No recipient found in list.
	return nil, fmt.Errorf("no recipient found")
}
