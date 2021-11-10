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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/elastic/harp/pkg/bundle"
	"github.com/elastic/harp/pkg/sdk/types"
	"github.com/elastic/harp/pkg/tasks"
)

// ReadTask implements secret container reading task.
type ReadTask struct {
	ContainerReader tasks.ReaderProvider
	OutputWriter    tasks.WriterProvider
	PackageName     string
	SecretKey       string
}

// Run the task.
func (t *ReadTask) Run(ctx context.Context) error {
	// Check arguments
	if types.IsNil(t.ContainerReader) {
		return errors.New("unable to run task with a nil containerReader provider")
	}
	if types.IsNil(t.OutputWriter) {
		return errors.New("unable to run task with a nil outputWriter provider")
	}
	if t.PackageName == "" {
		return errors.New("unable to proceed with blank packageName")
	}

	// Create input reader
	reader, err := t.ContainerReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to open input bundle: %w", err)
	}

	// Load bundle
	b, err := bundle.FromContainerReader(reader)
	if err != nil {
		return fmt.Errorf("unable to load bundle content: %w", err)
	}

	// Read a secret from bundle
	s, err := bundle.Read(b, t.PackageName)
	if err != nil {
		return fmt.Errorf("unable to read bundle content: %w", err)
	}

	// Prepare output writer
	writer, err := t.OutputWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to get output writer: %w", err)
	}

	if t.SecretKey != "" {
		if v, ok := s[t.SecretKey]; ok {
			fmt.Fprintf(writer, "%s", v)
		} else {
			return fmt.Errorf("requested field does not exist '%s': %w", t.SecretKey, err)
		}
	} else {
		// Dump the secret value
		if err := json.NewEncoder(writer).Encode(s); err != nil {
			return fmt.Errorf("unable to encode secret value as json: %w", err)
		}
	}

	// No error
	return nil
}
