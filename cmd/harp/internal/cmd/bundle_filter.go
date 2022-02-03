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
type bundleFilterParams struct {
	inputPath    string
	outputPath   string
	excludePaths []string
	keepPaths    []string
	jmesPath     string
	regoPolicy   string
	reverseLogic bool
}

var bundleFilterCmd = func() *cobra.Command {
	params := &bundleFilterParams{}

	cmd := &cobra.Command{
		Use:     "filter",
		Aliases: []string{"grep", "f"},
		Short:   "Filter package names",
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-bundle-filter", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Prepare task
			t := &bundle.FilterTask{
				ContainerReader: cmdutil.FileReader(params.inputPath),
				OutputWriter:    cmdutil.FileWriter(params.outputPath),
				ExcludePaths:    params.excludePaths,
				KeepPaths:       params.keepPaths,
				JMESPath:        params.jmesPath,
				RegoPolicy:      params.regoPolicy,
				ReverseLogic:    params.reverseLogic,
			}

			// Run the task
			if err := t.Run(ctx); err != nil {
				log.For(ctx).Fatal("unable to execute task", zap.Error(err))
			}
		},
	}

	// Parameters
	cmd.Flags().StringVar(&params.inputPath, "in", "", "Container input ('-' for stdin or filename)")
	cmd.Flags().StringVar(&params.outputPath, "out", "", "Container path ('-' for stdout or filename)")
	cmd.Flags().StringArrayVar(&params.excludePaths, "exclude", []string{}, "Exclude path")
	cmd.Flags().StringArrayVar(&params.keepPaths, "keep", []string{}, "Keep path")
	cmd.Flags().StringVar(&params.jmesPath, "query", "", "JMESPath query used as package filter")
	cmd.Flags().StringVar(&params.regoPolicy, "policy", "", "OPA Rego policy file as package filter")
	cmd.Flags().BoolVar(&params.reverseLogic, "not", false, "Reverse filter logic expression")

	return cmd
}
