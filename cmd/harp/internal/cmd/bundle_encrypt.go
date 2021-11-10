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
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/sdk/value"
	"github.com/elastic/harp/pkg/sdk/value/encryption"
	"github.com/elastic/harp/pkg/tasks/bundle"
)

// -----------------------------------------------------------------------------
type bundleEncryptParams struct {
	inputPath      string
	outputPath     string
	key            string
	keyAliases     []string
	skipUnresolved bool
}

var bundleEncryptCmd = func() *cobra.Command {
	params := &bundleEncryptParams{}

	cmd := &cobra.Command{
		Use:   "encrypt",
		Short: "Encrypt secret values",
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-bundle-encrypt", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Prepare task
			t := &bundle.EncryptTask{
				ContainerReader: cmdutil.FileReader(params.inputPath),
				OutputWriter:    cmdutil.FileWriter(params.outputPath),
			}
			switch {
			case params.key != "":
				// Create transformer according to used encryption key
				transformer, err := encryption.FromKey(params.key)
				if err != nil {
					log.For(ctx).Fatal("unable to initialize transformer", zap.Error(err))
				}

				// Use the given key a bundle transformer
				t.BundleTransformer = transformer
			case len(params.keyAliases) > 0:
				transformerMap := map[string]value.Transformer{}

				// Split all alias / key
				for _, alias := range params.keyAliases {
					// Split alias
					parts := strings.SplitN(alias, ":", 2)
					if len(parts) != 2 {
						log.For(ctx).Fatal("invalid alias, it must be formatted alias:key.", zap.String("alias", alias))
						return
					}

					// Create transformer according to used encryption key
					transformer, err := encryption.FromKey(parts[1])
					if err != nil {
						log.For(ctx).Fatal("unable to initialize transformer", zap.Error(err))
					}

					// Assign to map
					transformerMap[parts[0]] = transformer
				}

				// Use transformer map
				t.TransformerMap = transformerMap
				t.SkipUnresolved = params.skipUnresolved
			default:
				log.For(ctx).Fatal("--key or --key-alias must be provided")
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
	cmd.Flags().StringVar(&params.key, "key", "", "Secret value encryption key for full bundle encryption")
	cmd.Flags().StringSliceVar(&params.keyAliases, "key-alias", []string{}, "Secret value encryption key for partial bundle encryption ('alias:key')")
	cmd.Flags().BoolVarP(&params.skipUnresolved, "skip-unresolved-key-alias", "s", false, "Skip unresolved key alias during partial bundle encryption")

	return cmd
}
