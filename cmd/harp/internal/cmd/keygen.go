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

	"github.com/elastic/harp/build/fips"
)

// -----------------------------------------------------------------------------

var keygenCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "keygen",
		Aliases: []string{"kg"},
		Short:   "Key generation commands",
	}

	// Subcommands
	cmd.AddCommand(keygenFernetCmd())
	cmd.AddCommand(keygenAESCmd())
	cmd.AddCommand(keygenMasterKeyCmd())

	if !fips.Enabled() {
		cmd.AddCommand(keygenSecretBoxCmd())
		cmd.AddCommand(keygenChaChaCmd())
		cmd.AddCommand(keygenXChaChaCmd())
		cmd.AddCommand(keygenAESPMACSIVCmd())
		cmd.AddCommand(keygenAESSIVCmd())
		cmd.AddCommand(keygenPasetoCmd())
	}
	return cmd
}
