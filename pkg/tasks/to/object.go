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

package to

import (
	"context"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/elastic/harp/pkg/bundle"
	"github.com/elastic/harp/pkg/sdk/value/flatmap"
	"github.com/elastic/harp/pkg/tasks"
)

// ObjectTask implements secret-container publication process to json/yaml content.
type ObjectTask struct {
	ContainerReader tasks.ReaderProvider
	OutputWriter    tasks.WriterProvider
	Expand          bool
	JSON            bool
	YAML            bool
}

// Run the task.
func (t *ObjectTask) Run(ctx context.Context) error {
	// Create the reader
	reader, err := t.ContainerReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to open input bundle reader: %w", err)
	}

	// Create output writer
	writer, err := t.OutputWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to open writer: %w", err)
	}

	// Extract bundle from container
	b, err := bundle.FromContainerReader(reader)
	if err != nil {
		return fmt.Errorf("unable to load bundle: %w", err)
	}

	// Convert as map
	bundleMap, err := bundle.AsMap(b)
	if err != nil {
		return fmt.Errorf("unable to transform the bundle as a map: %w", err)
	}

	var toEncode interface{}

	// Expand if required
	if t.Expand {
		toEncode = flatmap.Expand(bundleMap, "")
	} else {
		toEncode = bundleMap
	}

	// Select strategy
	switch {
	case t.JSON:
		// Encode as JSON
		if err := json.NewEncoder(writer).Encode(toEncode); err != nil {
			return fmt.Errorf("unable to marshal JSON bundle content: %w", err)
		}
	case t.YAML:
		// Encode as YAML
		if err := yaml.NewEncoder(writer).Encode(toEncode); err != nil {
			return fmt.Errorf("unable to marshal YAML bundle content: %w", err)
		}
	default:
		// Encode as JSON
		if err := json.NewEncoder(writer).Encode(toEncode); err != nil {
			return fmt.Errorf("unable to marshal JSON bundle content: %w", err)
		}
	}

	// No error
	return nil
}
