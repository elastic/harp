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
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/awnumar/memguard"
	"go.uber.org/zap"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/nacl/box"

	"github.com/elastic/harp/pkg/container"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/sdk/security/crypto/bech32"
	"github.com/elastic/harp/pkg/sdk/security/crypto/x25519"
	"github.com/elastic/harp/pkg/sdk/types"
	"github.com/elastic/harp/pkg/tasks"
)

// SealTask implements secret container sealing task.
type SealTask struct {
	ContainerReader          tasks.ReaderProvider
	SealedContainerWriter    tasks.WriterProvider
	OutputWriter             tasks.WriterProvider
	Identities               []string
	DCKDMasterKey            *memguard.LockedBuffer
	DCKDTarget               string
	JSONOutput               bool
	DisableContainerIdentity bool
}

// Run the task.
//nolint:funlen,gocyclo // To refactor
func (t *SealTask) Run(ctx context.Context) error {
	// Create input reader
	reader, err := t.ContainerReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to open input reader: %w", err)
	}

	// Load input container
	in, err := container.Load(reader)
	if err != nil {
		return fmt.Errorf("unable to read input container: %w", err)
	}

	// Open output file
	writer, err := t.SealedContainerWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to create output bundle: %w", err)
	}

	// If using sealing seed
	peerPublicKeys := []*[32]byte{}

	// Given identities
	if len(t.Identities) == 0 {
		return fmt.Errorf("at least one sealing identity must be provided for recovery")
	}

	// Filter identities
	var filteredIdentities types.StringArray
	// Process identities (nizk proof of private key knowledge for private key ownership proof ?)
	for _, id := range t.Identities {
		// Check if identity is already added
		if !filteredIdentities.AddIfNotContains(id) {
			continue
		}

		// Check encoding
		hrp, publicKeyRaw, errDecode := bech32.Decode(id)
		if errDecode != nil {
			return fmt.Errorf("invalid '%s' as public identity: %w", id, errDecode)
		}

		// Validate public key
		if !x25519.IsValidPublicKey(publicKeyRaw) {
			log.For(ctx).Warn("Public key ignored, it looks invalid", zap.String("key", id), zap.String("hrp", hrp))
			continue
		}

		// Copy public key
		var publicKey [32]byte
		copy(publicKey[:], publicKeyRaw[:32])

		// Append to identity
		peerPublicKeys = append(peerPublicKeys, &publicKey)
	}

	var containerKey string

	if !t.DisableContainerIdentity {
		// Generate container key
		containerPublicKey, containerPrivateKey, errContainerGen := t.generateContainerKey()
		if errContainerGen != nil {
			return fmt.Errorf("unable to generate container key: %w", errContainerGen)
		}

		// Append to identity
		peerPublicKeys = append(peerPublicKeys, containerPublicKey)
		containerKey = base64.RawURLEncoding.EncodeToString(containerPrivateKey[:])
	}

	// Seal the container
	sealedContainer, err := container.Seal(in, peerPublicKeys...)
	if err != nil {
		return fmt.Errorf("unable to seal container: %w", err)
	}

	// Dump to writer
	if err = container.Dump(writer, sealedContainer); err != nil {
		return fmt.Errorf("unable to write sealed container: %w", err)
	}

	// Get output writer
	outputWriter, err := t.OutputWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve output writer: %w", err)
	}

	if containerKey != "" {
		// Display as json
		if t.JSONOutput {
			if err := json.NewEncoder(outputWriter).Encode(map[string]interface{}{
				"container_key": containerKey,
			}); err != nil {
				return fmt.Errorf("unable to display as json: %w", err)
			}
		} else {
			// Display container key
			fmt.Fprintf(outputWriter, "Container key : %s\n", containerKey)
		}
	}

	// No error
	return nil
}

func (t *SealTask) generateContainerKey() (*[32]byte, *[32]byte, error) {
	// Generate random container key
	seed := rand.Reader

	// Master key derivation
	if t.DCKDMasterKey != nil {
		// Argon2ID(masterKey, Blake2B-512(Target), 1, 64Mb, 4, 64)
		// Don't clean bytes, already done by memguard.
		masterKey := t.DCKDMasterKey.Bytes()

		// Generate deterministic salt
		salt := blake2b.Sum512([]byte(t.DCKDTarget))
		defer memguard.WipeBytes(salt[:])

		// Derive deterministic container key using Argon2id
		dk := argon2.IDKey(masterKey[:32], salt[:], 1, 64*1024, 4, 64)
		defer memguard.WipeBytes(dk)

		// Assign to seed
		seed = bytes.NewBuffer(dk)
	}

	// Generate container key
	pub, priv, errGen := box.GenerateKey(seed)
	if errGen != nil {
		return nil, nil, fmt.Errorf("unable to generate container key: %w", errGen)
	}

	// No error
	return pub, priv, nil
}
