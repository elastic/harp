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

package container

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/awnumar/memguard"

	"github.com/elastic/harp/pkg/container"
	"github.com/elastic/harp/pkg/tasks"
)

// UnsealTask implements secret container unsealing task.
type UnsealTask struct {
	ContainerReader tasks.ReaderProvider
	OutputWriter    tasks.WriterProvider
	ContainerKey    *memguard.LockedBuffer
}

// Run the task.
func (t *UnsealTask) Run(ctx context.Context) error {
	// Create input reader
	reader, err := t.ContainerReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to open input bundle reader: %w", err)
	}

	// Load input container
	in, err := container.Load(reader)
	if err != nil {
		return fmt.Errorf("unable to read input container: %v", err)
	}

	// Decode container key
	privateKeyRaw, err := base64.RawURLEncoding.DecodeString(t.ContainerKey.String())
	if err != nil {
		return fmt.Errorf("unable to decode container key: %w", err)
	}
	defer memguard.WipeBytes(privateKeyRaw)

	// Unseal the bundle
	out, err := container.Unseal(in, memguard.NewBufferFromBytes(privateKeyRaw))
	if err != nil {
		return fmt.Errorf("unable to unseal bundle content: %w", err)
	}

	// Create output writer
	writer, err := t.OutputWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to open output bundle: %w", err)
	}

	// Dump all content
	if err := container.Dump(writer, out); err != nil {
		return fmt.Errorf("unable to write unsealed container: %w", err)
	}

	// No error
	return nil
}
