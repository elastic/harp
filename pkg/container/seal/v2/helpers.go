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

	"golang.org/x/crypto/hkdf"

	"github.com/awnumar/memguard"
	"google.golang.org/protobuf/proto"

	containerv1 "github.com/elastic/harp/api/gen/go/harp/container/v1"
	"github.com/elastic/harp/pkg/sdk/security"
)

func tryRecipientKeys(derivedKey *[32]byte, recipients []*containerv1.Recipient) (*[32]byte, error) {
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

		// Try to decrypt the secretbox with the derived key.
		clearText, err := decrypt(r.Key, derivedKey)
		if err != nil {
			return nil, fmt.Errorf("invalid recipient encryption key")
		}

		var payloadKey [32]byte
		copy(payloadKey[:], clearText)

		// Encryption key found, return no error.
		return &payloadKey, nil
	}

	// No recipient found in list.
	return nil, fmt.Errorf("no recipient found")
}

func prepareSignature(rand io.Reader, encryptionKey *[32]byte) (*ecdsa.PrivateKey, []byte, error) {
	// Generate ephemeral signing key
	sigPriv, err := ecdsa.GenerateKey(signatureCurve, cryptorand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate signing keypair")
	}

	// Compress public key point
	sigPub := elliptic.MarshalCompressed(sigPriv.Curve, sigPriv.PublicKey.X, sigPriv.PublicKey.Y)

	// Encrypt public signing key
	encryptedPubSig, err := encrypt(rand, sigPub, encryptionKey)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to encrypt public signing key: %w", err)
	}

	// Cleanup
	memguard.WipeBytes(sigPub)

	// No error
	return sigPriv, encryptedPubSig, nil
}

func signContainer(sigPriv *ecdsa.PrivateKey, headers *containerv1.Header, container *containerv1.Container) (content, containerSig []byte, err error) {
	// Serialize protobuf payload
	content, err = proto.Marshal(container)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to encode container content: %w", err)
	}

	// Compute header hash
	headerHash, err := computeHeaderHash(headers)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to compute header hash: %w", err)
	}

	// Compute protected content hash
	protectedHash := computeProtectedHash(headerHash, content)

	// Sign the protected content
	r, s, err := ecdsa.Sign(cryptorand.Reader, sigPriv, protectedHash)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to sign protected content: %w", err)
	}

	// Container signature
	containerSig = append(r.Bytes(), s.Bytes()...)

	// No error
	return content, containerSig, nil
}

func generatedEncryptionKey(rand io.Reader) (*[32]byte, error) {
	// Generate payload encryption key
	var payloadKey [encryptionKeySize]byte
	if _, err := io.ReadFull(rand, payloadKey[:]); err != nil {
		return nil, fmt.Errorf("unable to generate payload key for encryption")
	}

	// No error
	return &payloadKey, nil
}

func encrypt(rand io.Reader, plaintext []byte, key *[32]byte) ([]byte, error) {
	// Check cleartext message size.
	if len(plaintext) > messageLimit {
		return nil, errors.New("value too large")
	}

	// Generate random nonce
	var seed [seedSize]byte
	if _, err := io.ReadFull(rand, seed[:]); err != nil {
		return nil, fmt.Errorf("unable to generate random encryption nonce: %w", err)
	}

	// Derive keys from seed and secret key
	ek, n2, ak, err := kdf(key, seed[:])
	if err != nil {
		return nil, fmt.Errorf("unable to derive keys from seed: %w", err)
	}

	// Prepare an AES-256-CTR stream cipher
	block, err := aes.NewCipher(ek)
	if err != nil {
		return nil, fmt.Errorf("unable to prepare block cipher: %w", err)
	}
	ciph := cipher.NewCTR(block, n2)

	// Encrypt the payload
	c := make([]byte, len(plaintext))
	ciph.XORKeyStream(c, plaintext)

	// Compute MAC
	t, err := mac(ak, seed[:], c)
	if err != nil {
		return nil, fmt.Errorf("paseto: unable to compute MAC: %w", err)
	}

	// Serialize final payload
	// n || c || t
	body := append([]byte{}, seed[:]...)
	body = append(body, c...)
	body = append(body, t...)

	// No error
	return body, nil
}

func decrypt(ciphertext []byte, key *[32]byte) ([]byte, error) {
	// Check arguments
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphered text too short")
	}

	// Extract components
	n := ciphertext[:seedSize]
	t := ciphertext[len(ciphertext)-macSize:]
	c := ciphertext[seedSize : len(ciphertext)-macSize]

	// Derive keys from seed and secret key
	ek, n2, ak, err := kdf(key, n)
	if err != nil {
		return nil, fmt.Errorf("unable to derive keys from seed: %w", err)
	}

	// Compute MAC
	t2, err := mac(ak, n, c)
	if err != nil {
		return nil, fmt.Errorf("unable to compute MAC: %w", err)
	}

	// Time-constant compare MAC
	if !security.SecureCompare(t, t2) {
		return nil, errors.New("invalid pre-authentication header")
	}

	// Prepare an AES-256-CTR stream cipher
	block, err := aes.NewCipher(ek)
	if err != nil {
		return nil, fmt.Errorf("unable to prepare block cipher: %w", err)
	}
	ciph := cipher.NewCTR(block, n2)

	// Decrypt the payload
	m := make([]byte, len(c))
	ciph.XORKeyStream(m, c)

	// No error
	return m, nil
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

	// Encrypt the payload key
	encryptedKey, err := encrypt(rand, payloadKey[:], recipientKey)
	if err != nil {
		return nil, fmt.Errorf("unable to encrypt payload key for recipient: %w", err)
	}

	// Pack recipient
	recipient := &containerv1.Recipient{
		Identifier: identifier,
		Key:        encryptedKey,
	}

	// Return recipient
	return recipient, nil
}

func deriveSharedKeyFromRecipient(publicKey *ecdsa.PublicKey, privateKey *ecdsa.PrivateKey) (*[32]byte, error) {
	// Compute Z - ECDH(localPrivate, remotePublic)
	Z, _ := privateKey.Curve.ScalarMult(publicKey.X, publicKey.Y, privateKey.D.Bytes())

	// Prepare info: ( AlgorithmID || PartyInfo || KeyLength )
	fixedInfo := []byte{}
	fixedInfo = append(fixedInfo, lengthPrefixedArray([]byte("A256CTR"))...)
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

func kdf(key *[32]byte, n []byte) (ek, n2, ak []byte, err error) {
	// Check arguments
	if key == nil {
		return nil, nil, nil, errors.New("unable to derive keys from a nil seed")
	}

	// Prepare HKDF-HMAC-SHA384
	encKDF := hkdf.New(sha512.New384, key[:], nil, append([]byte("harp-encryption-key-v2"), n...))

	// Derive encryption key
	tmp := make([]byte, encryptionKeySize+nonceSize)
	if _, err := io.ReadFull(encKDF, tmp); err != nil {
		return nil, nil, nil, fmt.Errorf("unable to generate encryption key from seed: %w", err)
	}

	// Split encryption key (Ek) and nonce (n2)
	ek = tmp[:encryptionKeySize]
	n2 = tmp[encryptionKeySize:]

	// Derive authentication key
	authKDF := hkdf.New(sha512.New384, key[:], nil, append([]byte("harp-auth-key-for-aead"), n...))

	// Derive authentication key
	ak = make([]byte, nonceSize)
	if _, err := io.ReadFull(authKDF, ak); err != nil {
		return nil, nil, nil, fmt.Errorf("unable to generate authentication key from seed: %w", err)
	}

	// No error
	return ek, n2, ak, nil
}

func mac(ak, n, c []byte) ([]byte, error) {
	// Compute pre-authenticated content
	preAuth, err := pae([]byte("harp-authentication-tag-v2"), n, c)
	if err != nil {
		return nil, err
	}

	// Compute MAC
	mac := hmac.New(sha512.New384, ak)

	// Hash pre-authentication content
	mac.Write(preAuth)

	// No error
	return mac.Sum(nil), nil
}

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
