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
	"github.com/hashicorp/consul/api"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/kv/consul"
	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/tasks/to"
)

// -----------------------------------------------------------------------------

type toConsulParams struct {
	inputPath    string
	secretAsLeaf bool
	prefix       string
}

var toConsulCmd = func() *cobra.Command {
	var params toConsulParams

	cmd := &cobra.Command{
		Use:   "consul",
		Short: "Publish bundle data into HashiCorp Consul",
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-kv-to-consul", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Create Consul client config from environment.
			config := api.DefaultConfig()

			// Creates a new client
			client, err := api.NewClient(config)
			if err != nil {
				log.For(ctx).Fatal("unable to connect to consul cluster", zap.Error(err))
				return
			}

			// Prepare store.
			store := consul.Store(client)
			defer log.SafeClose(store, "unable to close consul store")

			// Delegate to task
			t := &to.PublishKVTask{
				Store:           store,
				ContainerReader: cmdutil.FileReader(params.inputPath),
				SecretAsKey:     params.secretAsLeaf,
			}

			// Run the task
			if err := t.Run(ctx); err != nil {
				log.For(ctx).Fatal("unable to execute kv extract task", zap.Error(err))
				return
			}
		},
	}

	// Add parameters
	cmd.Flags().StringVar(&params.inputPath, "in", "-", "Container path ('-' for stdin or filename)")
	cmd.Flags().BoolVarP(&params.secretAsLeaf, "secret-as-leaf", "s", false, "Expand package path to secrets for provisioning")
	cmd.Flags().StringVar(&params.prefix, "prefix", "", "Path prefix for insertion")

	return cmd
}
