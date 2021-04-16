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

package share

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"

	"github.com/elastic/harp/pkg/tasks"
	"github.com/elastic/harp/pkg/vault"
)

// GetTask implements secret sharing via Vault Cubbyhole.
type GetTask struct {
	OutputWriter   tasks.WriterProvider
	BackendPrefix  string
	VaultNamespace string
	Token          string
}

// Run the task.
func (t *GetTask) Run(ctx context.Context) error {
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

	// Create cubbyhole service
	sf, errFactory := vault.FromVaultClient(client)
	if err != nil {
		return fmt.Errorf("unable to initialize service factory: %w", errFactory)
	}
	s, errService := sf.Cubbyhole(t.BackendPrefix)
	if errService != nil {
		return fmt.Errorf("unable to initialize service factory: %w", errFactory)
	}

	// Create output writer
	writer, err := t.OutputWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to open output writer: %w", err)
	}

	// Retrieve secret
	if err := s.Get(ctx, t.Token, writer); err != nil {
		return fmt.Errorf("unable to retrieve secret: %w", err)
	}

	// No error
	return nil
}
