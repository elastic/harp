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
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"fmt"
	"io"

	containerv1 "github.com/elastic/harp/api/gen/go/harp/container/v1"
	"github.com/elastic/harp/pkg/sdk/types"
)

// Seal a secret container
func (a *adapter) Seal(rand io.Reader, container *containerv1.Container, encodedPeersPublicKey ...string) (*containerv1.Container, error) {
	// Check parameters
	if types.IsNil(container) {
		return nil, fmt.Errorf("unable to process nil container")
	}
	if types.IsNil(container.Headers) {
		return nil, fmt.Errorf("unable to process nil container headers")
	}
	if len(encodedPeersPublicKey) == 0 {
		return nil, fmt.Errorf("unable to process empty public keys")
	}

	// Convert public keys
	peersPublicKey, err := a.publicKeys(encodedPeersPublicKey...)
	if err != nil {
		return nil, fmt.Errorf("unable to convert peer public keys: %w", err)
	}

	// Generate encryption key
	payloadKey, block, err := generatedEncryptionKey(rand)
	if err != nil {
		return nil, fmt.Errorf("unable to generate encryption key: %w", err)
	}

	// Prepare signature identity
	sigPriv, encryptedPubSig, err := prepareSignature(block)
	if err != nil {
		return nil, fmt.Errorf("unable to prepare signature materials: %w", err)
	}

	// Generate ephemeral encryption key
	encPriv, err := ecdsa.GenerateKey(encryptionCurve, rand)
	if err != nil {
		return nil, fmt.Errorf("unable to generate ephemeral encryption keypair")
	}

	// Prepare sealed container
	containerHeaders := &containerv1.Header{
		ContentType:         containerSealedContentType,
		EncryptionPublicKey: elliptic.MarshalCompressed(encPriv.Curve, encPriv.PublicKey.X, encPriv.PublicKey.Y),
		ContainerBox:        encryptedPubSig,
		Recipients:          []*containerv1.Recipient{},
		SealVersion:         SealVersion,
	}

	// Process recipients
	for _, peerPublicKey := range peersPublicKey {
		// Ignore nil key
		if peerPublicKey == nil {
			continue
		}

		// Pack recipient using its public key
		r, errPack := packRecipient(rand, payloadKey, encPriv, peerPublicKey)
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

	// Initialize AEAD cipher chain
	aead, err := cipher.NewGCMWithNonceSize(block, nonceSize)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize aead chain: %w", err)
	}

	// Encrypt payload
	encryptedPayload, err := encrypt(append(containerSig, content...), sigNonce, aead)
	if err != nil {
		return nil, fmt.Errorf("unable to encrypt container data: %w", err)
	}

	// No error
	return &containerv1.Container{
		Headers: containerHeaders,
		Raw:     encryptedPayload,
	}, nil
}
