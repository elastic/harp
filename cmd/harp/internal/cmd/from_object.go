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
	"github.com/elastic/harp/pkg/tasks/from"
)

// -----------------------------------------------------------------------------

var fromObjectCmd = func() *cobra.Command {
	var (
		inputPath  string
		outputPath string
		inputType  string
	)
	cmd := &cobra.Command{
		Use:   "object",
		Short: "Convert an object to a secret-container",
		Run: func(cmd *cobra.Command, _ []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-from-object", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Prepare task
			t := &from.ObjectTask{
				ObjectReader: cmdutil.FileReader(inputPath),
				OutputWriter: cmdutil.FileWriter(outputPath),
				JSON:         inputType == "json",
				YAML:         inputType == "yaml",
			}

			// Run the task
			if err := t.Run(ctx); err != nil {
				log.For(ctx).Fatal("unable to execute task", zap.Error(err))
			}
		},
	}

	// Parameters
	cmd.Flags().StringVar(&inputPath, "in", "-", "YAML object ('-' for stdin or filename)")
	cmd.Flags().StringVar(&outputPath, "out", "-", "Container output ('-' for stdout or filename)")
	cmd.Flags().StringVar(&inputType, "format", "yaml", "Input file format (yaml or json)")

	return cmd
}
