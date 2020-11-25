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
	"github.com/elastic/harp/pkg/tasks"
)

// DiffTask implements secret container difference task.
type DiffTask struct {
	SourceReader      tasks.ReaderProvider
	DestinationReader tasks.ReaderProvider
	OutputWriter      tasks.WriterProvider
}

// Run the task.
func (t *DiffTask) Run(ctx context.Context) error {
	// Create input reader
	readerSrc, err := t.SourceReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to open source bundle: %w", err)
	}

	// Load source bundle
	bSrc, err := bundle.FromContainerReader(readerSrc)
	if err != nil {
		return fmt.Errorf("unable to load source bundle content: %w", err)
	}

	// Create input reader
	readerDst, err := t.DestinationReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to open destination bundle: %w", err)
	}

	// Load source bundle
	bDst, err := bundle.Load(readerDst)
	if err != nil {
		return fmt.Errorf("unable to load destination bundle content: %w", err)
	}

	// Calculate diff
	report, err := bundle.Diff(bSrc, bDst)
	if err != nil {
		return fmt.Errorf("unable to calculate bundle difference: %w", err)
	}

	// Create output writer
	writer, err := t.OutputWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to open output writer: %w", err)
	}

	// Print report
	fmt.Fprintln(writer, report)

	// No error
	return nil
}
