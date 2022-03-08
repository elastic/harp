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
	"oras.land/oras-go/pkg/content"

	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/tasks/crate"
)

// -----------------------------------------------------------------------------

type crateCopyParams struct {
	from     string
	fromRef  string
	fromOpts content.RegistryOptions
	to       string
	toRef    string
	toOpts   content.RegistryOptions
}

var crateCopyCmd = func() *cobra.Command {
	params := &crateCopyParams{}

	longDesc := cmdutil.LongDesc(`
	Copy a crate from one source to another.`)

	examples := cmdutil.Examples(`
	# Copy a crate from a registry to file for debugging purpose
	harp crate copy --from-ref <registry>/<image>:<tag> --to files:out`)

	cmd := &cobra.Command{
		Use:     "copy",
		Short:   "Copy a crate",
		Long:    longDesc,
		Example: examples,
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-crate-copy", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Prepare task
			t := &crate.CopyTask{
				Source:                  params.from,
				SourceRef:               params.fromRef,
				SourceRegistryOpts:      params.fromOpts,
				Destination:             params.to,
				DestinationRef:          params.toRef,
				DestinationRegistryOpts: params.toOpts,
			}

			// Run the task
			if err := t.Run(ctx); err != nil {
				log.For(ctx).Fatal("unable to execute task", zap.Error(err))
			}
		},
	}

	// Parameters
	cmd.Flags().StringVar(&params.from, "from", "registry", "Target destination (registry, oci:<path>, files:<path>)")
	cmd.Flags().StringVar(&params.fromRef, "from-ref", "", "Source image reference")
	cmd.Flags().StringVarP(&params.fromOpts.Username, "from-username", "u", "", "Registry username")
	cmd.Flags().StringVarP(&params.fromOpts.Password, "from-password", "p", "", "Registry password")
	cmd.Flags().BoolVarP(&params.fromOpts.Insecure, "from-insecure", "", false, "Allow connections to SSL registry without certs")
	cmd.Flags().BoolVarP(&params.fromOpts.PlainHTTP, "from-plain-http", "", false, "Use plain http and not https")

	cmd.Flags().StringVar(&params.to, "to", "file", "Target destination (registry, oci:<path>, files:<path>)")
	cmd.Flags().StringVar(&params.toRef, "to-ref", "", "Source image reference")
	cmd.Flags().StringVar(&params.toOpts.Username, "to-username", "", "Registry username")
	cmd.Flags().StringVar(&params.toOpts.Password, "to-password", "", "Registry password")
	cmd.Flags().BoolVarP(&params.toOpts.Insecure, "to-insecure", "", false, "Allow connections to SSL registry without certs")
	cmd.Flags().BoolVarP(&params.toOpts.PlainHTTP, "to-plain-http", "", false, "Use plain http and not https")

	return cmd
}
