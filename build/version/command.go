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

package version

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// -----------------------------------------------------------------------------

var (
	displayAsJSON bool
	withModules   bool
)

// Command exports Cobra command builder
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Display service version",
		Run: func(_ *cobra.Command, _ []string) {
			bi := NewInfo()
			if displayAsJSON {
				_, _ = fmt.Fprintf(os.Stdout, "%s", bi.JSON())
			} else {
				_, _ = fmt.Fprintf(os.Stdout, "%s", bi.String())
				if withModules {
					_, _ = fmt.Fprintln(os.Stdout, "\nDependencies:")
					for _, dep := range bi.BuildDeps {
						_, _ = fmt.Fprintf(os.Stdout, "- %s\n", dep)
					}
				}
			}
		},
	}

	// Register parameters
	cmd.Flags().BoolVar(&displayAsJSON, "json", false, "Display build info as json")
	cmd.Flags().BoolVar(&withModules, "with-modules", false, "Display builtin go modules")

	// Return command
	return cmd
}
