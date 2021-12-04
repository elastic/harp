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
	"errors"
	"fmt"

	"github.com/elastic/harp/pkg/container/identity"
	"github.com/elastic/harp/pkg/sdk/types"
	"github.com/elastic/harp/pkg/sdk/value"
	"github.com/elastic/harp/pkg/tasks"
)

// RecoverTask implements secret container identity recovery task.
type RecoverTask struct {
	JSONReader   tasks.ReaderProvider
	OutputWriter tasks.WriterProvider
	Transformer  value.Transformer
	JSONOutput   bool
}

// Run the task.
func (t *RecoverTask) Run(ctx context.Context) error {
	// Check arguments
	if types.IsNil(t.JSONReader) {
		return errors.New("unable to run task with a nil jsonReader provider")
	}
	if types.IsNil(t.OutputWriter) {
		return errors.New("unable to run task with a nil outputWriter provider")
	}
	if types.IsNil(t.Transformer) {
		return errors.New("unable to run task with a nil transformer")
	}

	// Create input reader
	reader, err := t.JSONReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to read input reader: %w", err)
	}

	// Extract from reader
	input, err := identity.FromReader(reader)
	if err != nil {
		return fmt.Errorf("unable to extract an identity from reader: %w", err)
	}

	// Try to decrypt the private key
	key, err := input.Decrypt(ctx, t.Transformer)
	if err != nil {
		return fmt.Errorf("unable to decrypt private key: %w", err)
	}

	// Retrieve recoevery key
	recoveryPrivateKey, err := identity.RecoveryKey(key)
	if err != nil {
		return fmt.Errorf("unable to retrieve recovery key from identity: %w", err)
	}

	// Get output writer
	outputWriter, err := t.OutputWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve output writer: %w", err)
	}

	// Display as json
	if t.JSONOutput {
		if errJSON := json.NewEncoder(outputWriter).Encode(map[string]interface{}{
			"container_key": base64.RawURLEncoding.EncodeToString(recoveryPrivateKey),
		}); errJSON != nil {
			return fmt.Errorf("unable to display as json: %w", errJSON)
		}
	} else {
		// Display container key
		if _, err := fmt.Fprintf(outputWriter, "Container key : %s\n", base64.RawURLEncoding.EncodeToString(recoveryPrivateKey)); err != nil {
			return fmt.Errorf("unable to display result: %w", err)
		}
	}

	// No error
	return nil
}
