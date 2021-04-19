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
	"github.com/elastic/harp/pkg/tasks/share"
)

// -----------------------------------------------------------------------------

var shareGetCmd = func() *cobra.Command {
	var (
		outputPath    string
		backendPrefix string
		namespace     string
		token         string
	)

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get wrapped secret from Vault",
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "share-get", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Prepare task
			t := &share.GetTask{
				OutputWriter:   cmdutil.FileWriter(outputPath),
				BackendPrefix:  backendPrefix,
				VaultNamespace: namespace,
				Token:          token,
			}

			// Run the task
			if err := t.Run(ctx); err != nil {
				log.For(ctx).Fatal("unable to execute task", zap.Error(err))
			}
		},
	}

	// Parameters
	cmd.Flags().StringVar(&outputPath, "out", "-", "Output path ('-' for stdout or filename)")
	cmd.Flags().StringVar(&backendPrefix, "prefix", "cubbyhole", "Vault backend prefix")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Vault namespace")
	cmd.Flags().StringVar(&token, "token", "", "Wrapped token")
	log.CheckErr("unable to mark 'token' flag as required.", cmd.MarkFlagRequired("token"))

	return cmd
}
