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
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/sdk/security/diceware"
)

var passphraseWordCount int8

// -----------------------------------------------------------------------------

var passphraseCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "passphrase",
		Short: "Generate and print a diceware passphrase",
		Run:   runPassphrase,
	}

	// Parameters
	cmd.Flags().Int8VarP(&passphraseWordCount, "word-count", "w", 8, "Word count in diceware passphrase")

	return cmd
}

func runPassphrase(cmd *cobra.Command, _ []string) {
	ctx, cancel := cmdutil.Context(cmd.Context(), "harp-passphrase", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
	defer cancel()

	// Check lower limit
	if passphraseWordCount < 4 {
		passphraseWordCount = 4
	}

	// Generate passphrase
	passPhrase, err := diceware.Diceware(int(passphraseWordCount))
	if err != nil {
		log.For(ctx).Fatal("unable to generate diceware passphrase", zap.Error(err))
	}

	// Print the key
	// lgtm [go/clear-text-logging]
	fmt.Fprintln(os.Stdout, passPhrase)
}
