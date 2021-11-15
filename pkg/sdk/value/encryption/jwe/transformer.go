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

package jwe

import (
	"context"
	"fmt"

	"github.com/elastic/harp/pkg/sdk/types"
	"github.com/elastic/harp/pkg/sdk/value"
	"gopkg.in/square/go-jose.v2"
)

// PBKDF2SaltSize is the default size of the salt for PBKDF2, 128-bit salt.
const PBKDF2SaltSize = 16

// PBKDF2Iterations is the default number of iterations for PBKDF2, 100k
// iterations. Nist recommends at least 10k, 1Passsword uses 100k.
const PBKDF2Iterations = 500001

// transformer returns a JWE encryption transformer
func transformer(key interface{}, keyAlgorithm jose.KeyAlgorithm, contentEncryption jose.ContentEncryption) (value.Transformer, error) {
	if types.IsNil(key) {
		return nil, fmt.Errorf("jwe: encryption key must not be nil")
	}

	// Return decorator constructor
	return &jweTransformer{
		key:               key,
		keyAlgorithm:      keyAlgorithm,
		contentEncryption: contentEncryption,
	}, nil
}

// -----------------------------------------------------------------------------

type jweTransformer struct {
	key               interface{}
	keyAlgorithm      jose.KeyAlgorithm
	contentEncryption jose.ContentEncryption
}

func (d *jweTransformer) To(_ context.Context, input []byte) ([]byte, error) {
	// Prepare JOSE recipient
	recipient := jose.Recipient{
		Algorithm:  d.keyAlgorithm,
		Key:        d.key,
		PBES2Count: PBKDF2Iterations,
	}

	// JWE Header
	opts := new(jose.EncrypterOptions)

	// Prepare encryption
	encrypter, err := jose.NewEncrypter(d.contentEncryption, recipient, opts)
	if err != nil {
		return nil, fmt.Errorf("jwe: unable to initialize encrypter: %w", err)
	}

	// Encrypt the input
	jwe, err := encrypter.Encrypt(input)
	if err != nil {
		return nil, fmt.Errorf("jwe: unable to encrypt identity keypair: %w", err)
	}

	// Assemble final JWE
	out, err := jwe.CompactSerialize()
	if err != nil {
		return nil, fmt.Errorf("jwe: unable to serialize encrypted payload: %w", err)
	}

	// No error
	return []byte(out), nil
}

func (d *jweTransformer) From(_ context.Context, input []byte) ([]byte, error) {
	// Parse JWE Token
	jwe, errParse := jose.ParseEncrypted(string(input))
	if errParse != nil {
		return nil, fmt.Errorf("unable to parse JWE token")
	}

	// Try to decrypt with given passphrase
	payload, errDecrypt := jwe.Decrypt(d.key)
	if errDecrypt != nil {
		return nil, fmt.Errorf("unable to decrypt JWE token")
	}

	// No error
	return payload, nil
}
