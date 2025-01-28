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
	"github.com/elastic/harp/pkg/tasks/template"
)

type renderParams struct {
	InputPath     string
	OutputPath    string
	ValueFiles    []string
	SecretLoaders []string
	Values        []string
	StringValues  []string
	FileValues    []string
	LeftDelims    string
	RightDelims   string
	AltDelims     bool
	RootPath      string
	DryRun        bool
}

// -----------------------------------------------------------------------------

var renderCmd = func() *cobra.Command {
	params := &renderParams{}

	longDesc := cmdutil.LongDesc(`
	Generate a config filesytem from a template hierarchy or archive.
	`)
	examples := cmdutil.Examples(`
	# Generate a configuration filesystem from a folder hierarchy
	harp render --in templates/database --out postgres

	# Generate a configuration filesystem from an archive
	harp render --in templates.tar.gz --out configMap

	# Test template generation
	harp render --in templates.tar.gz --dry-run
	`)

	cmd := &cobra.Command{
		Use:     "render",
		Aliases: []string{"r"},
		Short:   "Render a template filesystem",
		Long:    longDesc,
		Example: examples,
		Run: func(cmd *cobra.Command, _ []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "template-render", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Prepare task
			t := &template.FileSystemTask{
				InputPath:          params.InputPath,
				OutputPath:         params.OutputPath,
				ValueFiles:         params.ValueFiles,
				Values:             params.Values,
				StringValues:       params.StringValues,
				FileValues:         params.FileValues,
				FileLoaderRootPath: params.RootPath,
				SecretLoaders:      params.SecretLoaders,
				LeftDelims:         params.LeftDelims,
				RightDelims:        params.RightDelims,
				AltDelims:          params.AltDelims,
				DryRun:             params.DryRun,
			}

			// Run the task
			if err := t.Run(ctx); err != nil {
				log.For(ctx).Fatal("unable to execute task", zap.Error(err))
			}
		},
	}

	// Parameters
	cmd.Flags().StringVar(&params.InputPath, "in", "", "Template input path (directory or archive)")
	log.CheckErr("unable to mark 'in' flag as required.", cmd.MarkFlagRequired("in"))
	cmd.Flags().StringVar(&params.OutputPath, "out", "", "Output path")
	cmd.Flags().StringVar(&params.RootPath, "root", "", "Defines file loader root base path")
	cmd.Flags().StringArrayVarP(&params.SecretLoaders, "secrets-from", "s", []string{"vault"}, "Specifies secret containers to load ('vault' for Vault loader or '-' for stdin or filename)")
	cmd.Flags().StringArrayVarP(&params.ValueFiles, "values", "f", []string{}, "Specifies value files to load")
	cmd.Flags().StringArrayVar(&params.Values, "set", []string{}, "Specifies value (k=v)")
	cmd.Flags().StringArrayVar(&params.StringValues, "set-string", []string{}, "Specifies value (k=string)")
	cmd.Flags().StringArrayVar(&params.FileValues, "set-file", []string{}, "Specifies value (k=filepath)")
	cmd.Flags().StringVar(&params.LeftDelims, "left-delimiter", "{{", "Template left delimiter (default to '{{')")
	cmd.Flags().StringVar(&params.RightDelims, "right-delimiter", "}}", "Template right delimiter (default to '}}')")
	cmd.Flags().BoolVar(&params.AltDelims, "alt-delims", false, "Define '[[' and ']]' as template delimiters.")
	cmd.Flags().BoolVar(&params.DryRun, "dry-run", false, "Generate in-memory only.")

	return cmd
}
