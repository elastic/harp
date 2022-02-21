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
	"io"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/sdk/value/encoding"
	"github.com/elastic/harp/pkg/sdk/value/encryption"
)

// -----------------------------------------------------------------------------
type transformDecryptParams struct {
	inputPath              string
	outputPath             string
	keyRaw                 string
	additionalData         string
	additionalDataEncoding string
}

var transformDecryptCmd = func() *cobra.Command {
	params := &transformDecryptParams{}

	cmd := &cobra.Command{
		Use:     "decrypt",
		Short:   "Decrypt the given value with a value transformer",
		Aliases: []string{"d"},
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-transform-decrypt", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Resolve tranformer
			t, err := encryption.FromKey(params.keyRaw)
			if err != nil {
				log.For(ctx).Fatal("unable to initialize a transformer form key", zap.Error(err))
			}
			if t == nil {
				log.For(ctx).Fatal("transformer is nil")
			}

			// Read input
			reader, err := cmdutil.Reader(params.inputPath)
			if err != nil {
				log.For(ctx).Fatal("unable to initialize input reader", zap.Error(err))
			}

			// Read input
			writer, err := cmdutil.Writer(params.outputPath)
			if err != nil {
				log.For(ctx).Fatal("unable to initialize output writer", zap.Error(err))
			}

			// Drain reader
			content, err := io.ReadAll(reader)
			if err != nil {
				log.For(ctx).Fatal("unable to drain input reader", zap.Error(err))
			}

			// Decode AAD if any
			if params.additionalData != "" {
				encoderReader, errDecode := encoding.NewReader(strings.NewReader(params.additionalData), params.additionalDataEncoding)
				if errDecode != nil {
					log.For(ctx).Fatal("unable to decode additional data", zap.Error(errDecode))
				}

				aad, errAADRead := io.ReadAll(encoderReader)
				if errAADRead != nil {
					log.For(ctx).Fatal("unable to read additional data", zap.Error(errAADRead))
				}

				// Set additional data
				ctx = encryption.WithAdditionalData(ctx, aad)
			}

			// Apply transformation
			out, err := t.From(ctx, content)
			if err != nil {
				log.For(ctx).Fatal("unable to apply transformer", zap.Error(err))
			}

			if _, err = writer.Write(out); err != nil {
				log.For(ctx).Fatal("unable to write result to writer", zap.Error(err))
			}
		},
	}

	// Parameters
	cmd.Flags().StringVar(&params.keyRaw, "key", "", "Transformer key")
	log.CheckErr("unable to mark 'key' flag as required.", cmd.MarkFlagRequired("key"))

	cmd.Flags().StringVar(&params.inputPath, "in", "-", "Input path ('-' for stdin or filename)")
	cmd.Flags().StringVar(&params.outputPath, "out", "-", "Output path ('-' for stdout or filename)")
	cmd.Flags().StringVar(&params.additionalData, "aad", "", "Additional data for AEAD encryption")
	cmd.Flags().StringVar(&params.additionalDataEncoding, "aad-encoding", "base64", "Additional data encoding")

	return cmd
}
