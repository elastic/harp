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
	"gopkg.in/square/go-jose.v2"

	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/tasks/keygen"
)

// -----------------------------------------------------------------------------
type keygenJWKParams struct {
	outputPath         string
	signatureAlgorithm string
	keyBits            int
	keyID              string
}

var keygenKeypairCmd = func() *cobra.Command {
	params := &keygenJWKParams{}

	cmd := &cobra.Command{
		Use:   "jwk",
		Short: "Generate a JWK encoded key pair",
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-keygen-jwk", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Prepare task
			t := &keygen.JWKTask{
				OutputWriter:       cmdutil.FileWriter(params.outputPath),
				SignatureAlgorithm: params.signatureAlgorithm,
				KeySize:            params.keyBits,
				KeyID:              params.keyID,
			}

			// Run the task
			if err := t.Run(ctx); err != nil {
				log.For(ctx).Fatal("unable to execute task", zap.Error(err))
			}
		},
	}

	// Add parameters
	cmd.Flags().StringVar(&params.signatureAlgorithm, "algorithm", string(jose.EdDSA), "Key type to generate")
	cmd.Flags().IntVar(&params.keyBits, "bits", 0, "Key size (in bits)")
	cmd.Flags().StringVar(&params.keyID, "key-id", "", "Key identifier")

	return cmd
}
