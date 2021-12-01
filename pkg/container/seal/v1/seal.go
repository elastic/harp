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
	"crypto/ed25519"
	"errors"
	"fmt"
	"io"

	"github.com/awnumar/memguard"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"
	"google.golang.org/protobuf/proto"

	containerv1 "github.com/elastic/harp/api/gen/go/harp/container/v1"
	"github.com/elastic/harp/pkg/sdk/security/crypto/extra25519"
	"github.com/elastic/harp/pkg/sdk/types"
)

// Seal a secret container
//nolint:funlen,gocyclo // To refactor
func (a *adapter) Seal(rand io.Reader, container *containerv1.Container, peersPublicKey ...interface{}) (*containerv1.Container, error) {
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

	// Check cleartext message size.
	if len(content) > messageLimit {
		return nil, errors.New("unable to seal the container, container is too large")
	}

	// Generate payload encryption key
	var payloadKey [encryptionKeySize]byte
	if _, err = io.ReadFull(rand, payloadKey[:]); err != nil {
		return nil, fmt.Errorf("unable to generate payload key for encryption")
	}

	// Generate ephemeral signing key
	sigPub, sigPriv, err := ed25519.GenerateKey(rand)
	if err != nil {
		return nil, fmt.Errorf("unable to generate signing keypair")
	}

	// Encrypt public signature key
	var pubSigNonce [nonceSize]byte
	copy(pubSigNonce[:], staticSignatureNonce)
	encryptedPubSig := secretbox.Seal(nil, sigPub, &pubSigNonce, &payloadKey)
	memguard.WipeBytes(pubSigNonce[:])

	// Generate ephemeral encryption key
	encPub, encPriv, err := box.GenerateKey(rand)
	if err != nil {
		return nil, fmt.Errorf("unable to generate ephemeral encryption keypair")
	}

	// Prepare sealed container
	containerHeaders := &containerv1.Header{
		ContentType:         containerSealedContentType,
		EncryptionPublicKey: encPub[:],
		ContainerBox:        encryptedPubSig,
		Recipients:          []*containerv1.Recipient{},
		SealVersion:         SealVersion,
	}

	// Process recipients
	for _, peerPublicKeyRaw := range peersPublicKey {
		if types.IsNil(peerPublicKeyRaw) {
			// Ignore nil key
			continue
		}
		peerPublicKey, ok := peerPublicKeyRaw.(*[publicKeySize]byte)
		if !ok {
			// Ignore invalid key types
			continue
		}
		if extra25519.IsEdLowOrder(peerPublicKey[:]) {
			return nil, fmt.Errorf("unable to process with low order public key")
		}

		// Pack recipient using its public key
		r, errPack := packRecipient(rand, &payloadKey, encPriv, peerPublicKey)
		if errPack != nil {
			return nil, fmt.Errorf("unable to pack container recipient (%X): %w", *peerPublicKey, err)
		}

		// Append to container
		containerHeaders.Recipients = append(containerHeaders.Recipients, r)
	}

	// Sanity check
	if len(containerHeaders.Recipients) == 0 {
		return nil, errors.New("unable to seal a container without recipients")
	}

	// Compute header hash
	headerHash, err := computeHeaderHash(containerHeaders)
	if err != nil {
		return nil, fmt.Errorf("unable to compute header hash: %w", err)
	}

	// Prepare protected content
	protectedHash := computeProtectedHash(headerHash, content)

	// Sign th protected content
	containerSig := ed25519.Sign(sigPriv, protectedHash)

	// Prepare encryption nonce form sigHash
	var sigNonce [nonceSize]byte
	copy(sigNonce[:], headerHash[:nonceSize])

	// No error
	return &containerv1.Container{
		Headers: containerHeaders,
		Raw:     secretbox.Seal(nil, append(containerSig, content...), &sigNonce, &payloadKey),
	}, nil
}
