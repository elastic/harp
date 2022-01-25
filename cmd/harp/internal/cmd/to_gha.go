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
	"github.com/elastic/harp/pkg/tasks/to"
)

// -----------------------------------------------------------------------------

type toGithubActionParams struct {
	inputPath    string
	owner        string
	repository   string
	secretFilter string
}

var toGithubActionCmd = func() *cobra.Command {
	var params toGithubActionParams

	cmd := &cobra.Command{
		Use:     "github-actions",
		Aliases: []string{"gha"},
		Short:   "Export all secrets to Github Actions as repository secrets.",
		Example: `$ export GITHUB_TOKEN=ghp_###############
$ harp to gha --in secret.container --owner elastic --owner harp --secret-filter "COSIGN_*"`,
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-to-gha", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Prepare task
			t := &to.GithubActionTask{
				ContainerReader: cmdutil.FileReader(params.inputPath),
				Owner:           params.owner,
				Repository:      params.repository,
				SecretFilter:    params.secretFilter,
			}

			// Run the task
			if err := t.Run(ctx); err != nil {
				log.For(ctx).Fatal("unable to execute task", zap.Error(err))
			}
		},
	}

	// Parameters
	cmd.Flags().StringVar(&params.inputPath, "in", "-", "Container path ('-' for stdin or filename)")
	cmd.Flags().StringVar(&params.owner, "owner", "", "Github owner/organization")
	cmd.Flags().StringVar(&params.repository, "repository", "", "Github repository")
	cmd.Flags().StringVar(&params.secretFilter, "secret-filter", "*", "Specify secret filter as Glob (*_KEY, private*)")

	return cmd
}
