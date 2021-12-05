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
	"errors"
	"fmt"

	"github.com/elastic/harp/build/fips"
	"github.com/elastic/harp/pkg/container/identity"
	"github.com/elastic/harp/pkg/container/identity/key"
	"github.com/elastic/harp/pkg/sdk/types"
	"github.com/elastic/harp/pkg/sdk/value"
	"github.com/elastic/harp/pkg/tasks"
)

type IdentityVersion uint

const (
	LegacyIdentity IdentityVersion = 1
	ModernIdentity IdentityVersion = 2
	NISTIdentity   IdentityVersion = 3
)

// IdentityTask implements secret container identity creation task.
type IdentityTask struct {
	OutputWriter tasks.WriterProvider
	Description  string
	Transformer  value.Transformer
	Version      IdentityVersion
}

// Run the task.
func (t *IdentityTask) Run(ctx context.Context) error {
	// Check arguments
	if types.IsNil(t.OutputWriter) {
		return errors.New("unable to run task with a nil outputWriter provider")
	}
	if types.IsNil(t.Transformer) {
		return errors.New("unable to run task with a nil transformer")
	}
	if t.Description == "" {
		return fmt.Errorf("description must not be blank")
	}

	// Select appropriate strategy.
	var generator identity.PrivateKeyGeneratorFunc

	if fips.Enabled() {
		generator = key.P384
	} else {
		switch t.Version {
		case LegacyIdentity:
			generator = key.Legacy
		case ModernIdentity:
			generator = key.Ed25519
		case NISTIdentity:
			generator = key.P384
		default:
			return fmt.Errorf("invalid or unsupported identity version '%d'", t.Version)
		}
	}

	// Create identity
	id, payload, err := identity.New(rand.Reader, t.Description, generator)
	if err != nil {
		return fmt.Errorf("unable to create a new identity: %w", err)
	}

	// Encrypt the private key.
	identityPrivate, err := t.Transformer.To(ctx, payload)
	if err != nil {
		return fmt.Errorf("unable to encrypt the private identity key: %w", err)
	}

	// Assign private key
	id.Private = &identity.PrivateKey{
		Content: base64.RawURLEncoding.EncodeToString(identityPrivate),
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
