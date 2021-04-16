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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/awnumar/memguard"
	"gopkg.in/square/go-jose.v2"

	"github.com/elastic/harp/pkg/container/identity"
	"github.com/elastic/harp/pkg/sdk/security"
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
		payload    []byte
		errDecrypt error
	)

	switch {
	case input.Private.Encoding == "jwe":
		// Parse JWE Token
		jwe, errParse := jose.ParseEncrypted(input.Private.Content)
		if errParse != nil {
			return fmt.Errorf("unable to parse JWE token")
		}

		// Try to decrypt with given passphrase
		payload, errDecrypt = jwe.Decrypt(t.PassPhrase.Bytes())
		if errDecrypt != nil {
			return fmt.Errorf("unable to decrypt JWE token")
		}
	case strings.HasPrefix(input.Private.Encoding, "kms:vault:"):
		payload, errDecrypt = t.unsealWithVaultTransitKey(ctx, input.Private.Content)
		if errDecrypt != nil {
			return fmt.Errorf("unable to decrypt using Vault")
		}
	default:
		return fmt.Errorf("unknown private key encoding '%s'", input.Private.Encoding)
	}

	// Decode key
	var key jsonWebKey
	if err = json.NewDecoder(bytes.NewReader(payload)).Decode(&key); err != nil {
		return fmt.Errorf("unable to decode payload as JSON: %w", err)
	}

	// Check validity
	if !security.SecureCompareString(input.Public, key.X) {
		return fmt.Errorf("invalid identity, key mismatch detected")
	}

	// Get output writer
	outputWriter, err := t.OutputWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve output writer: %w", err)
	}

	// Display as json
	if t.JSONOutput {
		if err := json.NewEncoder(outputWriter).Encode(map[string]interface{}{
			"container_key": key.D,
		}); err != nil {
			return fmt.Errorf("unable to display as json: %w", err)
		}
	} else {
		// Display container key
		fmt.Fprintf(outputWriter, "Container key : %s\n", key.D)
	}

	// No error
	return nil
}

func (t *RecoverTask) unsealWithVaultTransitKey(ctx context.Context, cipherText string) ([]byte, error) {
	// Connecto to Vault.
	v, err := vault.DefaultClient()
	if err != nil {
		return nil, err
	}

	// Check default value
	if t.VaultTransitPath == "" {
		t.VaultTransitPath = "transit"
	}

	// Build a transit encryption service.
	s, err := v.Transit(t.VaultTransitPath, t.VaultTransitKey)
	if err != nil {
		return nil, err
	}

	// Decrypt private key
	clearText, err := s.Decrypt(ctx, []byte(cipherText))
	if err != nil {
		return nil, err
	}

	// No error
	return clearText, nil
}
