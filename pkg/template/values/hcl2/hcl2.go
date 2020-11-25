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

package hcl2

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// Parser is a HCL2 parser
type Parser struct{}

// Unmarshal HCL2.0 scripts.
func (h *Parser) Unmarshal(p []byte, v interface{}) error {
	file, diags := hclsyntax.ParseConfig(p, "", hcl.Pos{Byte: 0, Line: 1, Column: 1})

	if diags.HasErrors() {
		var details []error
		details = append(details, diags.Errs()...)
		return fmt.Errorf("parse hcl2 config: \n %s", details)
	}

	content, err := convertFile(file)
	if err != nil {
		return fmt.Errorf("convert hcl2 to json: %w", err)
	}

	j, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("marshal hcl2 to json: %w", err)
	}

	if err := json.Unmarshal(j, v); err != nil {
		return fmt.Errorf("unmarshal hcl2 json: %w", err)
	}

	return nil
}
