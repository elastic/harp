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
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/awnumar/memguard"

	"github.com/elastic/harp/pkg/container/identity"
	"github.com/elastic/harp/pkg/sdk/security"
	"github.com/elastic/harp/pkg/sdk/security/crypto/bech32"
	"github.com/elastic/harp/pkg/sdk/value"
	"github.com/elastic/harp/pkg/sdk/value/encryption/jwe"
	"github.com/elastic/harp/pkg/tasks"
	"github.com/elastic/harp/pkg/vault"
)

// RecoverTask implements secret container identity recovery task.
type RecoverTask struct {
	JSONReader       tasks.ReaderProvider
	OutputWriter     tasks.WriterProvider
	PassPhrase       *memguard.LockedBuffer
	VaultTransitPath string
	VaultTransitKey  string
	JSONOutput       bool
}

// Run the task.
//nolint:gocyclo // To refactor
func (t *RecoverTask) Run(ctx context.Context) error {
	// Check exclusive parameters
	if t.PassPhrase == nil && t.VaultTransitKey == "" {
		return fmt.Errorf("passphrase or vaultTransitKey must be defined")
	}
	if t.PassPhrase != nil && t.PassPhrase.Size() > 0 && t.VaultTransitKey != "" {
		return fmt.Errorf("passphrase and vaultTransitKey are mutually exclusive")
	}

	// Create input reader
	reader, err := t.JSONReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to read input reader: %w", err)
	}

	// Extract from reader
	input, err := identity.FromReader(reader)
	if err != nil {
		return fmt.Errorf("unable to extract an identity from reader: %w", err)
	}

	var (
		transform      value.Transformer
		errTransformer error
	)
	switch {
	case t.PassPhrase != nil:
		transform, errTransformer = jwe.Transformer(jwe.TransformerKey(jwe.PBES2_HS512_A256KW, t.PassPhrase.String()))
	case t.VaultTransitKey != "":
		transform, errTransformer = vault.Transformer(vault.TransformerKey(t.VaultTransitPath, t.VaultTransitKey, vault.AESGCM))
	default:
		return fmt.Errorf("a passphrase or a vault transit key must be specified")
	}
	if errTransformer != nil {
		return fmt.Errorf("unable to initialize identity transformer: %w", errTransformer)
	}

	// Try to decrypt the private key
	key, err := input.Decrypt(ctx, transform)
	if err != nil {
		return fmt.Errorf("unable to decrypt private key: %w", err)
	}

	// Build public key
	_, pubKey, err := bech32.Decode(input.Public)
	if err != nil {
		return fmt.Errorf("invalid public key encoding: %w", err)
	}

	// Decode base64 public key
	pubKeyRaw, err := base64.RawURLEncoding.DecodeString(key.X)
	if err != nil {
		return fmt.Errorf("invalid public key, the decoded public is corrupted")
	}

	// Check validity
	if !security.SecureCompare(pubKey, pubKeyRaw) {
		return fmt.Errorf("invalid identity, key mismatch detected")
	}

	// Retrieve recoevery key
	recoveryPrivateKey, err := identity.RecoveryKey(key)
	if err != nil {
		return fmt.Errorf("unable to retrieve recovery key from identity: %w", err)
	}

	// Get output writer
	outputWriter, err := t.OutputWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve output writer: %w", err)
	}

	// Display as json
	if t.JSONOutput {
		if err := json.NewEncoder(outputWriter).Encode(map[string]interface{}{
			"container_key": base64.RawURLEncoding.EncodeToString(recoveryPrivateKey[:]),
		}); err != nil {
			return fmt.Errorf("unable to display as json: %w", err)
		}
	} else {
		// Display container key
		fmt.Fprintf(outputWriter, "Container key : %s\n", base64.RawURLEncoding.EncodeToString(recoveryPrivateKey[:]))
	}

	// No error
	return nil
}
