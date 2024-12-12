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

	"github.com/elastic/harp/pkg/bundle/patch"
	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/tasks/bundle"
	tplcmdutil "github.com/elastic/harp/pkg/template/cmdutil"
)

// -----------------------------------------------------------------------------
type bundlePatchParams struct {
	inputPath         string
	outputPath        string
	patchPath         string
	valueFiles        []string
	values            []string
	stringValues      []string
	fileValues        []string
	stopAtRuleIndex   int
	stopAtRuleID      string
	ignoreRuleIDs     []string
	ignoreRuleIndexes []int
}

var bundlePatchCmd = func() *cobra.Command {
	params := &bundlePatchParams{}

	cmd := &cobra.Command{
		Use:   "patch",
		Short: "Apply patch to the given bundle",
		Run: func(cmd *cobra.Command, _ []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-bundle-patch", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Load values
			valueOpts := tplcmdutil.ValueOptions{
				ValueFiles:   params.valueFiles,
				Values:       params.values,
				StringValues: params.stringValues,
				FileValues:   params.fileValues,
			}
			values, err := valueOpts.MergeValues()
			if err != nil {
				log.For(ctx).Fatal("unable to process values", zap.Error(err))
			}

			// Prepare patch options.
			opts := []patch.OptionFunc{
				patch.WithStopAtRuleID(params.stopAtRuleID),
				patch.WithStopAtRuleIndex(params.stopAtRuleIndex),
				patch.WithIgnoreRuleIDs(params.ignoreRuleIDs...),
				patch.WithIgnoreRuleIndexes(params.ignoreRuleIndexes...),
			}

			// Prepare task
			t := &bundle.PatchTask{
				ContainerReader: cmdutil.FileReader(params.inputPath),
				PatchReader:     cmdutil.FileReader(params.patchPath),
				OutputWriter:    cmdutil.FileWriter(params.outputPath),
				Values:          values,
				Options:         opts,
			}

			// Run the task
			if err := t.Run(ctx); err != nil {
				log.For(ctx).Fatal("unable to execute task", zap.Error(err))
			}
		},
	}

	// Parameters
	cmd.Flags().StringVar(&params.inputPath, "in", "-", "Container input ('-' for stdin or filename)")
	cmd.Flags().StringVar(&params.outputPath, "out", "", "Container output ('-' for stdout or a filename)")
	cmd.Flags().StringVar(&params.patchPath, "spec", "", "Patch specification path ('-' for stdin or filename)")
	log.CheckErr("unable to mark 'spec' flag as required.", cmd.MarkFlagRequired("spec"))
	cmd.Flags().StringArrayVar(&params.valueFiles, "values", []string{}, "Specifies value files to load")
	cmd.Flags().StringArrayVar(&params.values, "set", []string{}, "Specifies value (k=v)")
	cmd.Flags().StringArrayVar(&params.stringValues, "set-string", []string{}, "Specifies value (k=string)")
	cmd.Flags().StringArrayVar(&params.fileValues, "set-file", []string{}, "Specifies value (k=filepath)")
	cmd.Flags().StringVar(&params.stopAtRuleID, "stop-at-rule-id", "", "Stop patch evaluation before the given rule ID")
	cmd.Flags().IntVar(&params.stopAtRuleIndex, "stop-at-rule-index", -1, "Stop patch evaluation before the given rule index (0 for first rule)")
	cmd.Flags().StringArrayVar(&params.ignoreRuleIDs, "ignore-rule-id", []string{}, "List of Rule identifier to ignore during evaluation")
	cmd.Flags().IntSliceVar(&params.ignoreRuleIndexes, "ignore-rule-index", []int{}, "List of Rule index to ignore during evaluation")

	return cmd
}
