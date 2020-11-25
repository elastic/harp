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
	"github.com/awnumar/memguard"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/tasks/container"
)

// -----------------------------------------------------------------------------
type containerRecoveryParams struct {
	identityPath     string
	passPhrase       string
	jsonOutput       bool
	vaultTransitPath string
	vaultTransitKey  string
}

var containerRecoveryCmd = func() *cobra.Command {
	params := containerRecoveryParams{}

	cmd := &cobra.Command{
		Use:   "recover",
		Short: "Recover container key from identity",
		Run: func(cmd *cobra.Command, _ []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-container-recover", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Check mandatory flags
			if params.passPhrase == "" && params.vaultTransitKey == "" {
				log.For(ctx).Fatal("passphrase or vault-transit-path flag must be defined")
			}
			if params.passPhrase != "" && params.vaultTransitKey != "" {
				log.For(ctx).Fatal("passphrase and vault-transit-path flags are mutually exclusive")
			}

			// Prepare task
			t := &container.RecoverTask{
				JSONReader:       cmdutil.FileReader(params.identityPath),
				OutputWriter:     cmdutil.StdoutWriter(),
				PassPhrase:       memguard.NewBufferFromBytes([]byte(params.passPhrase)),
				VaultTransitPath: params.vaultTransitPath,
				VaultTransitKey:  params.vaultTransitKey,
				JSONOutput:       params.jsonOutput,
			}

			// Run the task
			if err := t.Run(ctx); err != nil {
				log.For(ctx).Fatal("unable to execute task", zap.Error(err))
			}
		},
	}

	// Flags
	cmd.Flags().StringVar(&params.identityPath, "identity", "", "Identity input  ('-' for stdout or filename)")
	cmd.Flags().StringVar(&params.passPhrase, "passphrase", "", "Identity private key passphrase")
	cmd.Flags().StringVar(&params.vaultTransitPath, "vault-transit-path", "transit", "Vault transit backend mount path")
	cmd.Flags().StringVar(&params.vaultTransitKey, "vault-transit-key", "", "Use Vault transit encryption to protect identity private key")
	cmd.Flags().BoolVar(&params.jsonOutput, "json", false, "Display container key as json")

	return cmd
}
