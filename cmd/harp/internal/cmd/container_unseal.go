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
type containerUnsealParams struct {
	inputPath       string
	outputPath      string
	containerKeyRaw string
}

var containerUnsealCmd = func() *cobra.Command {
	params := containerUnsealParams{}

	cmd := &cobra.Command{
		Use:   "unseal",
		Short: "Unseal a secret container",
		Run: func(cmd *cobra.Command, _ []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-container-unseal", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Prepare passphrase
			containerKey := memguard.NewBufferFromBytes([]byte(params.containerKeyRaw))
			if params.containerKeyRaw == "" {
				var err error
				// Read passphrase from stdin
				containerKey, err = cmdutil.ReadSecret("Enter container key", false)
				if err != nil {
					log.For(ctx).Fatal("unable to read passphrase", zap.Error(err))
				}
			}
			defer containerKey.Destroy()

			// Prepare task
			t := &container.UnsealTask{
				ContainerReader: cmdutil.FileReader(params.inputPath),
				OutputWriter:    cmdutil.StdoutWriter(),
				ContainerKey:    containerKey,
			}

			// Run the task
			if err := t.Run(ctx); err != nil {
				log.For(ctx).Fatal("unable to execute task", zap.Error(err))
			}
		},
	}

	// Parameters
	cmd.Flags().StringVar(&params.inputPath, "in", "", "Sealed container input ('-' for stdin or filename)")
	cmd.Flags().StringVar(&params.outputPath, "out", "", "Unsealed container output ('-' for stdout or filename)")
	cmd.Flags().StringVar(&params.containerKeyRaw, "key", "", "Container key")
	log.CheckErr("unable to mark 'key' flag as required.", cmd.MarkFlagRequired("key"))

	return cmd
}
