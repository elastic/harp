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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/harp/pkg/crate"
	"github.com/elastic/harp/pkg/crate/cratefile"
	"github.com/elastic/harp/pkg/tasks"
	"oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/target"
)

// PushTask implements secret-container publication process to and OCI compatible registry.
type PushTask struct {
	SpecReader   tasks.ReaderProvider
	OutputWriter tasks.WriterProvider
	Target       string
	Ref          string
	JSONOutput   bool
	RegistryOpts content.RegistryOptions
}

// Run the task.
func (t *PushTask) Run(ctx context.Context) error {
	// Create the reader
	reader, err := t.SpecReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to open input specification reader: %w", err)
	}

	// Create output writer
	writer, err := t.OutputWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to open writer: %w", err)
	}

	// Decode Cratefile
	spec, err := cratefile.Parse(reader, "", "hcl")
	if err != nil {
		return fmt.Errorf("unable to parse input cratefile: %w", err)
	}

	var (
		to target.Target
	)

	toParts := strings.SplitN(t.Target, ":", 2)
	// Build appropriate target instance
	switch toParts[0] {
	case "files":
		to = content.NewFile(toParts[1])
	case "registry":
		to, err = content.NewRegistry(t.RegistryOpts)
		if err != nil {
			return fmt.Errorf("could not create registry target: %w", err)
		}
	case "oci":
		to, err = content.NewOCI(toParts[1])
		if err != nil {
			return fmt.Errorf("could not read OCI layout at %s: %w", toParts[1], err)
		}
	default:
		return fmt.Errorf("unknown target argument: %s", t.Target)
	}

	// Prepare image from cratefile
	img, err := crate.Build(spec)
	if err != nil {
		return fmt.Errorf("unable to generate image descriptor from specification: %w", err)
	}

	// Push the container
	m, err := crate.Push(ctx, to, t.Ref, img)
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
