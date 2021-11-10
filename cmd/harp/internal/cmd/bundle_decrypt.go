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

	cmd := &cobra.Command{
		Use:   "decrypt",
		Short: "Decrypt secret values",
		Run: func(cmd *cobra.Command, args []string) {
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
