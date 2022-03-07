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
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xeipuuv/gojsonschema"

	"github.com/elastic/harp/pkg/bundle"
	"github.com/elastic/harp/pkg/bundle/patch"
	"github.com/elastic/harp/pkg/bundle/ruleset"
	"github.com/elastic/harp/pkg/bundle/template"
	"github.com/elastic/harp/pkg/sdk/cmdutil"
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
	# Validate a Bundle JSON dump from STDIN
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
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-lint", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			if err := runLint(ctx, params); err != nil {
				os.Exit(-1)
			}
		},
	}

	// Parameters
	cmd.Flags().StringVar(&params.inputPath, "in", "-", "Container input ('-' for stdin or filename)")
	cmd.Flags().StringVar(&params.outputPath, "out", "", "Container output ('' for stdout or filename)")
	cmd.Flags().StringVar(&params.schema, "schema", "Bundle", "Schema to use for validation (Bundle|BundleTemplate|RuleSet|BundlePatch")
	cmd.Flags().BoolVar(&params.schemaOnly, "schema-only", false, "Display the JSON Schema")

	return cmd
}

func runLint(_ context.Context, params *lintParams) error {
	var (
		schemaDefinition []byte
		linterFunc       func(io.Reader) ([]gojsonschema.ResultError, error)
	)

	// Select lint strategy
	switch {
	case strings.EqualFold(params.schema, "Bundle"):
		schemaDefinition = bundle.JSONSchema()
		linterFunc = bundle.Lint
	case strings.EqualFold(params.schema, "BundleTemplate"):
		schemaDefinition = template.JSONSchema()
		linterFunc = template.Lint
	case strings.EqualFold(params.schema, "RuleSet"):
		schemaDefinition = ruleset.JSONSchema()
		linterFunc = ruleset.Lint
	case strings.EqualFold(params.schema, "BundlePatch"):
		schemaDefinition = patch.JSONSchema()
		linterFunc = patch.Lint
	default:
		return fmt.Errorf("given specification '%s' is not supported by the lint command", params.schema)
	}

	// Create output writer
	writer, err := cmdutil.Writer(params.outputPath)
	if err != nil {
		return fmt.Errorf("unable to open output bundle: %w", err)
	}

	// Display jsonschema
	if params.schemaOnly {
		fmt.Fprintln(writer, string(schemaDefinition))
		return nil
	}

	// Create input reader
	reader, err := cmdutil.Reader(params.inputPath)
	if err != nil {
		return fmt.Errorf("unable to read input reader: %w", err)
	}

	// Execute the lint evaluation
	validationErrors, err := linterFunc(reader)
	switch {
	case len(validationErrors) > 0:
		for _, e := range validationErrors {
			fmt.Fprintf(writer, " - %s\n", e.String())
		}
		return err
	case err != nil:
		return fmt.Errorf("unexpected validation error occurred: %w", err)
	default:
	}

	// No error
	return nil
}
