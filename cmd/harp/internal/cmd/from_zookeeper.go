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
	"time"

	zk "github.com/go-zookeeper/zk"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/kv/zookeeper"
	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/tasks/from"
)

// -----------------------------------------------------------------------------

type fromZookeeperParams struct {
	outputPath           string
	basePaths            []string
	lastPathItemAsSecret bool

	endpoints   []string
	dialTimeout time.Duration
}

var fromZookeeperCmd = func() *cobra.Command {
	var params fromZookeeperParams
	cmd := &cobra.Command{
		Use:     "zookeeper",
		Aliases: []string{"zk"},
		Short:   "Extract KV pairs from Apache Zookeeper KV Store",
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-kv-from-zookeeper", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			runFromZookeeper(ctx, &params)
		},
	}

	// Add parameters
	cmd.Flags().StringArrayVar(&params.endpoints, "endpoints", []string{"127.0.0.1:2181"}, "Zookeeper client endpoints")
	cmd.Flags().DurationVar(&params.dialTimeout, "dial-timeout", 15*time.Second, "Zookeeper client dial timeout")
	cmd.Flags().BoolVarP(&params.lastPathItemAsSecret, "last-path-item-as-secret-key", "k", false, "Use the last path element as secret key")

	return cmd
}

func runFromZookeeper(ctx context.Context, params *fromZookeeperParams) {
	// Create config
	client, _, err := zk.Connect(params.endpoints, params.dialTimeout)
	if err != nil {
		log.For(ctx).Fatal("unable to connect to zookeeper cluster", zap.Error(err))
		return
	}

	// Prepare store.
	store := zookeeper.Store(client)
	defer log.SafeClose(store, "unable to close zk store")

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
