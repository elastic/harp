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
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/sdk/value/hash"
)

// -----------------------------------------------------------------------------

type transformMultihashParams struct {
	inputPath  string
	outputPath string
	algorithms []string
	encoding   string
	jsonOutput bool
}

var transformMultihashCmd = func() *cobra.Command {
	params := &transformMultihashParams{}

	longDesc := cmdutil.LongDesc(fmt.Sprintf(`
		Process the input to compute the hashes according to selected hash algorithms.

		The command input is limited to size lower than 250 MB.

		Supported Algorithms:
		  %s`, strings.Join(hash.SupportedAlgorithms(), ", ")))

	examples := cmdutil.Examples(`
	# Compute md5, sha1, sha256, sha512 in one read from a file
	harp transform multihash --in livecd.iso

	# Compute sha256, sha512 only
	harp transform multihash --algorithm sha256 --algorithm sha512 --in livecd.iso

	# Compute sha256, sha512 only with JSON ouput
	harp transform multihash --json --algorithm sha256 --algorithm sha512 --in livecd.iso
	`)

	cmd := &cobra.Command{
		Use:     "multihash",
		Aliases: []string{"mh"},
		Short:   "Multiple hash  from given input",
		Long:    longDesc,
		Example: examples,
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-transform-multihash", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
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

			// Prepare hasher
			hMap, err := hash.NewMultiHash(reader, params.algorithms...)
			if err != nil {
				log.For(ctx).Fatal("unable to initialize hasher", zap.Error(err))
			}

			// Display as json
			if params.jsonOutput {
				if err := json.NewEncoder(writer).Encode(hMap); err != nil {
					log.For(ctx).Fatal("unable to encode result as json", zap.Error(err))
				}
			} else {
				// Sort map keys to get a stable output.
				keys := make([]string, 0, len(hMap))
				for k := range hMap {
					keys = append(keys, k)
				}
				sort.Strings(keys)

				for _, k := range keys {
					// Display container key
					if _, err := fmt.Fprintf(writer, "%s=%s\n", k, hMap[k]); err != nil {
						log.For(ctx).Fatal("unable to display result", zap.Error(err))
					}
				}
			}
		},
	}

	// Parameters
	cmd.Flags().StringVar(&params.inputPath, "in", "-", "Input path ('-' for stdin or filename)")
	cmd.Flags().StringVar(&params.outputPath, "out", "-", "Output path ('-' for stdout or filename)")
	cmd.Flags().StringSliceVar(&params.algorithms, "algorithm", []string{"md5", "sha1", "sha256", "sha512"}, "Hash algorithms to use")
	cmd.Flags().BoolVar(&params.jsonOutput, "json", false, "Display multihash result as json")

	return cmd
}
