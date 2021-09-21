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

package to

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"

	"github.com/elastic/harp/pkg/bundle"
	bundlevault "github.com/elastic/harp/pkg/bundle/vault"
	"github.com/elastic/harp/pkg/tasks"
	"github.com/elastic/harp/pkg/vault"
)

// VaultTask implements secret-container publication process to Vault.
type VaultTask struct {
	ContainerReader tasks.ReaderProvider
	BackendPrefix   string
	PushMetadata    bool
	AsVaultMetadata bool
	VaultNamespace  string
	MaxWorkerCount  int64
}

// Run the task.
func (t *VaultTask) Run(ctx context.Context) error {
	// Initialize vault connection
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return fmt.Errorf("unable to initialize Vault connection: %w", err)
	}

	// If a namespace is specified
	if t.VaultNamespace != "" {
		client.SetNamespace(t.VaultNamespace)
	}

	// Verify vault connection
	if _, errAuth := vault.CheckAuthentication(client); errAuth != nil {
		return fmt.Errorf("vault connection verification failed: %w", errAuth)
	}

	// Create the reader
	reader, err := t.ContainerReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to open input bundle reader: %w", err)
	}

	// Extract bundle from container
	b, err := bundle.FromContainerReader(reader)
	if err != nil {
		return fmt.Errorf("unable to load bundle: %w", err)
	}

	// Process push operation
	if err := bundlevault.Push(ctx, b, client,
		bundlevault.WithPrefix(t.BackendPrefix),
		bundlevault.WithSecretMetadata(t.PushMetadata),
		bundlevault.WithVaultMetadata(t.AsVaultMetadata),
		bundlevault.WithMaxWorkerCount(t.MaxWorkerCount),
	); err != nil {
		return fmt.Errorf("error occurs during vault export (prefix: '%s'): %w", t.BackendPrefix, err)
	}

	// No error
	return nil
}
