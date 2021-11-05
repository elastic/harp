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

package container

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/awnumar/memguard"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"
	"google.golang.org/protobuf/proto"

	containerv1 "github.com/elastic/harp/api/gen/go/harp/container/v1"
	"github.com/elastic/harp/pkg/sdk/types"
)

const (
	containerMagic             = uint32(0x53CB3701)
	containerVersion           = uint16(0x0002)
	containerSealedContentType = "application/vnd.harp.v1.SealedContainer"
	publicKeySize              = 32
	privateKeySize             = 32
	encryptionKeySize          = 32
)

// Load a reader to extract as a container.
func Load(r io.Reader) (*containerv1.Container, error) {
	// Check parameters
	if types.IsNil(r) {
		return nil, fmt.Errorf("unable to process nil reader")
	}

	// Read magic
	var magic uint32
	if err := binary.Read(r, binary.BigEndian, &magic); err != nil {
		return nil, fmt.Errorf("unable to read magic code: %w", err)
	}

	// Check magic value
	if magic != containerMagic {
		return nil, fmt.Errorf("invalid magic signature")
	}

	// Read container version
	var version uint16
	if err := binary.Read(r, binary.BigEndian, &version); err != nil {
		return nil, fmt.Errorf("unable to read container version: %w", err)
	}

	// Check magic value
	if version != containerVersion {
		return nil, fmt.Errorf("invalid container version %d", version)
	}

	// Drain input reader
	decoded, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("unable to container content")
	}

	// Check content length
	if len(decoded) == 0 {
		return nil, fmt.Errorf("container is empty")
	}

	// Deserialize protobuf payload
	container := &containerv1.Container{}
	if err = proto.Unmarshal(decoded, container); err != nil {
		return nil, fmt.Errorf("unable to decode content as container")
	}

	// Check headers
	if container.Headers == nil {
		container.Headers = &containerv1.Header{}
	}

	// No error
	return container, nil
}

// Dump the marshaled container instance to writer.
// nolint:interfacer // Tighly coupled to type
func Dump(w io.Writer, c *containerv1.Container) error {
	// Check parameters
	if types.IsNil(w) {
		return fmt.Errorf("unable to process nil writer")
	}
	if c == nil {
		return fmt.Errorf("unable to process nil container")
	}

	// Serialize protobuf payload
	payload, err := proto.Marshal(c)
	if err != nil {
		return fmt.Errorf("unable to encode container content: %w", err)
	}

	// Write packets
	if err = binary.Write(w, binary.BigEndian, containerMagic); err != nil {
		return fmt.Errorf("unable to write container magic: %w", err)
	}
	if err = binary.Write(w, binary.BigEndian, containerVersion); err != nil {
		return fmt.Errorf("unable to write container version: %w", err)
	}
	if _, err = w.Write(payload); err != nil {
		return fmt.Errorf("unable to write container content: %w", err)
	}

	// No error
	return nil
}

// Seal a secret container
func Seal(container *containerv1.Container, peersPublicKey ...*[32]byte) (*containerv1.Container, error) {
	// Check parameters
	if types.IsNil(container) {
		return nil, fmt.Errorf("unable to process nil container")
	}
	if types.IsNil(container.Headers) {
		return nil, fmt.Errorf("unable to process nil container headers")
	}
	if len(peersPublicKey) == 0 {
		return nil, fmt.Errorf("unable to process empty public keys")
	}

	// Serialize protobuf payload
	content, err := proto.Marshal(container)
	if err != nil {
		return container, fmt.Errorf("unable to encode container content: %w", err)
	}

	// Generate payload encryption key
	var payloadKey [32]byte
	if _, err = io.ReadFull(rand.Reader, payloadKey[:]); err != nil {
		return nil, fmt.Errorf("unable to generate payload key for encryption")
	}

	// Generate ephemeral signing key
	sigPub, sigPriv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, fmt.Errorf("unable to generate signing keypair")
	}

	// Encrypt public signature key
	var pubSigNonce [24]byte
	copy(pubSigNonce[:], "harp_container_psigk_box")
	encryptedPubSig := secretbox.Seal(nil, sigPub, &pubSigNonce, &payloadKey)
	memguard.WipeBytes(pubSigNonce[:])

	// Generate ephemeral encryption key
	encPub, encPriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("unable to generate ephemeral encryption keypair")
	}

	// Prepare sealed container
	containerHeaders := &containerv1.Header{
		ContentType:         containerSealedContentType,
		EncryptionPublicKey: encPub[:],
		ContainerBox:        encryptedPubSig,
		Recipients:          []*containerv1.Recipient{},
	}

	// Process recipients
	for _, peerPublicKey := range peersPublicKey {
		// Ignore nil key
		if peerPublicKey == nil {
			continue
		}

		// Pack recipient using its public key
		r, errPack := packRecipient(&payloadKey, encPriv, peerPublicKey)
		if errPack != nil {
			return nil, fmt.Errorf("unable to pack container recipient (%X): %w", *peerPublicKey, err)
		}

		// Append to container
		containerHeaders.Recipients = append(containerHeaders.Recipients, r)
	}

	// Compute header hash
	headerHash, err := computeHeaderHash(containerHeaders)
	if err != nil {
		return nil, fmt.Errorf("unable to compute header hash: %w", err)
	}

	// Prepare protected content
	protected := bytes.Buffer{}
	protected.Write([]byte("harp encrypted signature"))
	protected.WriteByte(0x00)
	protected.Write(headerHash)
	contentHash := blake2b.Sum512(content)
	protected.Write(contentHash[:])

	// Sign th protected content
	containerSig := ed25519.Sign(sigPriv, protected.Bytes())

	// Prepare encryption nonce form sigHash
	var sigNonce [24]byte
	copy(sigNonce[:], headerHash[:24])

	// No error
	return &containerv1.Container{
		Headers: containerHeaders,
		Raw:     secretbox.Seal(nil, append(containerSig, content...), &sigNonce, &payloadKey),
	}, nil
}

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
	var publicKey [publicKeySize]byte
	copy(publicKey[:], container.Headers.EncryptionPublicKey[:publicKeySize])

	// Check identity private encryption key
	privRaw := identity.Bytes()
	if len(privRaw) != privateKeySize {
		return nil, fmt.Errorf("invalid identity private key length")
	}
	var pk [privateKeySize]byte
	copy(pk[:], privRaw[:privateKeySize])

	// Precompute identifier
	derivedKey := deriveSharedKeyFromRecipient(&publicKey, &pk)

	// Try recipients
	payloadKey, err := tryRecipientKeys(&derivedKey, container.Headers.Recipients)
	if err != nil {
		return nil, fmt.Errorf("unable to unseal container: error occurred during recipient key tests: %w", err)
	}

	// Check private key
	if len(payloadKey) != encryptionKeySize {
		return nil, fmt.Errorf("unable to unseal container: invalid encryption key size")
	}
	var encryptionKey [encryptionKeySize]byte
	copy(encryptionKey[:], payloadKey[:encryptionKeySize])

	// Prepare sig nonce
	var pubSigNonce [24]byte
	copy(pubSigNonce[:], "harp_container_psigk_box")

	// Decrypt signing public key
	containerSignKeyRaw, ok := secretbox.Open(nil, container.Headers.ContainerBox, &pubSigNonce, &encryptionKey)
	if !ok {
		return nil, fmt.Errorf("invalid container key")
	}
	if len(containerSignKeyRaw) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("unable to unseal container: invalid signature key size")
	}

	// Compute headers hash
	headerHash, err := computeHeaderHash(container.Headers)
	if err != nil {
		return nil, fmt.Errorf("unable to compute header hash: %w", err)
	}

	// Extract payload nonce
	var payloadNonce [24]byte
	copy(payloadNonce[:], headerHash[:24])

	// Decrypt payload
	payloadRaw, ok := secretbox.Open(nil, container.Raw, &payloadNonce, &encryptionKey)
	if !ok || len(payloadRaw) < ed25519.SignatureSize {
		return nil, fmt.Errorf("invalid ciphered content")
	}

	// Prepare protected content
	protected := bytes.Buffer{}
	protected.Write([]byte("harp encrypted signature"))
	protected.WriteByte(0x00)
	protected.Write(headerHash)
	contentHash := blake2b.Sum512(payloadRaw[ed25519.SignatureSize:])
	protected.Write(contentHash[:])

	// Extract signature / content
	detachedSig := payloadRaw[:ed25519.SignatureSize]
	content := payloadRaw[ed25519.SignatureSize:]

	// Validate signature
	if !ed25519.Verify(containerSignKeyRaw, protected.Bytes(), detachedSig) {
		return nil, fmt.Errorf("invalid container signature")
	}

	// Unmarshal inner container
	out := &containerv1.Container{}
	if err := proto.Unmarshal(content, out); err != nil {
		return nil, fmt.Errorf("unable to unpack inner content: %w", err)
	}

	// No error
	return out, nil
}
