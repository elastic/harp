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

	"github.com/elastic/harp/build/fips"
	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/sdk/value"
	"github.com/elastic/harp/pkg/sdk/value/encryption"
	"github.com/elastic/harp/pkg/sdk/value/encryption/jwe"
	"github.com/elastic/harp/pkg/tasks/container"
	"github.com/elastic/harp/pkg/vault"
)

// -----------------------------------------------------------------------------
type containerIdentityParams struct {
	outputPath       string
	description      string
	key              string
	passPhrase       string
	vaultTransitPath string
	vaultTransitKey  string
	version          uint
}

var containerIdentityCmd = func() *cobra.Command {
	params := containerIdentityParams{}

	cmd := &cobra.Command{
		Use:     "identity",
		Aliases: []string{"id"},
		Short:   "Generate container identity",
		Run: func(cmd *cobra.Command, _ []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-identity", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Prepare value transformer
			var (
				transformer    value.Transformer
				errTransformer error
			)
			switch {
			case params.key != "":
				transformer, errTransformer = encryption.FromKey(params.key)
			case params.passPhrase != "":
				transformer, errTransformer = jwe.Transformer(jwe.PBES2_HS512_A256KW, params.passPhrase)
			case params.vaultTransitKey != "" && params.vaultTransitPath != "":
				transformer, errTransformer = vault.Transformer(params.vaultTransitPath, params.vaultTransitKey, vault.Chacha20Poly1305)
			default:
				log.For(ctx).Fatal("unable to initialize value transformer, key or vault-transit-path or passphrase must be provided")
				return
			}
			if errTransformer != nil {
				log.For(ctx).Fatal("unable to initialize value transformer", zap.Error(errTransformer))
				return
			}

			// Prepare task
			t := &container.IdentityTask{
				OutputWriter: cmdutil.FileWriter(params.outputPath),
				Description:  params.description,
				Transformer:  transformer,
				Version:      container.IdentityVersion(params.version + 1),
			}

			// Run the task
			if err := t.Run(ctx); err != nil {
				log.For(ctx).Fatal("unable to execute task", zap.Error(err))
			}
		},
	}

	// Select default identity version.
	identityVersion := uint(container.ModernIdentity)
	if fips.Enabled() {
		identityVersion = uint(container.NISTIdentity)
	}

	// Flags
	cmd.Flags().StringVar(&params.outputPath, "out", "", "Identity information output ('-' for stdout or filename)")
	cmd.Flags().StringVar(&params.key, "key", "", "Transformer key")
	cmd.Flags().StringVar(&params.passPhrase, "passphrase", "", "Identity private key passphrase")
	cmd.Flags().StringVar(&params.vaultTransitPath, "vault-transit-path", "transit", "Vault transit backend mount path")
	cmd.Flags().StringVar(&params.vaultTransitKey, "vault-transit-key", "", "Use Vault transit encryption to protect identity private key")
	cmd.Flags().StringVar(&params.description, "description", "", "Identity description")
	log.CheckErr("unable to mark 'description' flag as required.", cmd.MarkFlagRequired("description"))
	cmd.Flags().UintVar(&params.version, "version", identityVersion-1, "Select identity version (0:legacy, 1:modern, 2:nist)")
	return cmd
}
