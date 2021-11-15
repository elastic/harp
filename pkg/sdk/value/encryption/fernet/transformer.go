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

package fernet

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/fernet/fernet-go"

	"github.com/elastic/harp/pkg/sdk/value"
	"github.com/elastic/harp/pkg/sdk/value/encryption"
)

func init() {
	encryption.Register("fernet", Transformer)
}

// Transformer returns a fernet encryption transformer
func Transformer(key string) (value.Transformer, error) {
	// Remove the prefix
	key = strings.TrimPrefix(key, "fernet:")

	// Check given keys
	k, err := fernet.DecodeKey(key)
	if err != nil {
		return nil, fmt.Errorf("fernet: unable to initialize fernet transformer: %w", err)
	}

	// Return decorator constructor
	return &fernetTransformer{
		key: k,
	}, nil
}

// -----------------------------------------------------------------------------

type fernetTransformer struct {
	key *fernet.Key
}

func (d *fernetTransformer) To(_ context.Context, input []byte) ([]byte, error) {
	// Encrypt value
	out, err := fernet.EncryptAndSign(input, d.key)
	if err != nil {
		return nil, fmt.Errorf("fernet: unable to transform input value: %w", err)
	}

	// No error
	return out, nil
}

func (d *fernetTransformer) From(_ context.Context, input []byte) ([]byte, error) {
	// Encrypt value
	out := fernet.VerifyAndDecrypt(input, 0, []*fernet.Key{d.key})
	if out == nil {
		return nil, errors.New("fernet: unable to decrypt value")
	}

	// No error
	return out, nil
}
