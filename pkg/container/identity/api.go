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

package identity

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/elastic/harp/pkg/container/identity/key"
	"github.com/elastic/harp/pkg/sdk/types"
	"github.com/elastic/harp/pkg/sdk/value"
)

// Identity object to hold container sealer identity information.
type Identity struct {
	APIVersion  string      `json:"@apiVersion"`
	Kind        string      `json:"@kind"`
	Timestamp   time.Time   `json:"@timestamp"`
	Description string      `json:"@description"`
	Public      string      `json:"public"`
	Private     *PrivateKey `json:"private"`
	Signature   string      `json:"signature"`
}

// HasPrivateKey returns true if identity as a wrapped private.
func (i *Identity) HasPrivateKey() bool {
	return i.Private != nil
}

// Decrypt private key with given transformer.
func (i *Identity) Decrypt(ctx context.Context, t value.Transformer) (*key.JSONWebKey, error) {
	// Check arguments
	if types.IsNil(t) {
		return nil, fmt.Errorf("can't process with nil transformer")
	}
	if !i.HasPrivateKey() {
		return nil, fmt.Errorf("trying to decrypt a nil private key")
	}

	// Decode payload
	payload, err := base64.RawURLEncoding.DecodeString(i.Private.Content)
	if err != nil {
		return nil, fmt.Errorf("unable to decode private key: %w", err)
	}

	// Apply transformation
	clearText, err := t.From(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("unable to decrypt identity payload: %w", err)
	}

	// Decode key
	var pk key.JSONWebKey
	if err = json.NewDecoder(bytes.NewReader(clearText)).Decode(&pk); err != nil {
		return nil, fmt.Errorf("unable to decode payload as JSON: %w", err)
	}

	// Return result
	return &pk, nil
}

// Verify the identity signature using its own public key.
func (i *Identity) Verify() error {
	// Clear the signature
	id := &Identity{}
	*id = *i

	// Clean protected
	id.Signature = ""
	id.Private = nil

	// Prepare protected
	protected, err := json.Marshal(id)
	if err != nil {
		return fmt.Errorf("unable to serialize identity for signature: %w", err)
	}

	// Decode the signature
	sig, err := base64.RawURLEncoding.DecodeString(i.Signature)
	if err != nil {
		return fmt.Errorf("unable to decode the signature: %w", err)
	}

	// Decode public key
	pubKey, err := key.FromString(id.Public)
	if err != nil {
		return fmt.Errorf("unable to decode public key: %w", err)
	}

	// Validate signature
	if pubKey.Verify(protected, sig) {
		return nil
	}

	return errors.New("unable to validate identity signature")
}

// PrivateKey wraps encoded private and related informations.
type PrivateKey struct {
	Encoding string `json:"encoding,omitempty"`
	Content  string `json:"content"`
}
