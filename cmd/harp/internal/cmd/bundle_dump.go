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
type bundleDumpParams struct {
	inputPath      string
	dataOnly       bool
	metadataOnly   bool
	pathOnly       bool
	jmesPathFilter string
	skipTemplate   bool
}

var bundleDumpCmd = func() *cobra.Command {
	params := &bundleDumpParams{}

	cmd := &cobra.Command{
		Use:   "dump",
		Short: "Dump as JSON",
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-bundle-dump", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Prepare task
			t := &bundle.DumpTask{
				ContainerReader: cmdutil.FileReader(params.inputPath),
				OutputWriter:    cmdutil.StdoutWriter(),
				DataOnly:        params.dataOnly,
				MetadataOnly:    params.metadataOnly,
				PathOnly:        params.pathOnly,
				JMESPathFilter:  params.jmesPathFilter,
				IgnoreTemplate:  params.skipTemplate,
			}

			// Run the task
			if err := t.Run(ctx); err != nil {
				log.For(ctx).Fatal("unable to execute task", zap.Error(err))
			}
		},
	}

	// Parameters
	cmd.Flags().StringVar(&params.inputPath, "in", "", "Container input ('-' for stdin or filename)")
	cmd.Flags().BoolVar(&params.dataOnly, "content-only", false, "Display content only (data-only alias)")
	cmd.Flags().BoolVar(&params.dataOnly, "data-only", false, "Display data only")
	cmd.Flags().BoolVar(&params.metadataOnly, "metadata-only", false, "Display metadata only")
	cmd.Flags().BoolVar(&params.pathOnly, "path-only", false, "Display path only")
	cmd.Flags().StringVar(&params.jmesPathFilter, "query", "", "Specify a JMESPath query to format output")
	cmd.Flags().BoolVar(&params.skipTemplate, "skip-template", false, "Drop template from dump")

	return cmd
}
