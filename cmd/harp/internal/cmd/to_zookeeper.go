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
	"time"

	zk "github.com/go-zookeeper/zk"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/kv/zookeeper"
	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/tasks/to"
)

// -----------------------------------------------------------------------------

type toZookeeperParams struct {
	inputPath    string
	secretAsLeaf bool
	prefix       string

	endpoints   []string
	dialTimeout time.Duration
}

var toZookeeperCmd = func() *cobra.Command {
	var params toZookeeperParams

	cmd := &cobra.Command{
		Use:     "zookeeper",
		Aliases: []string{"zk"},
		Short:   "Publish bundle data into Apache Zookeeper",
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-kv-to-zookeeper", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

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
			t := &to.PublishKVTask{
				Store:           store,
				ContainerReader: cmdutil.FileReader(params.inputPath),
				SecretAsKey:     params.secretAsLeaf,
				Prefix:          params.prefix,
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

	cmd.Flags().StringArrayVar(&params.endpoints, "endpoints", []string{"127.0.0.1:2181"}, "Zookeeper client endpoints")
	cmd.Flags().DurationVar(&params.dialTimeout, "dial-timeout", 15*time.Second, "Zookeeper client dial timeout")

	return cmd
}
