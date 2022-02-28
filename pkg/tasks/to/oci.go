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
	"errors"
	"fmt"

	"github.com/elastic/harp/pkg/container"
	"github.com/elastic/harp/pkg/container/oci"
	"github.com/elastic/harp/pkg/tasks"
)

// OCITask implements secret-container publication process to and OCI compatible registry.
type OCITask struct {
	ContainerReader tasks.ReaderProvider
	OutputWriter    tasks.WriterProvider
	Repository      string
	Path            string
	JSONOutput      bool
}

// Run the task.
func (t *OCITask) Run(ctx context.Context) error {
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

	// Read input container
	c, err := container.Load(reader)
	if err != nil {
		return fmt.Errorf("unable to load container: %w", err)
	}

	// Container must be sealed
	if !container.IsSealed(c) {
		return errors.New("the container must be sealed to be published in an OCI registry")
	}

	// Push the container
	m, err := oci.Push(ctx, c, t.Repository, t.Path)
	if err != nil {
		return fmt.Errorf("unable to push contain,er to registry: %w", err)
	}

	if t.JSONOutput {
		if err := json.NewEncoder(writer).Encode(m); err != nil {
			return fmt.Errorf("unbale to encode manifest as JSON: %w", err)
		}
	} else {
		fmt.Fprintf(writer, "Container successfully pushed !")
		fmt.Fprintf(writer, "Digest: %x", m.Digest)
		fmt.Fprintf(writer, "Size: %d", m.Size)
	}

	// No error
	return nil
}

// -----------------------------------------------------------------------------
