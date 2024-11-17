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

package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/tasks/to"
)

// -----------------------------------------------------------------------------

var toVaultCmd = func() *cobra.Command {
	var (
		inputPath         string
		backendPrefix     string
		namespace         string
		withMetadata      bool
		withVaultMetadata bool
		maxWorkerCount    int64
	)

	cmd := &cobra.Command{
		Use:   "vault",
		Short: "Push a secret container in Hashicorp Vault",
		Run: func(cmd *cobra.Command, _ []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-to-vault", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Prepare task
			t := &to.VaultTask{
				ContainerReader: cmdutil.FileReader(inputPath),
				BackendPrefix:   backendPrefix,
				PushMetadata:    withMetadata || withVaultMetadata,
				AsVaultMetadata: withVaultMetadata,
				VaultNamespace:  namespace,
				MaxWorkerCount:  maxWorkerCount,
			}

			// Run the task
			if err := t.Run(ctx); err != nil {
				log.For(ctx).Fatal("unable to execute task", zap.Error(err))
			}
		},
	}

	// Parameters
	cmd.Flags().StringVar(&inputPath, "in", "-", "Container path ('-' for stdin or filename)")
	cmd.Flags().StringVar(&backendPrefix, "prefix", "", "Vault backend prefix")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Vault namespace")
	cmd.Flags().BoolVar(&withMetadata, "with-metadata", false, "Push container metadata as secret data")
	cmd.Flags().BoolVar(&withVaultMetadata, "with-vault-metadata", false, "Push container metadata as secret metadata (requires Vault >=1.9)")
	cmd.Flags().Int64Var(&maxWorkerCount, "worker-count", 4, "Active worker count limit")

	return cmd
}
