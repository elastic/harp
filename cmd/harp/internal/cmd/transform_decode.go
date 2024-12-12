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

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/sdk/value/encoding"
)

// -----------------------------------------------------------------------------

type transformDecodeParams struct {
	inputPath  string
	outputPath string
	encoding   string
}

var transformDecodeCmd = func() *cobra.Command {
	params := &transformDecodeParams{}

	longDesc := cmdutil.LongDesc(`
	Decode the given input stream using the selected decoding strategy.

	Supported codecs:
	  * identity - returns the unmodified input
	  * hex/base16 - returns the hexadecimal encoded input
	  * base32 - returns the Base32 encoded input
	  * base32hex - returns the Base32 with extended alphabet encoded input
	  * base64 - returns the Base64 encoded input
	  * base64raw - returns the Base64 encoded input without "=" padding
	  * base64url - returns the Base64 encoded input using URL safe characters
	  * base64urlraw - returns the Base64 encoded input using URL safe characters without "=" padding
	  * base85 - returns the Base85 encoded input`)

	examples := cmdutil.Examples(`
		# Decode base64 from stdin
		echo "dGVzdAo=" | harp transform decode --encoding base64

		# Decode base64url from a file
		harp transform decode --in test.txt --encoding base64url`)

	cmd := &cobra.Command{
		Use:     "decode",
		Short:   "Decode given input",
		Long:    longDesc,
		Example: examples,
		Run: func(cmd *cobra.Command, _ []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-transform-decode", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Read input
			reader, err := cmdutil.Reader(params.inputPath)
			if err != nil {
				log.For(ctx).Fatal("unable to initialize input reader", zap.Error(err))
			}

			// Output writer
			writer, err := cmdutil.Writer(params.outputPath)
			if err != nil {
				log.For(ctx).Fatal("unable to initialize output writer", zap.Error(err))
			}

			// Read and decode
			out, err := encoding.NewReader(reader, params.encoding)
			if err != nil {
				log.For(ctx).Fatal("unable to prepare input decoder", zap.Error(err))
			}

			// Process input as a stream.
			if _, err := io.Copy(writer, out); err != nil {
				log.For(ctx).Fatal("unable to process input", zap.Error(err))
			}
		},
	}

	// Parameters
	cmd.Flags().StringVar(&params.inputPath, "in", "-", "Input path ('-' for stdin or filename)")
	cmd.Flags().StringVar(&params.outputPath, "out", "-", "Output path ('-' for stdout or filename)")
	cmd.Flags().StringVar(&params.encoding, "encoding", "identity", "Encoding strategy")

	return cmd
}
