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

package values

import (
	"reflect"
	"testing"

	"github.com/elastic/harp/pkg/template/values/hcl1"
	"github.com/elastic/harp/pkg/template/values/hcl2"
	"github.com/elastic/harp/pkg/template/values/hocon"
	"github.com/elastic/harp/pkg/template/values/toml"
	"github.com/elastic/harp/pkg/template/values/xml"
	"github.com/elastic/harp/pkg/template/values/yaml"
)

func TestGetParser(t *testing.T) {
	testTable := []struct {
		name        string
		fileType    string
		expected    Parser
		expectError bool
	}{
		{
			name:        "Test getting HOCON parser",
			fileType:    "hocon",
			expected:    new(hocon.Parser),
			expectError: false,
		},
		{
			name:        "Test getting TOML parser",
			fileType:    "toml",
			expected:    new(toml.Parser),
			expectError: false,
		},
		{
			name:        "Test getting XML parser",
			fileType:    "xml",
			expected:    new(xml.Parser),
			expectError: false,
		},
		{
			name:        "Test getting Terraform parser from HCL1 input",
			fileType:    "hcl1",
			expected:    new(hcl1.Parser),
			expectError: false,
		},
		{
			name:        "Test getting Terraform parser from HCL2 input",
			fileType:    "tf",
			expected:    new(hcl2.Parser),
			expectError: false,
		},
		{
			name:        "Test getting Terraform parser from YAML input",
			fileType:    "yaml",
			expected:    new(yaml.Parser),
			expectError: false,
		},
		{
			name:        "Test getting Terraform parser from JSON input",
			fileType:    "json",
			expected:    new(yaml.Parser),
			expectError: false,
		},
		{
			name:        "Test getting invalid filetype",
			fileType:    "epicfailure",
			expected:    nil,
			expectError: true,
		},
	}

	for _, testUnit := range testTable {
		t.Run(testUnit.name, func(t *testing.T) {
			received, err := GetParser(testUnit.fileType)

			if !reflect.DeepEqual(received, testUnit.expected) {
				t.Errorf("expected: %T \n got this: %T", testUnit.expected, received)
			}
			if !testUnit.expectError && err != nil {
				t.Errorf("error here: %v", err)
			}
			if testUnit.expectError && err == nil {
				t.Error("error expected but not received")
			}
		})
	}
}
