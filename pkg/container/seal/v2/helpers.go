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

package v2

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	cryptorand "crypto/rand"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/awnumar/memguard"
	"golang.org/x/crypto/hkdf"
	"google.golang.org/protobuf/proto"

	containerv1 "github.com/elastic/harp/api/gen/go/harp/container/v1"
	"github.com/elastic/harp/pkg/sdk/security"
	"github.com/elastic/harp/pkg/sdk/types"
)

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

func prepareSignature(block cipher.Block) (*ecdsa.PrivateKey, []byte, error) {
	// Generate ephemeral signing key
	sigPriv, err := ecdsa.GenerateKey(signatureCurve, cryptorand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate signing keypair")
	}

	// Compress public key point
	sigPub := elliptic.MarshalCompressed(sigPriv.Curve, sigPriv.PublicKey.X, sigPriv.PublicKey.Y)

	// Encrypt public signature key
	var pubSigNonce [nonceSize]byte
	copy(pubSigNonce[:], "harp_psigkey")

	// Initialize AEAD cipher chain
	aead, err := cipher.NewGCMWithNonceSize(block, nonceSize)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to initialize aead chain: %w", err)
	}

	// Encrypt public signing key
	encryptedPubSig, err := encrypt(sigPub, pubSigNonce[:], aead)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to encrypt public signing key: %w", err)
	}

	// Cleanup
	memguard.WipeBytes(pubSigNonce[:])
	memguard.WipeBytes(sigPub)

	// No error
	return sigPriv, encryptedPubSig, nil
}

func signContainer(sigPriv *ecdsa.PrivateKey, headers *containerv1.Header, container *containerv1.Container) (content, containerSig, sigNonce []byte, err error) {
	// Serialize protobuf payload
	content, err = proto.Marshal(container)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unable to encode container content: %w", err)
	}

	// Compute header hash
	headerHash, err := computeHeaderHash(headers)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unable to compute header hash: %w", err)
	}

	// Compute protected content hash
	protectedHash := computeProtectedHash(headerHash, content)

	// Sign the protected content
	r, s, err := ecdsa.Sign(cryptorand.Reader, sigPriv, protectedHash)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unable to sign protected content: %w", err)
	}

	// Container signature
	containerSig = append(r.Bytes(), s.Bytes()...)

	// Prepare encryption nonce form sigHash
	sigNonce = make([]byte, nonceSize)
	copy(sigNonce, headerHash[:nonceSize])

	// No error
	return content, containerSig, sigNonce, nil
}

func generatedEncryptionKey(rand io.Reader) (*[32]byte, cipher.Block, error) {
	// Generate payload encryption key
	var payloadKey [encryptionKeySize]byte
	if _, err := io.ReadFull(rand, payloadKey[:]); err != nil {
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

func encrypt(plaintext, n []byte, ciph cipher.AEAD) ([]byte, error) {
	// Check cleartext message size.
	if len(plaintext) > messageLimit {
		return nil, errors.New("value too large")
	}
	if types.IsNil(ciph) {
		return nil, errors.New("aead cipher must not be nil")
	}

	// Allocate output buffer
	out := make([]byte, 0, ciph.Overhead()+len(plaintext))

	// Seal with aead (nonce is handled externally)
	out = ciph.Seal(out, n, plaintext, nil)

	// No error
	return out, nil
}

func decrypt(ciphertext, nonce []byte, ciph cipher.AEAD) ([]byte, error) {
	// Check arguments
	if len(ciphertext) < ciph.NonceSize() {
		return nil, errors.New("ciphered text too short")
	}
	if types.IsNil(ciph) {
		return nil, errors.New("aead cipher must not be nil")
	}

	clearText, err := ciph.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("failed to decrypt given message")
	}

	// No error
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

func computeProtectedHash(headerHash, content []byte) []byte {
	// Prepare protected content
	protected := bytes.Buffer{}
	protected.Write([]byte("harp fips encrypted signature"))
	protected.WriteByte(0x00)
	protected.Write(headerHash)
	contentHash := sha512.Sum512(content)
	protected.Write(contentHash[:])

	// No error
	return protected.Bytes()
}

func packRecipient(rand io.Reader, payloadKey *[32]byte, ephPrivKey *ecdsa.PrivateKey, peerPublicKey *ecdsa.PublicKey) (*containerv1.Recipient, error) {
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
	if _, errRand := io.ReadFull(rand, recipientNonce[:]); errRand != nil {
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
	binary.BigEndian.PutUint32(result, value)

	return result
}
