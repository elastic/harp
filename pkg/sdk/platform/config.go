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

package platform

import (
	"github.com/elastic/harp/pkg/sdk/platform/diagnostic"
)

// InstrumentationConfig holds all platform instrumentation settings
type InstrumentationConfig struct {
	Network    string `toml:"network" default:"tcp" comment:"Network class used for listen (tcp, tcp4, tcp6, unixsocket)"`
	Listen     string `toml:"listen" default:":5556" comment:"Listen address for instrumentation server"`
	Diagnostic struct {
		Enabled bool              `toml:"enabled" default:"false" comment:"Enable diagnostic handlers"`
		Config  diagnostic.Config `toml:"Config" comment:"Diagnostic settings"`
	} `toml:"Diagnostic" comment:"###############################\n Diagnotic Settings \n##############################"`
	Logs struct {
		Level string `toml:"level" default:"warn" comment:"Log level: debug, info, warn, error, dpanic, panic, and fatal"`
	} `toml:"Logs" comment:"###############################\n Logs Settings \n##############################"`
}
