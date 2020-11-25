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
	"github.com/dchest/uniuri"
	"golang.org/x/crypto/blake2b"
	"gopkg.in/square/go-jose.v2"

	"github.com/elastic/harp/pkg/container/identity"
	"github.com/elastic/harp/pkg/tasks"
	"github.com/elastic/harp/pkg/vault"
)

// PBKDF2SaltSize is the default size of the salt for PBKDF2, 128-bit salt.
const PBKDF2SaltSize = 16

// PBKDF2Iterations is the default number of iterations for PBKDF2, 100k
// iterations. Nist recommends at least 10k, 1Passsword uses 100k.
const PBKDF2Iterations = 500001

type jsonWebKey struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	D   string `json:"d"`
}

// IdentityTask implements secret container identity creation task.
type IdentityTask struct {
	OutputWriter     tasks.WriterProvider
	Description      string
	PassPhrase       *memguard.LockedBuffer
	VaultTransitPath string
	VaultTransitKey  string
}

// Run the task.
//nolint:gocyclo // to refactor
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
	id, payload, err := identity.New(t.Description)
	if err != nil {
		return err
	}

	if t.PassPhrase != nil {
		pk, errPassPhrase := t.sealWithPassPhrase(ctx, payload)
		if errPassPhrase != nil {
			return fmt.Errorf("unable to seal identity using passphrase: %v", errPassPhrase)
		}

		// Assign to identity
		id.Private = pk
	}
	if t.VaultTransitKey != "" {
		pk, errVault := t.sealWithVaultTransitKey(ctx, payload)
		if errVault != nil {
			return fmt.Errorf("unable to seal identity using Vault: %v", errVault)
		}

		// Assign to identity
		id.Private = pk
	}

	// Check unhandled identity error
	if id.Private == nil {
		return fmt.Errorf("invalid identity generated")
	}

	// Retrieve output writer
	writer, err := t.OutputWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve output writer handle: %v", err)
	}

	// Create identity output
	if err := json.NewEncoder(writer).Encode(id); err != nil {
		return fmt.Errorf("unable to serialize final identity: %v", err)
	}

	// No error
	return nil
}

func (t *IdentityTask) sealWithPassPhrase(_ context.Context, payload []byte) (*identity.PrivateKey, error) {
	// Encrypt JWK using PBES2
	recipient := jose.Recipient{
		Algorithm:  jose.PBES2_HS512_A256KW,
		Key:        t.PassPhrase.Bytes(),
		PBES2Count: PBKDF2Iterations,
		PBES2Salt:  []byte(uniuri.NewLen(PBKDF2SaltSize)),
	}

	// JWE Header
	opts := new(jose.EncrypterOptions)
	opts.WithContentType(jose.ContentType("jwk+json"))

	// Prepare encryption using AES-256GCM
	encrypter, err := jose.NewEncrypter(jose.A256GCM, recipient, opts)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize encrypter: %v", err)
	}

	// Encrypt the Identity JWK
	jwe, err := encrypter.Encrypt(payload)
	if err != nil {
		return nil, fmt.Errorf("unable to encrypt identity keypair: %v", err)
	}

	// Assemble final JWE
	identityPrivate, err := jwe.CompactSerialize()
	if err != nil {
		panic(err)
	}

	// Wrap private key
	return &identity.PrivateKey{
		Encoding: "jwe",
		Content:  identityPrivate,
	}, nil
}

func (t *IdentityTask) sealWithVaultTransitKey(ctx context.Context, payload []byte) (*identity.PrivateKey, error) {
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

	// Encrypt private key
	cipherText, err := s.Encrypt(ctx, payload)
	if err != nil {
		return nil, err
	}

	// Prepare key ID
	h := blake2b.Sum256([]byte(fmt.Sprintf("%s/%s", t.VaultTransitPath, t.VaultTransitKey)))

	// No error
	return &identity.PrivateKey{
		Encoding: fmt.Sprintf("kms:vault:%s", base64.RawURLEncoding.EncodeToString(h[:])),
		Content:  string(cipherText),
	}, nil
}
