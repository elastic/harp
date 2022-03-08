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
	"github.com/elastic/harp/pkg/tasks/lint"
)

// -----------------------------------------------------------------------------

type lintParams struct {
	inputPath  string
	outputPath string
	schema     string
	schemaOnly bool
}

var lintCmd = func() *cobra.Command {
	params := &lintParams{}

	longDesc := cmdutil.LongDesc(`
		Validate input YAML/JSON content with the selected JSONSchema definition.
	`)
	examples := cmdutil.Examples(`
	# Validate a JSON dump with schema detection from STDIN
	harp lint

	# Validate a BundleTemplate from a file
	harp lint --schema BundleTemplate --in template.yaml

	# Validate a RuleSet
	harp lint --schema RuleSet --in ruleset.yaml

	# Validate a BundlePatch
	harp lint --schema BundlePatch --in patch.yaml

	# Display a schema definition
	harp lint --schema Bundle --schema-only`)

	cmd := &cobra.Command{
		Use:     "lint",
		Short:   "Configuration linter commands",
		Long:    longDesc,
		Example: examples,
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-lint", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Prepare task
			t := &lint.ValidateTask{
				SourceReader: cmdutil.FileReader(params.inputPath),
				OutputWriter: cmdutil.FileWriter(params.outputPath),
				Schema:       params.schema,
				SchemaOnly:   params.schemaOnly,
			}

			// Run the task
			if err := t.Run(ctx); err != nil {
				log.For(ctx).Fatal("unable to execute task", zap.Error(err))
			}
		},
	}

	// Parameters
	cmd.Flags().StringVar(&params.inputPath, "in", "-", "Container input ('-' for stdin or filename)")
	cmd.Flags().StringVar(&params.outputPath, "out", "", "Container output ('' for stdout or filename)")
	cmd.Flags().StringVar(&params.schema, "schema", "", "Override schema detection for validation (Bundle|BundleTemplate|RuleSet|BundlePatch")
	cmd.Flags().BoolVar(&params.schemaOnly, "schema-only", false, "Display the JSON Schema")

	return cmd
}
