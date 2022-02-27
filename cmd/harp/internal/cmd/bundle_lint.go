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
type bundleLintParams struct {
	inputPath string
	specPath  string
}

var bundleLintCmd = func() *cobra.Command {
	params := &bundleLintParams{}

	longDesc := cmdutil.LongDesc(`
	Apply a RuleSet specification to the given bundle.

	This command is used to check a Bundle structure (Package => Secrets).
	A control gate could be implemented with this command to enforce a bundle
	structure by decoupling the bundle content and the usage contract.`)

	examples := cmdutil.Examples(`
	# Lint a bundle from STDIN
	harp bundle lint --spec cso.yaml`)

	cmd := &cobra.Command{
		Use:     "lint",
		Short:   "Lint the bundle using the given RuleSet spec",
		Long:    longDesc,
		Example: examples,
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-bundle-lint", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Prepare task
			t := &bundle.LintTask{
				ContainerReader: cmdutil.FileReader(params.inputPath),
				RuleSetReader:   cmdutil.FileReader(params.specPath),
			}

			// Run the task
			if err := t.Run(ctx); err != nil {
				log.For(ctx).Fatal("unable to execute task", zap.Error(err))
			}
		},
	}

	// Parameters
	cmd.Flags().StringVar(&params.inputPath, "in", "-", "Container input ('-' for stdin or filename)")
	cmd.Flags().StringVar(&params.specPath, "spec", "", "RuleSet specification path ('-' for stdin or filename)")
	log.CheckErr("unable to mark 'spec' flag as required.", cmd.MarkFlagRequired("spec"))

	return cmd
}
