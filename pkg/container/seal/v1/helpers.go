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

package v1

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/awnumar/memguard"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"
	"google.golang.org/protobuf/proto"

	containerv1 "github.com/elastic/harp/api/gen/go/harp/container/v1"
	"github.com/elastic/harp/pkg/sdk/security"
)

func deriveSharedKeyFromRecipient(publicKey, privateKey *[privateKeySize]byte) *[encryptionKeySize]byte {
	// Prepare nonce
	var nonce [nonceSize]byte
	copy(nonce[:], "harp_derived_id_sboxkey0")

	// Prepare payload
	zero := make([]byte, 32)
	memguard.WipeBytes(zero)

	// Use box as a key agreement function
	var sharedKey [encryptionKeySize]byte
	derivedKey := box.Seal(nil, zero, &nonce, publicKey, privateKey)
	copy(sharedKey[:], derivedKey[len(derivedKey)-encryptionKeySize:])

	// No error
	return &sharedKey
}

func computeHeaderHash(headers *containerv1.Header) ([]byte, error) {
	// Check arguments
	if headers == nil {
		return nil, errors.New("unable process with nil headers")
	}

	// Prepare signature
	header, err := proto.Marshal(headers)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal container headers")
	}

	// Hash serialized proto
	hash := blake2b.Sum512(header)

	// No error
	return hash[:], nil
}

func computeProtectedHash(headerHash, content []byte) []byte {
	// Prepare protected content
	protected := bytes.Buffer{}
	protected.Write([]byte(signatureDomainSeparation))
	protected.WriteByte(0x00)
	protected.Write(headerHash)
	contentHash := blake2b.Sum512(content)
	protected.Write(contentHash[:])

	// No error
	return protected.Bytes()
}

func packRecipient(rand io.Reader, payloadKey, ephPrivKey, peerPublicKey *[publicKeySize]byte) (*containerv1.Recipient, error) {
	// Check arguments
	if payloadKey == nil {
		return nil, fmt.Errorf("unable to proceed with nil payload key")
	}
	if ephPrivKey == nil {
		return nil, fmt.Errorf("unable to proceed with nil private key")
	}
	if peerPublicKey == nil {
		return nil, fmt.Errorf("unable to proceed with nil public key")
	}

	// Create identifier
	recipientKey := deriveSharedKeyFromRecipient(peerPublicKey, ephPrivKey)

	// Calculate identifier
	identifier, err := keyIdentifierFromDerivedKey(recipientKey)
	if err != nil {
		return nil, fmt.Errorf("unable to derive key identifier: %w", err)
	}

	// Generate recipient nonce
	var recipientNonce [nonceSize]byte
	if _, err := io.ReadFull(rand, recipientNonce[:]); err != nil {
		return nil, fmt.Errorf("unable to generate recipient nonce for encryption")
	}

	// Pack recipient
	recipient := &containerv1.Recipient{
		Identifier: identifier,
		Key:        secretbox.Seal(recipientNonce[:], payloadKey[:], &recipientNonce, recipientKey),
	}

	// Return recipient
	return recipient, nil
}

func keyIdentifierFromDerivedKey(derivedKey *[encryptionKeySize]byte) ([]byte, error) {
	// Hash the derived key
	h, err := blake2b.New512([]byte("harp signcryption box key identifier"))
	if err != nil {
		return nil, fmt.Errorf("unable to generate recipient identifier hasher")
	}
	if _, err := h.Write(derivedKey[:]); err != nil {
		return nil, fmt.Errorf("unable to generate recipient identifier")
	}

	// Return 32 bytes trucanted hash.
	return h.Sum(nil)[0:keyIdentifierSize], nil
}

func tryRecipientKeys(derivedKey *[encryptionKeySize]byte, recipients []*containerv1.Recipient) ([]byte, error) {
	// Calculate recipient identifier
	identifier, err := keyIdentifierFromDerivedKey(derivedKey)
	if err != nil {
		return nil, fmt.Errorf("unable to generate identifier: %w", err)
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
		payloadKey, isValid := secretbox.Open(nil, r.Key[nonceSize:], &nonce, derivedKey)
		if !isValid {
			return nil, fmt.Errorf("invalid recipient encryption key")
		}

		// Encryption key found, return no error.
		return payloadKey, nil
	}

	// No recipient found in list.
	return nil, fmt.Errorf("no recipient found")
}
