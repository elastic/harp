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
	"github.com/elastic/harp/pkg/tasks/crate"
)

// -----------------------------------------------------------------------------

type cratePushParams struct {
	inputPath  string
	outputPath string
	to         string
	path       string
	json       bool
}

var cratePushCmd = func() *cobra.Command {
	params := &cratePushParams{}

	longDesc := cmdutil.LongDesc(`
	Export a sealed container to an OCI compatible registry.

	The container will be pushed as 'harp.sealed' file inside a dedicated OCI layer.
	This will allow you to reuse general purpose OCI registry to push and pull
	prepared secret containers.`)

	examples := cmdutil.Examples(`
	# Push a container from STDIN to github container registry
	harp crate push --to registry:ghcr.io/elastic/harp --ref region-boostrap:v1

	# Push a container from a file to Google container registry
	harp crate push --to region.gcr.io --ref YOUR_GCP_PROJECT_ID/project-secrets:latest`)
	cmd := &cobra.Command{
		Use:     "push",
		Short:   "Push a crate",
		Long:    longDesc,
		Example: examples,
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-crate-push", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Prepare task
			t := &crate.PushTask{
				ContainerReader: cmdutil.FileReader(params.inputPath),
				OutputWriter:    cmdutil.FileWriter(params.outputPath),
				Target:          params.to,
				Path:            params.path,
				JSONOutput:      params.json,
			}

			// Run the task
			if err := t.Run(ctx); err != nil {
				log.For(ctx).Fatal("unable to execute task", zap.Error(err))
			}
		},
	}

	// Parameters
	cmd.Flags().StringVar(&params.inputPath, "in", "-", "Container path ('-' for stdin or filename)")
	cmd.Flags().StringVar(&params.outputPath, "out", "-", "Output path ('-' for stdout or filename)")
	cmd.Flags().StringVar(&params.to, "to", "", "Target destination (registry:<url>, oci:<path>, files:<path>)")
	cmd.Flags().StringVar(&params.path, "path", "harp.sealed", "Container path")
	cmd.Flags().BoolVar(&params.json, "json", false, "Enable JSON output")

	return cmd
}
