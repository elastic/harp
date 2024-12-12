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
	"github.com/elastic/harp/pkg/tasks/bundle"
)

// -----------------------------------------------------------------------------

var bundleReadCmd = func() *cobra.Command {
	var (
		inputPath   string
		packageName string
		secretKey   string
	)

	cmd := &cobra.Command{
		Use:   "read",
		Short: "Read a secret from bundle",
		Run: func(cmd *cobra.Command, _ []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-bundle-read", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Prepare task
			t := &bundle.ReadTask{
				ContainerReader: cmdutil.FileReader(inputPath),
				OutputWriter:    cmdutil.StdoutWriter(),
				PackageName:     packageName,
				SecretKey:       secretKey,
			}

			// Run the task
			if err := t.Run(ctx); err != nil {
				log.For(ctx).Fatal("unable to execute task", zap.Error(err))
			}
		},
	}

	// Parameters
	cmd.Flags().StringVar(&inputPath, "in", "", "Container input ('-' for stdin or filename)")
	cmd.Flags().StringVar(&packageName, "path", "", "Secret path")
	log.CheckErr("unable to mark 'path' flag as required.", cmd.MarkFlagRequired("path"))
	cmd.Flags().StringVar(&secretKey, "field", "", "Secret field")

	return cmd
}
