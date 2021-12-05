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

package jws

import (
	"context"
	"fmt"

	"gopkg.in/square/go-jose.v2"

	"github.com/elastic/harp/pkg/sdk/types"
)

type jwsTransformer struct {
	key  jose.SigningKey
	opts *jose.SignerOptions
}

// -----------------------------------------------------------------------------

func (d *jwsTransformer) To(_ context.Context, input []byte) ([]byte, error) {
	if types.IsNil(d.key.Key) {
		return nil, fmt.Errorf("jws: signer key must not be nil")
	}

	// Initialize a signer
	signer, err := jose.NewSigner(d.key, d.opts)
	if err != nil {
		return nil, fmt.Errorf("jws: unable to initialize a signer: %w", err)
	}

	// Sign input
	sig, err := signer.Sign(input)
	if err != nil {
		return nil, fmt.Errorf("jws: unable to sign the content: %w", err)
	}

	// Serialize content
	out, err := sig.CompactSerialize()
	if err != nil {
		return nil, fmt.Errorf("jws: unable to serialize final payload: %w", err)
	}

	// No error
	return []byte(out), nil
}

func (d *jwsTransformer) From(_ context.Context, input []byte) ([]byte, error) {
	// Parse the signed object
	sig, err := jose.ParseSigned(string(input))
	if err != nil {
		return nil, fmt.Errorf("jws: unable to parse input: %w", err)
	}

	// Verify signature
	payload, err := sig.Verify(d.key.Key)
	if err != nil {
		return nil, fmt.Errorf("jws: unable to valid signature: %w", err)
	}

	// No error
	return payload, nil
}
