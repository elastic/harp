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

package paseto

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"errors"
	"fmt"

	pasetov3 "zntr.io/paseto/v3"
	pasetov4 "zntr.io/paseto/v4"

	"github.com/elastic/harp/pkg/sdk/types"
)

type pasetoTransformer struct {
	key interface{}
}

// -----------------------------------------------------------------------------

func (d *pasetoTransformer) To(_ context.Context, input []byte) ([]byte, error) {
	if types.IsNil(d.key) {
		return nil, fmt.Errorf("paseto: signer key must not be nil")
	}

	var (
		out     []byte
		payload string
		err     error
		f, s    []byte
	)

	switch sk := d.key.(type) {
	case ed25519.PrivateKey:
		payload, err = pasetov4.Sign(input, sk, f, s)
	case *ecdsa.PrivateKey:
		payload, err = pasetov3.Sign(input, sk, f, s)
	default:
		return nil, errors.New("paseto: key is not supported")
	}
	if err != nil {
		return nil, fmt.Errorf("paseto: unable so sign input: %w", err)
	}

	out = []byte(payload)

	// No error
	return out, nil
}

func (d *pasetoTransformer) From(_ context.Context, input []byte) ([]byte, error) {
	var (
		payload []byte
		err     error
		f, i    []byte
	)

	switch sk := d.key.(type) {
	case ed25519.PublicKey:
		payload, err = pasetov4.Verify(string(input), sk, f, i)
	case *ecdsa.PublicKey:
		payload, err = pasetov3.Verify(string(input), sk, f, i)
	default:
		return nil, errors.New("paseto: key is not supported")
	}
	if err != nil {
		return nil, fmt.Errorf("paseto: unable so sign input: %w", err)
	}

	// No error
	return payload, nil
}
