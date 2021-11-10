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
	"errors"
	"fmt"
	"path"

	"github.com/elastic/harp/pkg/bundle"
	"github.com/elastic/harp/pkg/sdk/types"
	"github.com/elastic/harp/pkg/tasks"
)

// PrefixerTask implements secret container prefix management task.
type PrefixerTask struct {
	ContainerReader tasks.ReaderProvider
	OutputWriter    tasks.WriterProvider
	Prefix          string
}

// Run the task.
func (t *PrefixerTask) Run(ctx context.Context) error {
	// Check arguments
	if types.IsNil(t.ContainerReader) {
		return errors.New("unable to run task with a nil containerReader provider")
	}
	if types.IsNil(t.OutputWriter) {
		return errors.New("unable to run task with a nil outputWriter provider")
	}
	if t.Prefix == "" {
		return errors.New("unable to proceed with blank prefix")
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

	// Iterate over all packages
	for _, p := range b.Packages {
		p.Name = path.Clean(path.Join(t.Prefix, p.Name))
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
