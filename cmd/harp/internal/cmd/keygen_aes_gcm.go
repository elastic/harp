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
	"encoding/base64"
	"fmt"
	"os"

	"github.com/awnumar/memguard"
	"github.com/spf13/cobra"

	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
)

// -----------------------------------------------------------------------------

var keygenAESCmd = func() *cobra.Command {
	var keySize uint16

	cmd := &cobra.Command{
		Use:     "aes-gcm",
		Aliases: []string{"aes"},
		Short:   "Generate and print an AES-GCM key",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-keygen-aes", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Validate key size
			switch keySize {
			case 128, 192, 256:
				break
			default:
				log.For(ctx).Fatal("invalid specificed key size, only 128, 192 and 256 are supported.")
			}

			fmt.Fprintf(os.Stdout, "aes-gcm:%s", base64.URLEncoding.EncodeToString(memguard.NewBufferRandom(int(keySize/8)).Bytes()))
		},
	}

	// Parameters
	cmd.Flags().Uint16Var(&keySize, "size", 128, "Specify an AES key size (128, 192, 256)")

	return cmd
}
