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

type cratePushParams struct {
	inputPath  string
	outputPath string
	to         string
	ref        string
	json       bool
	opts       content.RegistryOptions
}

var cratePushCmd = func() *cobra.Command {
	params := &cratePushParams{}

	longDesc := cmdutil.LongDesc(`
	Export a crate to an OCI compatible registry.`)

	examples := cmdutil.Examples(``)

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
				SpecReader:   cmdutil.FileReader(params.inputPath),
				OutputWriter: cmdutil.FileWriter(params.outputPath),
				Target:       params.to,
				Ref:          params.ref,
				JSONOutput:   params.json,
				RegistryOpts: params.opts,
			}

			// Run the task
			if err := t.Run(ctx); err != nil {
				log.For(ctx).Fatal("unable to execute task", zap.Error(err))
			}
		},
	}

	// Parameters
	cmd.Flags().StringVarP(&params.inputPath, "cratefile", "f", "Cratefile", "Specification path ('-' for stdin or filename)")
	cmd.Flags().StringVar(&params.outputPath, "out", "-", "Output path ('-' for stdout or filename)")
	cmd.Flags().StringVar(&params.to, "to", "", "Target destination (registry:<url>, oci:<path>, files:<path>)")
	log.CheckErr("unable to mark 'to' flag as required.", cmd.MarkFlagRequired("to"))
	cmd.Flags().StringVar(&params.ref, "ref", "harp.sealed", "Container path")
	cmd.Flags().BoolVar(&params.json, "json", false, "Enable JSON output")
	cmd.Flags().StringArrayVarP(&params.opts.Configs, "config", "c", nil, "Authentication config path")
	cmd.Flags().StringVarP(&params.opts.Username, "username", "u", "", "Registry username")
	cmd.Flags().StringVarP(&params.opts.Password, "password", "p", "", "Registry password")
	cmd.Flags().BoolVarP(&params.opts.Insecure, "insecure", "", false, "Allow connections to SSL registry without certs")
	cmd.Flags().BoolVarP(&params.opts.PlainHTTP, "plain-http", "", false, "Use plain http and not https")

	return cmd
}
