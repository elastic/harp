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
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/awnumar/memguard"

	"github.com/elastic/harp/pkg/container"
	"github.com/elastic/harp/pkg/container/seal"
	sealv1 "github.com/elastic/harp/pkg/container/seal/v1"
	sealv2 "github.com/elastic/harp/pkg/container/seal/v2"
	"github.com/elastic/harp/pkg/sdk/types"
	"github.com/elastic/harp/pkg/tasks"
)

// SealTask implements secret container sealing task.
type SealTask struct {
	ContainerReader          tasks.ReaderProvider
	SealedContainerWriter    tasks.WriterProvider
	OutputWriter             tasks.WriterProvider
	PeerPublicKeys           []string
	DCKDMasterKey            string
	DCKDTarget               string
	JSONOutput               bool
	DisableContainerIdentity bool
	SealVersion              uint
}

// Run the task.
//nolint:funlen,gocyclo // to refactor
func (t *SealTask) Run(ctx context.Context) error {
	// Check arguments
	if types.IsNil(t.ContainerReader) {
		return errors.New("unable to run task with a nil containerReader provider")
	}
	if types.IsNil(t.SealedContainerWriter) {
		return errors.New("unable to run task with a nil sealedContainerWriter provider")
	}
	if types.IsNil(t.OutputWriter) {
		return errors.New("unable to run task with a nil outputWriter provider")
	}
	if len(t.PeerPublicKeys) == 0 {
		return errors.New("at least one public key must be provided for recovery")
	}

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

	var containerKey string
	if !t.DisableContainerIdentity {
		opts := []seal.GenerateOption{}

		// Check container sealing master key usage
		if t.DCKDMasterKey != "" {
			// Process target
			if t.DCKDTarget == "" {
				return errors.New("target flag (string) is mandatory for key derivation")
			}

			// Decode master key
			masterKeyRaw, errDecode := base64.RawURLEncoding.DecodeString(t.DCKDMasterKey)
			if errDecode != nil {
				return fmt.Errorf("unable to decode master key: %w", errDecode)
			}

			// Enable deterministic container key generation
			opts = append(opts, seal.WithDeterministicKey(memguard.NewBufferFromBytes(masterKeyRaw), t.DCKDTarget))
		}

		// Initialize seal strategy
		var ss seal.Strategy
		switch t.SealVersion {
		case 1:
			ss = sealv1.New()
		case 2:
			ss = sealv2.New()
		default:
			ss = sealv1.New()
		}

		// Generate container key
		containerPublicKey, containerSecretKey, errGenerate := ss.GenerateKey(opts...)
		if errGenerate != nil {
			return fmt.Errorf("unable to generate container key: %w", errGenerate)
		}

		// Append to identities
		t.PeerPublicKeys = append(t.PeerPublicKeys, containerPublicKey)

		// Assign container key
		containerKey = containerSecretKey
	}

	// Seal the container
	sealedContainer, err := container.Seal(rand.Reader, in, t.PeerPublicKeys...)
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

	if !t.DisableContainerIdentity {
		// Get output writer
		outputWriter, err := t.OutputWriter(ctx)
		if err != nil {
			return fmt.Errorf("unable to retrieve output writer: %w", err)
		}

		// Display as json
		if t.JSONOutput {
			if err := json.NewEncoder(outputWriter).Encode(map[string]interface{}{
				"container_key": containerKey,
			}); err != nil {
				return fmt.Errorf("unable to display as json: %w", err)
			}
		} else {
			// Display container key
			if _, err := fmt.Fprintf(outputWriter, "Container key : %s\n", containerKey); err != nil {
				return fmt.Errorf("unable to display result: %w", err)
			}
		}
	}

	// No error
	return nil
}
