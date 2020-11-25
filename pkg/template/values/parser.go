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
	"fmt"

	"github.com/elastic/harp/pkg/template/values/hcl1"
	"github.com/elastic/harp/pkg/template/values/hcl2"
	"github.com/elastic/harp/pkg/template/values/hocon"
	"github.com/elastic/harp/pkg/template/values/toml"
	"github.com/elastic/harp/pkg/template/values/xml"
	"github.com/elastic/harp/pkg/template/values/yaml"
)

// Parser is the interface implemented by objects that can unmarshal
// bytes into a golang interface
type Parser interface {
	Unmarshal(p []byte, v interface{}) error
}

// GetParser gets a file parser based on the file type and input
func GetParser(fileType string) (Parser, error) {
	switch fileType {
	case "toml":
		return &toml.Parser{}, nil
	case "hocon":
		return &hocon.Parser{}, nil
	case "xml":
		return &xml.Parser{}, nil
	case "json", "yaml", "yml":
		return &yaml.Parser{}, nil
	case "hcl", "tf", "hcl2", "tfvars":
		return &hcl2.Parser{}, nil
	case "hcl1":
		return &hcl1.Parser{}, nil
	default:
		return nil, fmt.Errorf("unknown filetype given: %v", fileType)
	}
}
