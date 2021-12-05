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
	"encoding/json"
	"fmt"
	"io"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"

	"github.com/elastic/harp/pkg/container/identity/key"
	"github.com/elastic/harp/pkg/sdk/types"
)

const (
	apiVersion = "harp.elastic.co/v1"
	kind       = "ContainerIdentity"
)

// -----------------------------------------------------------------------------

type PrivateKeyGeneratorFunc func(io.Reader) (*key.JSONWebKey, string, error)

// New identity from description.
func New(random io.Reader, description string, generator PrivateKeyGeneratorFunc) (*Identity, []byte, error) {
	// Check arguments
	if err := validation.Validate(description, validation.Required, is.ASCII); err != nil {
		return nil, nil, fmt.Errorf("unable to create identity with invalid description: %w", err)
	}

	// Delegate to generator
	jwk, encodedPub, err := generator(random)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate identity private key: %w", err)
	}

	// Encode JWK as json
	payload, err := json.Marshal(jwk)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to serialize identity keypair: %w", err)
	}

	// Prepae identity object
	id := &Identity{
		APIVersion:  apiVersion,
		Kind:        kind,
		Timestamp:   time.Now().UTC(),
		Description: description,
		Public:      encodedPub,
	}

	// Encode to json for signature
	protected, err := json.Marshal(id)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to serialize identity for signature: %w", err)
	}

	// Sign the protected data
	sig, err := jwk.Sign(protected)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to sign protected data: %w", err)
	}

	// Auto-assign the signature
	id.Signature = sig

	// Return unsealed identity
	return id, payload, nil
}

// FromReader extract identity instance from reader.
func FromReader(r io.Reader) (*Identity, error) {
	// Check arguments
	if types.IsNil(r) {
		return nil, fmt.Errorf("unable to read nil reader")
	}

	// Convert input as a map
	var input Identity
	if err := json.NewDecoder(r).Decode(&input); err != nil {
		return nil, fmt.Errorf("unable to decode input JSON: %w", err)
	}

	// Check component
	if input.Private == nil {
		return nil, fmt.Errorf("invalid identity: missing private component")
	}

	// Validate self signature
	if errVerify := input.Verify(); errVerify != nil {
		return nil, fmt.Errorf("unable to verify identity: %w", errVerify)
	}

	// Return no error
	return &input, nil
}
