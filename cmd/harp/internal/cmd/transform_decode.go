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
	"encoding/hex"
	"io"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
)

// -----------------------------------------------------------------------------

type transformDecodeParams struct {
	inputPath  string
	outputPath string
	encoding   string
}

var transformDecodeCmd = func() *cobra.Command {
	params := &transformDecodeParams{}

	cmd := &cobra.Command{
		Use:   "decode",
		Short: "Decode given input",
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-transform-decode", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

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

			var (
				out       []byte
				errDecode error
			)
			// Apply transformation
			switch params.encoding {
			case "identity":
				out = content
			case "hex":
				out, errDecode = hex.DecodeString(string(content))
			case "base64":
				out, errDecode = base64.StdEncoding.DecodeString(string(content))
			case "base64raw":
				out, errDecode = base64.RawStdEncoding.DecodeString(string(content))
			case "base64url":
				out, errDecode = base64.URLEncoding.DecodeString(string(content))
			case "base64urlraw":
				out, errDecode = base64.RawURLEncoding.DecodeString(string(content))
			default:
				log.For(ctx).Fatal("unhandled decoding strategy", zap.String("encoding", params.encoding))
			}
			if errDecode != nil {
				log.For(ctx).Fatal("unable to decode input", zap.Error(err))
			}

			if _, err = writer.Write(out); err != nil {
				log.For(ctx).Fatal("unable to write result to writer", zap.Error(err))
			}
		},
	}

	// Parameters
	cmd.Flags().StringVar(&params.inputPath, "in", "-", "Input path ('-' for stdin or filename)")
	cmd.Flags().StringVar(&params.outputPath, "out", "-", "Output path ('-' for stdin or filename)")
	cmd.Flags().StringVar(&params.encoding, "encoding", "identity", "Encoding strategy (hex, base64, base64raw, base64url, base64urlraw)")

	return cmd
}
