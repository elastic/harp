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

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/tasks/share"
)

// -----------------------------------------------------------------------------

var sharePutCmd = func() *cobra.Command {
	var (
		inputPath     string
		backendPrefix string
		namespace     string
		ttl           time.Duration
		jsonOutput    bool
	)

	cmd := &cobra.Command{
		Use:   "put",
		Short: "Put secret in Vault Cubbyhole and return a wrapped token",
		Run: func(cmd *cobra.Command, _ []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "share-put", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Prepare task
			t := &share.PutTask{
				InputReader:    cmdutil.FileReader(inputPath),
				OutputWriter:   cmdutil.StdoutWriter(),
				BackendPrefix:  backendPrefix,
				VaultNamespace: namespace,
				TTL:            ttl,
				JSONOutput:     jsonOutput,
			}

			// Run the task
			if err := t.Run(ctx); err != nil {
				log.For(ctx).Fatal("unable to execute task", zap.Error(err))
			}
		},
	}

	// Parameters
	cmd.Flags().StringVar(&inputPath, "in", "-", "Input path ('-' for stdin or filename)")
	cmd.Flags().StringVar(&backendPrefix, "prefix", "cubbyhole", "Vault backend prefix")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Vault namespace")
	cmd.Flags().DurationVar(&ttl, "ttl", 30*time.Second, "Token expiration")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Display result as json")

	return cmd
}
