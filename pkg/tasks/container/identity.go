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
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/awnumar/memguard"

	"github.com/elastic/harp/pkg/container/identity"
	"github.com/elastic/harp/pkg/sdk/value"
	"github.com/elastic/harp/pkg/sdk/value/encryption/jwe"
	"github.com/elastic/harp/pkg/tasks"
	"github.com/elastic/harp/pkg/vault"
)

// IdentityTask implements secret container identity creation task.
type IdentityTask struct {
	OutputWriter     tasks.WriterProvider
	Description      string
	PassPhrase       *memguard.LockedBuffer
	VaultTransitPath string
	VaultTransitKey  string
}

// Run the task.
func (t *IdentityTask) Run(ctx context.Context) error {
	// Check arguments
	if t.Description == "" {
		return fmt.Errorf("description must not be blank")
	}

	// Check exclusive parameters
	if t.PassPhrase == nil && t.VaultTransitKey == "" {
		return fmt.Errorf("passphrase or vaultTransitKey must be defined")
	}
	if t.PassPhrase != nil && t.PassPhrase.Size() > 0 && t.VaultTransitKey != "" {
		return fmt.Errorf("passphrase and vaultTransitKey are mutually exclusive")
	}

	// Create identity
	id, payload, err := identity.New(rand.Reader, t.Description)
	if err != nil {
		return fmt.Errorf("unable to create a new identity: %w", err)
	}

	var (
		transform      value.Transformer
		encoding       string
		errTransformer error
	)
	switch {
	case t.PassPhrase != nil:
		transform, errTransformer = jwe.Transformer(jwe.TransformerKey(jwe.PBES2_HS512_A256KW, t.PassPhrase.String()))
		encoding = "jwe"
	case t.VaultTransitKey != "":
		transform, errTransformer = vault.Transformer(vault.TransformerKey(t.VaultTransitPath, t.VaultTransitKey, vault.AESGCM))
		encoding = "vault"
	default:
		return fmt.Errorf("a passphrase or a vault transit key must be specified")
	}
	if errTransformer != nil {
		return fmt.Errorf("unable to initialize identity transformer: %w", errTransformer)
	}

	// Encrypt the private key.
	identityPrivate, err := transform.To(ctx, payload)
	if err != nil {
		return fmt.Errorf("unable to encrypt the private identity key: %w", err)
	}

	// Assign private key
	id.Private = &identity.PrivateKey{
		Encoding: encoding,
		Content:  base64.RawURLEncoding.EncodeToString(identityPrivate),
	}

	// Retrieve output writer
	writer, err := t.OutputWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve output writer handle: %w", err)
	}

	// Create identity output
	if err := json.NewEncoder(writer).Encode(id); err != nil {
		return fmt.Errorf("unable to serialize final identity: %w", err)
	}

	// No error
	return nil
}
