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
	"github.com/elastic/harp/pkg/sdk/value/encryption"
	"github.com/elastic/harp/pkg/tasks/bundle"
)

// -----------------------------------------------------------------------------

var bundleDecryptCmd = func() *cobra.Command {
	var (
		inputPath  string
		outputPath string
		key        string
	)

	cmd := &cobra.Command{
		Use:   "decrypt",
		Short: "Decrypt secret values",
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-bundle-decrypt", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Create transformer according to used encryption key
			transformer, err := encryption.FromKey(key)
			if err != nil {
				log.For(ctx).Fatal("unable to initialize transformer", zap.Error(err))
			}

			// Prepare task
			t := &bundle.DecryptTask{
				ContainerReader: cmdutil.FileReader(inputPath),
				OutputWriter:    cmdutil.FileWriter(outputPath),
				Transformer:     transformer,
			}

			// Run the task
			if err := t.Run(ctx); err != nil {
				log.For(ctx).Fatal("unable to execute task", zap.Error(err))
			}
		},
	}

	// Parameters
	cmd.Flags().StringVar(&inputPath, "in", "", "Container input ('-' for stdin or filename)")
	cmd.Flags().StringVar(&outputPath, "out", "", "Container output ('-' for stdout or filename)")
	cmd.Flags().StringVar(&key, "key", "", "Secret value decryption key")
	log.CheckErr("unable to mark 'key' flag as required.", cmd.MarkFlagRequired("key"))

	return cmd
}
