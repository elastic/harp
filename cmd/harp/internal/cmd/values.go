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
	"encoding/json"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
	tplcmdutil "github.com/elastic/harp/pkg/template/cmdutil"
)

// -----------------------------------------------------------------------------

var valuesCmd = func() *cobra.Command {
	var (
		outputPath   string
		valueFiles   []string
		values       []string
		stringValues []string
		fileValues   []string
	)

	cmd := &cobra.Command{
		Use:     "values",
		Aliases: []string{"v"},
		Short:   "Template value preprocessor",
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-values", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
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

			// Create output writer
			writer, err := cmdutil.Writer(outputPath)
			if err != nil {
				log.For(ctx).Fatal("unable to create output writer", zap.Error(err), zap.String("path", outputPath))
			}

			// Write rendered content
			if err := json.NewEncoder(writer).Encode(values); err != nil {
				log.For(ctx).Fatal("unable to dump values as JSON", zap.Error(err), zap.String("path", outputPath))
			}
		},
	}

	// Parameters
	cmd.Flags().StringVar(&outputPath, "out", "", "Output file ('-' for stdout or a filename)")
	cmd.Flags().StringArrayVarP(&valueFiles, "values", "f", []string{}, "Specifies value files to load. Use <path>:<type>[:<prefix>] to override type detection (json,yaml,xml,hocon,toml)")
	cmd.Flags().StringArrayVar(&values, "set", []string{}, "Specifies value (k=v)")
	cmd.Flags().StringArrayVar(&stringValues, "set-string", []string{}, "Specifies value (k=string)")
	cmd.Flags().StringArrayVar(&fileValues, "set-file", []string{}, "Specifies value (k=filepath)")

	return cmd
}
