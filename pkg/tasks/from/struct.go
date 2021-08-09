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
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle"
	"github.com/elastic/harp/pkg/sdk/value/flatmap"
	"github.com/elastic/harp/pkg/tasks"
	"gopkg.in/yaml.v2"
)

// StructTask implements secret-container creation from a YAML/JSON structure.
type StructTask struct {
	StructReader tasks.ReaderProvider
	OutputWriter tasks.WriterProvider
}

// Run the task.
func (t *StructTask) Run(ctx context.Context) error {
	var (
		reader io.Reader
		writer io.Writer
		b      *bundlev1.Bundle
		err    error
	)

	// Create input reader
	reader, err = t.StructReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to read input reader: %w", err)
	}

	// Read input as struct
	in, err := ioutil.ReadAll(reader)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("unable to drain input reader: %w", err)
	}

	// Decode as YAML any object
	var source map[string]interface{}
	if errYaml := yaml.Unmarshal(in, &source); errYaml != nil {
		return fmt.Errorf("unable to decode source as YAML: %w", err)
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
