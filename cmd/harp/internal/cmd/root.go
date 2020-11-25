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

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/elastic/harp/build/version"
	iconfig "github.com/elastic/harp/cmd/harp/internal/config"
	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/config"
	configcmd "github.com/elastic/harp/pkg/sdk/config/cmd"
	"github.com/elastic/harp/pkg/sdk/log"
)

// -----------------------------------------------------------------------------

var (
	cfgFile string
	conf    = &iconfig.Configuration{}
)

// -----------------------------------------------------------------------------

// RootCmd describes root command of the tool
var mainCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "harp",
		Short: "Extensible secret management tool",
	}

	// Register falgs
	cmd.Flags().StringVar(&cfgFile, "config", "", "config file")

	// Register sub commands
	cmd.AddCommand(version.Command())
	cmd.AddCommand(configcmd.NewConfigCommand(conf, "HARP"))

	cmd.AddCommand(bundleCmd())
	cmd.AddCommand(containerCmd())
	cmd.AddCommand(keygenCmd())
	cmd.AddCommand(passphraseCmd())
	cmd.AddCommand(docCmd())
	cmd.AddCommand(bugCmd())

	cmd.AddCommand(pluginCmd())
	cmd.AddCommand(csoCmd())

	cmd.AddCommand(templateCmd())
	cmd.AddCommand(valuesCmd())

	cmd.AddCommand(fromCmd())
	cmd.AddCommand(toCmd())

	cmd.AddCommand(transformCmd())

	// Return command
	return cmd
}

// -----------------------------------------------------------------------------

// Execute main command
func Execute() error {
	args := os.Args

	// Initialize configuration
	initConfig()

	// Initialize root command
	cmd := mainCmd()

	// Initialize plugin handler
	pluginHandler := cmdutil.NewDefaultPluginHandler(validPluginFilenamePrefixes)

	// If has more than 1 arguments
	if len(args) > 1 {
		cmdPathPieces := args[1:]

		// only look for suitable extension executables if
		// the specified command does not already exist
		if _, _, err := cmd.Find(cmdPathPieces); err != nil {
			if err := cmdutil.HandlePluginCommand(pluginHandler, cmdPathPieces); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		}
	}

	return cmd.Execute()
}

// -----------------------------------------------------------------------------

func initConfig() {
	if err := config.Load(conf, "HARP", cfgFile); err != nil {
		log.Bg().Fatal("Unable to load settings", zap.Error(err))
	}
}
