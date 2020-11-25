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

package bundle

import (
	"context"
	"fmt"

	"github.com/elastic/harp/pkg/bundle"
	"github.com/elastic/harp/pkg/bundle/patch"
	"github.com/elastic/harp/pkg/tasks"
)

// PatchTask implements secret container patching task.
type PatchTask struct {
	PatchReader     tasks.ReaderProvider
	ContainerReader tasks.ReaderProvider
	OutputWriter    tasks.WriterProvider
	Values          map[string]interface{}
}

// Run the task.
func (t *PatchTask) Run(ctx context.Context) error {
	// Retrieve the patch reader
	patchReader, err := t.PatchReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve patch reader: %w", err)
	}

	// Parse the input specification
	spec, err := patch.YAML(patchReader)
	if err != nil {
		return fmt.Errorf("unable to parse patch file: %w", err)
	}

	// Retrieve the container reader
	containerReader, err := t.ContainerReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve patch reader: %w", err)
	}

	// Load bundle
	b, err := bundle.FromContainerReader(containerReader)
	if err != nil {
		return fmt.Errorf("unable to load bundle content: %w", err)
	}

	// Apply the patch speicification to generate an output bundle
	if err = patch.Apply(spec, b, t.Values); err != nil {
		return fmt.Errorf("unable to generate output bundle from patch: %w", err)
	}

	// Retrieve the container reader
	outputWriter, err := t.OutputWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve output writer: %w", err)
	}

	// Dump all content
	if err = bundle.ToContainerWriter(outputWriter, b); err != nil {
		return fmt.Errorf("unable to dump bundle content: %w", err)
	}

	// No error
	return nil
}
