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
type bundleDiffParams struct {
	sourcePath      string
	destinationPath string
	generatePatch   bool
}

var bundleDiffCmd = func() *cobra.Command {
	params := &bundleDiffParams{}

	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Display container differences",
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-bundle-diff", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Prepare task
			t := &bundle.DiffTask{
				SourceReader:      cmdutil.FileReader(params.sourcePath),
				DestinationReader: cmdutil.FileReader(params.destinationPath),
				OutputWriter:      cmdutil.StdoutWriter(),
				GeneratePatch:     params.generatePatch,
			}

			// Run the task
			if err := t.Run(ctx); err != nil {
				log.For(ctx).Fatal("unable to execute task", zap.Error(err))
			}
		},
	}

	// Parameters
	cmd.Flags().StringVar(&params.sourcePath, "old", "", "Container path ('-' for stdin or filename)")
	cmd.Flags().StringVar(&params.destinationPath, "new", "", "Container path ('-' for stdin or filename)")
	log.CheckErr("unable to mark 'dst' flag as required.", cmd.MarkFlagRequired("dst"))
	cmd.Flags().BoolVar(&params.generatePatch, "patch", false, "Output as a bundle patch")

	return cmd
}
