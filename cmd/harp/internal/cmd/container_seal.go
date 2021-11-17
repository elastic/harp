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
	"encoding/base64"

	"github.com/awnumar/memguard"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/container/identity"
	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/tasks/container"
)

// -----------------------------------------------------------------------------

type containerSealParams struct {
	identities          []string
	identityFilePaths   []string
	inputPath           string
	outputPath          string
	masterKey           string
	target              string
	noContainerIdentity bool
	jsonOutput          bool
}

var containerSealCmd = func() *cobra.Command {
	params := containerSealParams{}

	cmd := &cobra.Command{
		Use:   "seal",
		Short: "Seal a secret container",
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-container-seal", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Load idenity from files
			if len(params.identityFilePaths) > 0 {
				for _, f := range params.identityFilePaths {
					if f == "" {
						// Ignore empty
						continue
					}

					// Open for reading
					r, err := cmdutil.Reader(f)
					if err != nil {
						log.For(ctx).Fatal("unable to read identity file", zap.Error(err), zap.String("identity", f))
					}

					// Decode identity
					id, err := identity.FromReader(r)
					if err != nil {
						log.For(ctx).Fatal("unable to decode identity from file", zap.Error(err), zap.String("identity", f))
					}

					// Append to identity list
					params.identities = append(params.identities, id.Public)
				}
			}

			// Convert identities to sealing keys
			peerPublicKeys, err := identity.SealingKeys(params.identities...)
			if err != nil {
				log.For(ctx).Fatal("unable to transform identity to a sealing key", zap.Error(err))
				return
			}

			// Prepare task
			t := &container.SealTask{
				ContainerReader:          cmdutil.FileReader(params.inputPath),
				SealedContainerWriter:    cmdutil.FileWriter(params.outputPath),
				OutputWriter:             cmdutil.StdoutWriter(),
				JSONOutput:               params.jsonOutput,
				PeerPublicKeys:           peerPublicKeys,
				DisableContainerIdentity: params.noContainerIdentity,
			}

			// Check container sealing master key usage
			if params.masterKey != "" {
				// Process target
				if params.target == "" {
					log.For(ctx).Fatal("target flag (string) is mandatory for key derivation")
				}

				// Assign target parameter
				t.DCKDTarget = params.target

				// Decode master key
				masterKeyRaw, err := base64.RawURLEncoding.DecodeString(params.masterKey)
				if err != nil {
					log.For(ctx).Fatal("unable to decode master key", zap.Error(err))
				}

				// Check appropriate lengh
				if len(masterKeyRaw) != 32 {
					log.For(ctx).Fatal("invalid master key length, it should be 32 bytes after decoding")
				}

				// Assign as seed
				t.DCKDMasterKey = memguard.NewBufferFromBytes(masterKeyRaw)
			}

			// Run the task
			if err := t.Run(ctx); err != nil {
				log.For(ctx).Fatal("unable to execute task", zap.Error(err))
			}
		},
	}

	// Parameters
	cmd.Flags().StringVar(&params.inputPath, "in", "", "Unsealed container input ('-' for stdin or filename)")
	cmd.Flags().StringVar(&params.outputPath, "out", "", "Sealed container output ('-' for stdout or filename)")
	log.CheckErr("unable to mark 'out' flag as required.", cmd.MarkFlagRequired("out"))
	cmd.Flags().BoolVar(&params.jsonOutput, "json", false, "Display seal info as json")
	cmd.Flags().StringArrayVar(&params.identities, "identity", []string{}, "Identity allowed to unseal")
	cmd.Flags().StringArrayVar(&params.identityFilePaths, "identity-file", []string{}, "Files with identity allowed to unseal")
	cmd.Flags().BoolVar(&params.noContainerIdentity, "no-container-identity", false, "Disable container identity")
	cmd.Flags().StringVar(&params.masterKey, "dckd-master-key", "", "Master key used for deterministic container key derivation")
	cmd.Flags().StringVar(&params.target, "dckd-target", "", "Target parameter for deterministic container key derivation")

	return cmd
}
