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

package config

import (
	"fmt"
	"os"
	"strings"

	defaults "github.com/mcuadros/go-defaults"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/sdk/flags"
	"github.com/elastic/harp/pkg/sdk/log"
)

// Load a config
// Apply defaults first, then environment, then finally config file.
func Load(conf interface{}, envPrefix, cfgFile string) error {
	// Apply defaults first
	defaults.SetDefaults(conf)

	// Uppercase the prefix
	upPrefix := strings.ToUpper(envPrefix)

	// Overrides with environment
	for k := range flags.AsEnvVariables(conf, "", false) {
		envName := fmt.Sprintf("%s_%s", upPrefix, k)
		log.CheckErr("unable to bind environment variable", viper.BindEnv(strings.ToLower(strings.Replace(k, "_", ".", -1)), envName), zap.String("var", envName))
	}

	// Apply file settings
	if cfgFile != "" {
		// If the config file doesn't exists, let's exit
		if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
			return fmt.Errorf("config: unable to open non-existing file '%s': %w", cfgFile, err)
		}

		log.Bg().Info("Load settings from file", zap.String("path", cfgFile))

		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("config: unable to decode config file '%s': %w", cfgFile, err)
		}
	}

	// Update viper values
	if err := viper.Unmarshal(conf); err != nil {
		return fmt.Errorf("config: unable to apply config '%s': %w", cfgFile, err)
	}

	// No error
	return nil
}
