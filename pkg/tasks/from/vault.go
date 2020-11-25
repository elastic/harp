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

package from

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"

	"github.com/elastic/harp/pkg/bundle"
	bundlevault "github.com/elastic/harp/pkg/bundle/vault"
	"github.com/elastic/harp/pkg/tasks"
)

// VaultTask implements secret-container building from Vault K/V.
type VaultTask struct {
	OutputWriter   tasks.WriterProvider
	SecretPaths    []string
	VaultNamespace string
	WithMetadata   bool
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

	// Call exporter
	b, err := bundlevault.Pull(ctx, client, t.SecretPaths,
		bundlevault.WithMetadata(t.WithMetadata),
	)
	if err != nil {
		return fmt.Errorf("error occurs during vault export: %w", err)
	}

	// Create output writer
	writer, err := t.OutputWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to open output bundle: %w", err)
	}

	// Dump bundle
	if err = bundle.ToContainerWriter(writer, b); err != nil {
		return fmt.Errorf("unable to produce exported bundle: %w", err)
	}

	// No error
	return nil
}
