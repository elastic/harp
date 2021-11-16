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
	"encoding/json"
	"fmt"

	"github.com/awnumar/memguard"

	"github.com/elastic/harp/pkg/container"
	"github.com/elastic/harp/pkg/tasks"
)

// SealTask implements secret container sealing task.
type SealTask struct {
	ContainerReader          tasks.ReaderProvider
	SealedContainerWriter    tasks.WriterProvider
	OutputWriter             tasks.WriterProvider
	Identities               []string
	DCKDMasterKey            *memguard.LockedBuffer
	DCKDTarget               string
	JSONOutput               bool
	DisableContainerIdentity bool
}

// Run the task.
func (t *SealTask) Run(ctx context.Context) error {
	// Create input reader
	reader, err := t.ContainerReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to open input reader: %w", err)
	}

	// Load input container
	in, err := container.Load(reader)
	if err != nil {
		return fmt.Errorf("unable to read input container: %w", err)
	}

	// If using sealing seed
	peerPublicKeys, err := container.PublicKeysFromIdentities(t.Identities...)
	if err != nil {
		return fmt.Errorf("unable to process identities: %w", err)
	}

	var containerKey string

	if !t.DisableContainerIdentity {
		opts := []container.GenerateOption{}
		// Enable deterministic generation
		if t.DCKDMasterKey != nil {
			opts = append(opts, container.WithDeterministicKey(t.DCKDMasterKey, t.DCKDTarget))
		}

		// Generate container key
		containerPublicKey, containerPrivateKey, errContainerGen := container.GenerateKey(opts...)
		if errContainerGen != nil {
			return fmt.Errorf("unable to generate container key: %w", errContainerGen)
		}

		// Append to identity
		peerPublicKeys = append(peerPublicKeys, containerPublicKey)
		containerKey = base64.RawURLEncoding.EncodeToString(containerPrivateKey[:])
	}

	// Seal the container
	sealedContainer, err := container.Seal(in, peerPublicKeys...)
	if err != nil {
		return fmt.Errorf("unable to seal container: %w", err)
	}

	// Open output file
	writer, err := t.SealedContainerWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to create output bundle: %w", err)
	}

	// Dump to writer
	if err = container.Dump(writer, sealedContainer); err != nil {
		return fmt.Errorf("unable to write sealed container: %w", err)
	}

	// Get output writer
	outputWriter, err := t.OutputWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve output writer: %w", err)
	}

	if containerKey != "" {
		// Display as json
		if t.JSONOutput {
			if err := json.NewEncoder(outputWriter).Encode(map[string]interface{}{
				"container_key": containerKey,
			}); err != nil {
				return fmt.Errorf("unable to display as json: %w", err)
			}
		} else {
			// Display container key
			fmt.Fprintf(outputWriter, "Container key : %s\n", containerKey)
		}
	}

	// No error
	return nil
}
