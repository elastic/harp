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

type toOCIParams struct {
	inputPath  string
	outputPath string
	repository string
	path       string
	json       bool
}

var toOCICmd = func() *cobra.Command {
	params := &toOCIParams{}

	longDesc := cmdutil.LongDesc(`
	Export a sealed container to an OCI compatible registry.

	The container will be pushed as 'harp.sealed' file inside a dedicated OCI layer.
	This will allow you to reuse general purpose OCI registry to push and pull
	prepared secret containers.`)

	examples := cmdutil.Examples(`
	# Push a container from STDIN to github container registry
	harp to oci --repository ghcr.io/elastic/harp/production-secrets:v1

	# Push a container from a file to Google container registry
	harp to oci --repository region.gcr.io/YOUR_GCP_PROJECT_ID/project-secrets:latest`)
	cmd := &cobra.Command{
		Use:     "oci",
		Short:   "Push a sealed secret container in an OCI compliant registry",
		Long:    longDesc,
		Example: examples,
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-to-oci", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Prepare task
			t := &to.OCITask{
				ContainerReader: cmdutil.FileReader(params.inputPath),
				OutputWriter:    cmdutil.FileWriter(params.outputPath),
				Repository:      params.repository,
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
	cmd.Flags().StringVar(&params.repository, "repository", "", "Repository address")
	cmd.Flags().StringVar(&params.path, "path", "harp.sealed", "Container path")
	cmd.Flags().BoolVar(&params.json, "json", false, "Enable JSON output")

	return cmd
}
