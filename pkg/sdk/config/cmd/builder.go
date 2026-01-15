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
	"sort"
	"strings"

	defaults "github.com/mcuadros/go-defaults"
	toml "github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/sdk/flags"
	"github.com/elastic/harp/pkg/sdk/log"
)

var configNewAsEnvFlag bool

// NewConfigCommand initialize a cobra config command tree
func NewConfigCommand(conf interface{}, envPrefix string) *cobra.Command {
	// Uppercase the prefix
	upPrefix := strings.ToUpper(envPrefix)

	// config
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage Service Configuration",
	}

	// config new
	configNewCmd := &cobra.Command{
		Use:   "new",
		Short: "Initialize a default configuration",
		Run: func(cmd *cobra.Command, _ []string) {
			defaults.SetDefaults(conf)

			if !configNewAsEnvFlag {
				btes, err := toml.Marshal(conf)
				if err != nil {
					log.For(cmd.Context()).Fatal("Error during configuration export", zap.Error(err))
				}
				_, _ = fmt.Fprintln(os.Stdout, string(btes))
			} else {
				m := flags.AsEnvVariables(conf, upPrefix, true)
				keys := []string{}

				for k := range m {
					keys = append(keys, k)
				}

				sort.Strings(keys)
				for _, k := range keys {
					_, _ = fmt.Fprintf(os.Stdout, "export %s=\"%s\"\n", k, m[k])
				}
			}
		},
	}

	// flags
	configNewCmd.Flags().BoolVar(&configNewAsEnvFlag, "env", false, "Print configuration as environment variable")
	configCmd.AddCommand(configNewCmd)

	// Return base command
	return configCmd
}
