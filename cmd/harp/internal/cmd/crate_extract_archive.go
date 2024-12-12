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
	"github.com/elastic/harp/pkg/tasks/crate"
)

// -----------------------------------------------------------------------------

type crateExtractArchiveParams struct {
	inputPath  string
	outputPath string
}

var createExtractArchiveCmd = func() *cobra.Command {
	params := &crateExtractArchiveParams{}

	longDesc := cmdutil.LongDesc(`
	Extract a template archive to the given path.`)

	examples := cmdutil.Examples(``)

	cmd := &cobra.Command{
		Use:     "extract-archive",
		Short:   "Extract a template archive",
		Long:    longDesc,
		Example: examples,
		Run: func(cmd *cobra.Command, _ []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-crate-extract-archive", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Prepare task
			t := &crate.ExtractArchiveTask{
				ArchiveReader: cmdutil.FileReader(params.inputPath),
				OutputPath:    params.outputPath,
			}

			// Run the task
			if err := t.Run(ctx); err != nil {
				log.For(ctx).Fatal("unable to execute task", zap.Error(err))
			}
		},
	}

	// Parameters
	cmd.Flags().StringVarP(&params.inputPath, "archive", "a", "", "Archive path ('-' for stdin or filename)")
	cmd.Flags().StringVar(&params.outputPath, "out", "-", "Output path ('-' for stdout or filename)")

	return cmd
}
