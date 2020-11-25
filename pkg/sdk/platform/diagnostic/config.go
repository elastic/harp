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

package diagnostic

// Config holds information for diagnostic handlers.
type Config struct {
	GOPS struct {
		Enabled   bool   `toml:"enabled" default:"false" comment:"Enable GOPS agent"`
		RemoteURL string `toml:"remoteDebugURL" comment:"start a gops agent on specified URL. Ex: localhost:9999"`
	}
	PProf struct {
		Enabled bool `toml:"enabled" default:"true" comment:"Enable PProf handler"`
	}
	ZPages struct {
		Enabled bool `toml:"enabled" default:"true" comment:"Enable zPages handler"`
	}
}

// Validate checks that the configuration is valid.
func (c *Config) Validate() error {
	return nil
}
