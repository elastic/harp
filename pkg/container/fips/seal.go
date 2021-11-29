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
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/awnumar/memguard"
	"github.com/davecgh/go-spew/spew"
	containerv1 "github.com/elastic/harp/api/gen/go/harp/container/v1"
	"github.com/elastic/harp/pkg/sdk/types"
	"google.golang.org/protobuf/proto"
)

// Seal a secret container
func Seal(container *containerv1.Container, peersPublicKey ...interface{}) (*containerv1.Container, error) {
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

	// Generate encryption key
	payloadKey, block, err := generatedEncryptionKey()
	if err != nil {
		return nil, fmt.Errorf("unable to generate encryption key: %w", err)
	}

	// Prepare signature identity
	sigPriv, encryptedPubSig, err := prepareSignature(block)
	if err != nil {
		return nil, fmt.Errorf("unable to prepare signature materials: %w", err)
	}

	// Generate ephemeral encryption key
	encPriv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("unable to generate ephemeral encryption keypair")
	}

	// Prepare sealed container
	containerHeaders := &containerv1.Header{
		ContentType:         containerSealedContentType,
		EncryptionPublicKey: elliptic.MarshalCompressed(encPriv.Curve, encPriv.PublicKey.X, encPriv.PublicKey.Y),
		ContainerBox:        encryptedPubSig,
		Recipients:          []*containerv1.Recipient{},
	}

	// Process recipients
	for _, peerPublicKeyRaw := range peersPublicKey {
		// Ignore nil key
		if peerPublicKeyRaw == nil {
			continue
		}
		peerPublicKey, ok := peerPublicKeyRaw.(*ecdsa.PublicKey)
		if !ok {
			continue
		}

		// Pack recipient using its public key
		r, errPack := packRecipient(payloadKey, encPriv, peerPublicKey)
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

	// Sign given container
	content, containerSig, sigNonce, err := signContainer(sigPriv, containerHeaders, container)
	if err != nil {
		return nil, fmt.Errorf("unable to sign container data: %w", err)
	}

	spew.Dump(content)

	// Initialize AEAD cipher chain
	aead, err := cipher.NewGCMWithNonceSize(block, nonceSize)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize aead chain: %w", err)
	}

	// Encrypt payload
	encryptedPayload, err := encrypt(append(containerSig, content...), sigNonce[:], aead)
	if err != nil {
		return nil, fmt.Errorf("unable to encrypt container data: %w", err)
	}

	// No error
	return &containerv1.Container{
		Headers: containerHeaders,
		Raw:     encryptedPayload,
	}, nil
}

// -----------------------------------------------------------------------------

func prepareSignature(block cipher.Block) (*ecdsa.PrivateKey, []byte, error) {
	// Generate ephemeral signing key
	sigPriv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
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
	protectedHash, err := computeProtectedHash(container, headerHash, content)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unable to compute protected hash: %w", err)
	}

	// Sign the protected content
	r, s, err := ecdsa.Sign(rand.Reader, sigPriv, protectedHash[:])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unable to sign protected content: %w", err)
	}

	// Container signature
	containerSig = append(r.Bytes(), s.Bytes()...)

	// Prepare encryption nonce form sigHash
	sigNonce = make([]byte, nonceSize)
	copy(sigNonce[:], headerHash[:nonceSize])

	// No error
	return content, containerSig, sigNonce, nil
}
