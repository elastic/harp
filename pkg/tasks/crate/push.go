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
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"oras.land/oras-go/pkg/content"

	"github.com/elastic/harp/pkg/crate"
	"github.com/elastic/harp/pkg/crate/cratefile"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/tasks"
)

// PushTask implements secret-container publication process to and OCI compatible registry.
type PushTask struct {
	SpecReader   tasks.ReaderProvider
	OutputWriter tasks.WriterProvider
	ContextPath  string
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

	// Prepare target resolver
	to, err := getResolver(t.Target, t.RegistryOpts)
	if err != nil {
		return fmt.Errorf("unable to initialize target: %w", err)
	}

	// Get absolute contxt path
	absContextPath, err := filepath.Abs(t.ContextPath)
	if err != nil {
		return fmt.Errorf("unable to get absolute context path: %w", err)
	}

	// Ensure the root actually exists
	fi, err := os.Stat(absContextPath)
	if err != nil {
		return fmt.Errorf("unable to check context path: %w", err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("context path '%s' must be a directory", absContextPath)
	}

	log.For(ctx).Info("Building image from context ...", zap.String("context", absContextPath))

	// Prepare image from cratefile
	img, err := crate.Build(os.DirFS(absContextPath), spec)
	if err != nil {
		return fmt.Errorf("unable to generate image descriptor from specification: %w", err)
	}

	// Push the container
	m, err := crate.Push(ctx, to, t.Ref, img)
	if err != nil {
		return fmt.Errorf("unable to push crate to registry: %w", err)
	}

	if t.JSONOutput {
		if err := json.NewEncoder(writer).Encode(m); err != nil {
			return fmt.Errorf("unable to encode manifest as JSON: %w", err)
		}
	} else {
		fmt.Fprintf(writer, "Container successfully pushed !\n")
		fmt.Fprintf(writer, "Digest: %s\n", m.Digest.Hex())
		fmt.Fprintf(writer, "Size: %d\n", m.Size)
	}

	// No error
	return nil
}

// -----------------------------------------------------------------------------
