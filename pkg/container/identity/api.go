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
	"fmt"
	"time"

	"github.com/elastic/harp/pkg/sdk/security"
	"github.com/elastic/harp/pkg/sdk/security/crypto/bech32"
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
}

// HasPrivateKey returns true if identity as a wrapped private.
func (i *Identity) HasPrivateKey() bool {
	return i.Private != nil
}

// Decrypt private key with given transformer.
func (i *Identity) Decrypt(ctx context.Context, t value.Transformer) (*JSONWebKey, error) {
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
	var key JSONWebKey
	if err = json.NewDecoder(bytes.NewReader(clearText)).Decode(&key); err != nil {
		return nil, fmt.Errorf("unable to decode payload as JSON: %w", err)
	}

	// Build public key
	_, pubKey, err := bech32.Decode(i.Public)
	if err != nil {
		return nil, fmt.Errorf("invalid public key encoding: %w", err)
	}

	// Decode base64 public key
	pubKeyRaw, err := base64.RawURLEncoding.DecodeString(key.X)
	if err != nil {
		return nil, fmt.Errorf("invalid public key, the decoded public is corrupted")
	}

	// Check validity
	if !security.SecureCompare(pubKey, pubKeyRaw) {
		return nil, fmt.Errorf("invalid identity, key mismatch detected")
	}

	// Return result
	return &key, nil
}

// PrivateKey wraps encoded private and related informations.
type PrivateKey struct {
	Encoding string `json:"encoding,omitempty"`
	Content  string `json:"content"`
}

// JSONWebKey holds internal container key attributes.
type JSONWebKey struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	D   string `json:"d"`
}
