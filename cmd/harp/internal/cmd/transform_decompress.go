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
	"github.com/elastic/harp/pkg/sdk/value/compression"
)

// -----------------------------------------------------------------------------

type transformDecompressParams struct {
	inputPath  string
	outputPath string
	algorithm  string
}

var transformDecompressCmd = func() *cobra.Command {
	params := &transformDecompressParams{}

	longDesc := cmdutil.LongDesc(`
	Decompress the given input stream using the selected compression algorithm.

	Supported compression:
	  * identity - returns the unmodified input
	  * gzip
	  * lzw/lzw-msb/lzw-lsb
	  * lz4
	  * s2/snappy
	  * zlib
	  * flate/deflate
	  * lzma
	  * zstd`)

	examples := cmdutil.Examples(`
	# Compress a file
	harp transform decompress --in README.md.gz --out README.md --algorithm gzip

	# Decompress to STDOUT
	harp transform compress --in README.md.gz --algorithm gzip

	# Decompress from STDIN
	harp transform compress --algorithm gzip`)

	cmd := &cobra.Command{
		Use:     "decompress",
		Short:   "Decompress given input",
		Long:    longDesc,
		Example: examples,
		Run: func(cmd *cobra.Command, _ []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-transform-decompress", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
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

			// Prepare compressor
			compressedReader, err := compression.NewReader(reader, params.algorithm)
			if err != nil {
				log.SafeClose(compressedReader, "unable to close the compression writer")
				log.For(ctx).Fatal("unable to write encoded content", zap.Error(err))
			}

			// Process input as a stream.
			if _, err := io.Copy(writer, compressedReader); err != nil {
				log.SafeClose(compressedReader, "unable to close the compression writer")
				log.For(ctx).Fatal("unable to process input", zap.Error(err))
			}

			// Close the writer
			log.SafeClose(compressedReader, "unable to close the compression writer")
		},
	}

	// Parameters
	cmd.Flags().StringVar(&params.inputPath, "in", "-", "Input path ('-' for stdin or filename)")
	cmd.Flags().StringVar(&params.outputPath, "out", "-", "Output path ('-' for stdout or filename)")
	cmd.Flags().StringVar(&params.algorithm, "algorithm", "gzip", "Compression algorithm")

	return cmd
}
