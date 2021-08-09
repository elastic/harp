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

package from

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle"
	"github.com/elastic/harp/pkg/sdk/value/flatmap"
	"github.com/elastic/harp/pkg/tasks"
)

// ObjectTask implements secret-container creation from a YAML/JSON structure.
type ObjectTask struct {
	ObjectReader tasks.ReaderProvider
	OutputWriter tasks.WriterProvider
	JSON         bool
	YAML         bool
}

// Run the task.
func (t *ObjectTask) Run(ctx context.Context) error {
	var (
		reader io.Reader
		writer io.Writer
		b      *bundlev1.Bundle
		err    error
	)

	// Create input reader
	reader, err = t.ObjectReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to read input reader: %w", err)
	}

	// Decode as YAML any object
	var source map[string]interface{}

	switch {
	case t.YAML:
		if errYaml := yaml.NewDecoder(reader).Decode(&source); errYaml != nil {
			return fmt.Errorf("unable to decode source as YAML: %w", err)
		}
	case t.JSON:
		if errYaml := json.NewDecoder(reader).Decode(&source); errYaml != nil {
			return fmt.Errorf("unable to decode source as JSON: %w", err)
		}
	default:
		return errors.New("json or yaml must be selected")
	}

	// Flatten the struct
	input := flatmap.Flatten(source)

	// Build the container from json
	b, err = bundle.FromMap(input)
	if err != nil {
		return fmt.Errorf("unable to create container from map: %w", err)
	}

	// Create output writer
	writer, err = t.OutputWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to open output writer: %w", err)
	}

	// Dump bundle
	if err = bundle.ToContainerWriter(writer, b); err != nil {
		return fmt.Errorf("unable to produce exported bundle: %w", err)
	}

	// No error
	return nil
}
