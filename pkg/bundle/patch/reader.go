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

package patch

import (
	"fmt"
	"io"

	"google.golang.org/protobuf/encoding/protojson"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/sdk/convert"
	"github.com/elastic/harp/pkg/sdk/types"
)

// YAML a given reader in order to extract a BundlePatch sepcification
func YAML(r io.Reader) (*bundlev1.Patch, error) {
	// Check arguments
	if types.IsNil(r) {
		return nil, fmt.Errorf("reader is nil")
	}

	// Drain the reader
	jsonReader, err := convert.YAMLtoJSON(r)
	if err != nil {
		return nil, fmt.Errorf("unable to parse input as BundlePatch: %w", err)
	}

	// Drain reader
	jsonData, err := io.ReadAll(jsonReader)
	if err != nil {
		return nil, fmt.Errorf("unable to drain all json reader content: %w", err)
	}

	// Initialize empty definition object
	def := bundlev1.Patch{}
	def.Reset()

	// Deserialize JSON with JSONPB wrapper
	if err := protojson.Unmarshal(jsonData, &def); err != nil {
		return nil, fmt.Errorf("unable to decode spec as json: %w", err)
	}

	// Validate spec
	if err := Validate(&def); err != nil {
		return nil, fmt.Errorf("unable to validate descriptor: %w", err)
	}

	// No error
	return &def, nil
}
