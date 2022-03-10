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

package crate

import (
	"context"
	"fmt"

	"github.com/elastic/harp/pkg/tasks"
	"github.com/elastic/harp/pkg/template/archive/vfs"
)

// ExtractArchiveTask implements archive extraction task.
type ExtractArchiveTask struct {
	ArchiveReader tasks.ReaderProvider
	OutputPath    string
}

// Run the task.
func (t *ExtractArchiveTask) Run(ctx context.Context) error {
	// Create the reader
	reader, err := t.ArchiveReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to open input reader: %w", err)
	}

	// Create virtual filesystem from input reader.
	fs, err := vfs.FromArchive(reader)
	if err != nil {
		return fmt.Errorf("unable to create archive filesystem: %w", err)
	}

	// Dump the filesystem to disk
	if err := vfs.Dump(fs, t.OutputPath); err != nil {
		return fmt.Errorf("unable to dump archive to disk: %w", err)
	}

	// No error
	return nil
}

// -----------------------------------------------------------------------------
