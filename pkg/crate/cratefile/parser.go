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

package cratefile

import (
	"fmt"
	"io"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

// ParseFile parses the given file for a configuration. The syntax of the
// file is determined based on the filename extension: "hcl" for HCL,
// "json" for JSON, other is an error.
func ParseFile(filename string) (*Config, error) {
	var config Config
	return &config, hclsimple.DecodeFile(filename, nil, &config)
}

// Parse parses the configuration from the given reader. The reader will be
// read to completion (EOF) before returning so ensure that the reader
// does not block forever.
//
// format is either "hcl" or "json"
func Parse(r io.Reader, filename, format string) (*Config, error) {
	src, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("unable to drain input reader: %w", err)
	}

	var config Config
	return &config, hclsimple.Decode("config.hcl", src, nil, &config)
}
