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
	"github.com/elastic/harp/pkg/sdk/value"
	"github.com/elastic/harp/pkg/sdk/value/encryption"
	"github.com/elastic/harp/pkg/tasks/bundle"
)

// -----------------------------------------------------------------------------
type bundleDecryptParams struct {
	inputPath          string
	outputPath         string
	keys               []string
	skipNotDecryptable bool
}

var bundleDecryptCmd = func() *cobra.Command {
	params := &bundleDecryptParams{}

	longDesc := cmdutil.LongDesc(`
	Decrypt a bundle content.

	For confidentiality purpose, bundle package value can be encrypted before
	the container sealing. It offers confidentiality properties so that the
	final consumer must know an additional decryption key to be allowed to
	read the package value.

	All package properties (name, labels, annotations) remain a clear-text
	message. Only package values (secret K/V) is encrypted.

	In order to decrypt the package value, harp uses the value encryption
	transformers. The required key must be provided in a format understandable
	by the encryption transformer factory.

	This act as in-transit/in-use encryption.
	`)

	examples := cmdutil.Examples(`
	# Decrypt a bundle from STDIN and produce output to STDOUT
	harp bundle decrypt --key <transformer key>

	# Decrypt a bundle from STDIN using multiple transformer keys
	harp bundle decrypt --key <transformer key 1> --key <transformer key 2>

	# Decrypt a bundle from STDIN and ignore secrets which could not be decrypted
	# with given transformer key (partial decryption / authorization by key)
	harp bundle decrypt --skip-not-decryptable --key <transformer-key>

	# Decrypt a bundle from STDIN and produce output to a file
	harp bundle decrypt --key <transformer key> --out decrypted.bundle
	`)

	cmd := &cobra.Command{
		Use:     "decrypt",
		Short:   "Decrypt secret values",
		Long:    longDesc,
		Example: examples,
		Run: func(cmd *cobra.Command, _ []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-bundle-decrypt", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Prepare transformer collection
			transformers := []value.Transformer{}

			// Split all alias / key
			for _, keyRaw := range params.keys {
				// Create transformer according to used encryption key
				transformer, err := encryption.FromKey(keyRaw)
				if err != nil {
					log.For(ctx).Fatal("unable to initialize transformer", zap.String("key", keyRaw), zap.Error(err))
					return
				}

				// Append to collection
				transformers = append(transformers, transformer)
			}

			// Prepare task
			t := &bundle.DecryptTask{
				ContainerReader:    cmdutil.FileReader(params.inputPath),
				OutputWriter:       cmdutil.FileWriter(params.outputPath),
				Transformers:       transformers,
				SkipNotDecryptable: params.skipNotDecryptable,
			}

			// Run the task
			if err := t.Run(ctx); err != nil {
				log.For(ctx).Fatal("unable to execute task", zap.Error(err))
			}
		},
	}

	// Parameters
	cmd.Flags().StringVar(&params.inputPath, "in", "", "Container input ('-' for stdin or filename)")
	cmd.Flags().StringVar(&params.outputPath, "out", "", "Container output ('-' for stdout or filename)")
	cmd.Flags().StringSliceVar(&params.keys, "key", []string{""}, "Secret value decryption key. Repeat to add multiple keys to try.")
	cmd.Flags().BoolVarP(&params.skipNotDecryptable, "skip-not-decryptable", "s", false, "Skip not decryptable secrets without raising an error.")

	return cmd
}
