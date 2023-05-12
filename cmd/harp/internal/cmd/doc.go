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
	"github.com/spf13/cobra/doc"

	"github.com/elastic/harp/pkg/sdk/cmdutil"
)

// -----------------------------------------------------------------------------

var docCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doc",
		Short: "Generates documentation and autocompletion",
	}

	// Subcommands
	cmd.AddCommand(docMarkdownCmd())

	return cmd
}

// -----------------------------------------------------------------------------

var docDestination string

var docMarkdownCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "markdown",
		Aliases: []string{"md"},
		Short:   "Documentation in Markdown format",
		RunE:    runDocMarkdown,
	}

	// Parameters
	cmd.Flags().StringVarP(&docDestination, "destination", "d", "", "destination for documentation")

	return cmd
}

//nolint:revive // refactor use of args
func runDocMarkdown(cmd *cobra.Command, args []string) error {
	// Context to attach all goroutines
	_, cancel := cmdutil.Context(cmd.Context(), "harp-doc-markdown", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
	defer cancel()

	// Disable flag
	cmd.Root().DisableAutoGenTag = true

	// Generate markdown tree
	return doc.GenMarkdownTree(cmd.Root(), docDestination)
}
