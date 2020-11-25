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
	tplcmdutil "github.com/elastic/harp/pkg/template/cmdutil"
)

// -----------------------------------------------------------------------------

var bundlePatchCmd = func() *cobra.Command {
	var (
		inputPath    string
		outputPath   string
		patchPath    string
		valueFiles   []string
		values       []string
		stringValues []string
		fileValues   []string
	)

	cmd := &cobra.Command{
		Use:   "patch",
		Short: "Apply patch to the given bundle",
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-bundle-patch", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Load values
			valueOpts := tplcmdutil.ValueOptions{
				ValueFiles:   valueFiles,
				Values:       values,
				StringValues: stringValues,
				FileValues:   fileValues,
			}
			values, err := valueOpts.MergeValues()
			if err != nil {
				log.For(ctx).Fatal("unable to process values", zap.Error(err))
			}

			// Prepare task
			t := &bundle.PatchTask{
				ContainerReader: cmdutil.FileReader(inputPath),
				PatchReader:     cmdutil.FileReader(patchPath),
				OutputWriter:    cmdutil.FileWriter(outputPath),
				Values:          values,
			}

			// Run the task
			if err := t.Run(ctx); err != nil {
				log.For(ctx).Fatal("unable to execute task", zap.Error(err))
			}
		},
	}

	// Parameters
	cmd.Flags().StringVar(&inputPath, "in", "-", "Container input ('-' for stdin or filename)")
	cmd.Flags().StringVar(&outputPath, "out", "", "Container output ('-' for stdout or a filename)")
	cmd.Flags().StringVar(&patchPath, "spec", "", "Patch specification path ('-' for stdin or filename)")
	log.CheckErr("unable to mark 'spec' flag as required.", cmd.MarkFlagRequired("spec"))
	cmd.Flags().StringArrayVar(&valueFiles, "values", []string{}, "Specifies value files to load")
	cmd.Flags().StringArrayVar(&values, "set", []string{}, "Specifies value (k=v)")
	cmd.Flags().StringArrayVar(&stringValues, "set-string", []string{}, "Specifies value (k=string)")
	cmd.Flags().StringArrayVar(&fileValues, "set-file", []string{}, "Specifies value (k=filepath)")

	return cmd
}
