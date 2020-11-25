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
	"fmt"

	"github.com/fernet/fernet-go"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
)

// -----------------------------------------------------------------------------

var keygenFernetCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fernet",
		Short: "Generate and print a fernet key",
		Run:   runKeygenFernet,
	}

	return cmd
}

func runKeygenFernet(cmd *cobra.Command, args []string) {
	ctx, cancel := cmdutil.Context(cmd.Context(), "harp-keygen-fernet", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
	defer cancel()

	// Generate a fernet key
	k := &fernet.Key{}
	if err := k.Generate(); err != nil {
		log.For(ctx).Fatal("unable to generate Fernet key", zap.Error(err))
	}

	// Print the key
	fmt.Print(k.Encode())
}
