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

	"github.com/hashicorp/consul/api"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/kv/consul"
	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/tasks/from"
)

// -----------------------------------------------------------------------------

type fromConsulParams struct {
	outputPath           string
	basePaths            []string
	lastPathItemAsSecret bool
}

var fromConsulCmd = func() *cobra.Command {
	var params fromConsulParams

	cmd := &cobra.Command{
		Use:   "consul",
		Short: "Extract KV pairs from Hashicorp Consul KV Store",
		Run: func(cmd *cobra.Command, _ []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-kv-from-consul", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			runFromConsul(ctx, &params)
		},
	}

	// Add parameters
	cmd.Flags().StringVar(&params.outputPath, "out", "-", "Container output path ('-' for stdout)")
	cmd.Flags().StringSliceVar(&params.basePaths, "paths", []string{}, "Exported base paths")
	cmd.Flags().BoolVarP(&params.lastPathItemAsSecret, "last-path-item-as-secret-key", "k", false, "Use the last path element as secret key")

	return cmd
}

func runFromConsul(ctx context.Context, params *fromConsulParams) {
	// Create Consul client config from environment.
	config := api.DefaultConfig()

	// Creates a new client
	client, err := api.NewClient(config)
	if err != nil {
		log.For(ctx).Fatal("unable to connect to consul cluster", zap.Error(err))
		return
	}

	// Prepare store.
	store := consul.Store(client.KV())
	defer log.SafeClose(store, "unable to close consul store")

	// Delegate to task
	t := &from.ExtractKVTask{
		Store:                   store,
		ContainerWriter:         cmdutil.FileWriter(params.outputPath),
		BasePaths:               params.basePaths,
		LastPathItemAsSecretKey: params.lastPathItemAsSecret,
	}

	// Run the task
	if err := t.Run(ctx); err != nil {
		log.For(ctx).Fatal("unable to execute kv extract task", zap.Error(err))
		return
	}
}
