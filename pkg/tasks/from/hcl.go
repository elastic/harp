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
	"fmt"
	"io"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle"
	"github.com/elastic/harp/pkg/bundle/hcl"
	"github.com/elastic/harp/pkg/tasks"
)

// JSONMapTask implements secret-container creation from JSON Map.
type HCLTask struct {
	HCLReader    tasks.ReaderProvider
	OutputWriter tasks.WriterProvider
}

// Run the task.
func (t *HCLTask) Run(ctx context.Context) error {
	var (
		reader io.Reader
		writer io.Writer
		b      *bundlev1.Bundle
		err    error
	)

	// Create input reader
	reader, err = t.HCLReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to read input reader: %w", err)
	}

	// Parse input as HCL configuration object.
	cfg, err := hcl.Parse(reader, "input", "hcl")
	if err != nil {
		return fmt.Errorf("unable to parse input HCL: %w", err)
	}

	// Build the container from hcl dsl
	b, err = bundle.FromHCL(cfg)
	if err != nil {
		return fmt.Errorf("unable to create container from hcl: %w", err)
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
