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
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	containerv1 "github.com/elastic/harp/api/gen/go/harp/container/v1"
	"golang.org/x/crypto/hkdf"
	"google.golang.org/protobuf/proto"
)

func generatedEncryptionKey() (*[32]byte, cipher.Block, error) {
	// Generate payload encryption key
	var payloadKey [encryptionKeySize]byte
	if _, err := io.ReadFull(rand.Reader, payloadKey[:]); err != nil {
		return nil, nil, fmt.Errorf("unable to generate payload key for encryption")
	}

	// Create AES block cipher
	block, err := aes.NewCipher(payloadKey[:])
	if err != nil {
		return nil, block, fmt.Errorf("unable to initialize block cipher: %w", err)
	}

	// No error
	return &payloadKey, block, err
}

func encrypt(plaintext []byte, n []byte, ciph cipher.AEAD) ([]byte, error) {
	if len(plaintext) > 64*1024*1024 {
		return nil, errors.New("value too large")
	}
	out := make([]byte, 0, ciph.Overhead()+len(plaintext))
	out = ciph.Seal(out, n, plaintext, nil)

	return out, nil
}

func decrypt(ciphertext, nonce []byte, ciph cipher.AEAD) ([]byte, error) {
	if len(ciphertext) < ciph.NonceSize() {
		return nil, errors.New("ciphered text too short")
	}

	clearText, err := ciph.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("failed to decrypt given message")
	}

	return clearText, nil
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
	hash := sha512.Sum512(header)

	// No error
	return hash[:], nil
}

func computeProtectedHash(container *containerv1.Container, headerHash, content []byte) ([]byte, error) {
	// Prepare protected content
	protected := bytes.Buffer{}
	protected.Write([]byte("harp fips encrypted signature"))
	protected.WriteByte(0x00)
	protected.Write(headerHash)
	contentHash := sha512.Sum512(content)
	protected.Write(contentHash[:])

	// No error
	return protected.Bytes(), nil
}

func packRecipient(payloadKey *[32]byte, ephPrivKey *ecdsa.PrivateKey, peerPublicKey *ecdsa.PublicKey) (*containerv1.Recipient, error) {
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
	recipientKey, err := deriveSharedKeyFromRecipient(peerPublicKey, ephPrivKey)
	if err != nil {
		return nil, fmt.Errorf("unable to execute key agreement: %w", err)
	}

	// Calculate identifier
	identifier, err := keyIdentifierFromDerivedKey(recipientKey)
	if err != nil {
		return nil, fmt.Errorf("unable to derive key identifier: %w", err)
	}

	// Generate recipient nonce
	var recipientNonce [nonceSize]byte
	if _, err := io.ReadFull(rand.Reader, recipientNonce[:]); err != nil {
		return nil, fmt.Errorf("unable to generate recipient nonce for encryption")
	}

	// Create AES block cipher
	block, err := aes.NewCipher(recipientKey[:])
	if err != nil {
		return nil, fmt.Errorf("unable to initialize block cipher: %w", err)
	}

	// Initialize AEAD cipher chain
	aead, err := cipher.NewGCMWithNonceSize(block, nonceSize)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize aead chain: %w", err)
	}

	// Encrypt the payload key
	encryptedKey, err := encrypt(payloadKey[:], recipientNonce[:], aead)
	if err != nil {
		return nil, fmt.Errorf("unable to encrypt payload key for recipient: %w", err)
	}

	// Pack recipient
	recipient := &containerv1.Recipient{
		Identifier: identifier,
		Key:        append(recipientNonce[:], encryptedKey...),
	}

	// Return recipient
	return recipient, nil
}

func deriveSharedKeyFromRecipient(publicKey *ecdsa.PublicKey, privateKey *ecdsa.PrivateKey) (*[32]byte, error) {
	// Compute Z - ECDH(localPrivate, remotePublic)
	Z, _ := privateKey.Curve.ScalarMult(publicKey.X, publicKey.Y, privateKey.D.Bytes())

	// Prepare info: ( AlgorithmID || PartyInfo || KeyLength )
	fixedInfo := []byte{}
	fixedInfo = append(fixedInfo, lengthPrefixedArray([]byte("A256GCM"))...)
	fixedInfo = append(fixedInfo, uint32ToBytes(encryptionKeySize)...)

	// HKDF-HMAC-SHA512
	kdf := hkdf.New(sha512.New, Z.Bytes(), nil, fixedInfo)

	var sharedSecret [encryptionKeySize]byte
	if _, err := io.ReadFull(kdf, sharedSecret[:]); err != nil {
		return nil, fmt.Errorf("unable to derive shared secret: %w", err)
	}

	// No error
	return &sharedSecret, nil
}

func keyIdentifierFromDerivedKey(derivedKey *[32]byte) ([]byte, error) {
	// HMAC-SHA512
	h := hmac.New(sha512.New, []byte("harp signcryption box key identifier"))
	if _, err := h.Write(derivedKey[:]); err != nil {
		return nil, fmt.Errorf("unable to generate recipient identifier")
	}

	// Return 32 bytes truncated hash.
	return h.Sum(nil)[0:encryptionKeySize], nil
}

func lengthPrefixedArray(value []byte) []byte {
	if len(value) == 0 {
		return []byte{}
	}
	result := make([]byte, 4)
	binary.BigEndian.PutUint32(result, uint32(len(value)))

	return append(result, value...)
}

func uint32ToBytes(value uint32) []byte {
	result := make([]byte, 4)
	binary.BigEndian.PutUint32(result, uint32(value))

	return result
}
