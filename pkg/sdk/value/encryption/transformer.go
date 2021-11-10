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

package encryption

import (
	"errors"
	"fmt"
	"strings"

	"github.com/elastic/harp/pkg/sdk/types"
	"github.com/elastic/harp/pkg/sdk/value"
	"github.com/elastic/harp/pkg/sdk/value/encryption/aead"
	"github.com/elastic/harp/pkg/sdk/value/encryption/fernet"
	"github.com/elastic/harp/pkg/sdk/value/encryption/secretbox"
)

const (
	secretboxPrefix  = "secretbox:"
	aesgcmPrefix     = "aes-gcm:"
	aespmacsivPrefix = "aes-pmac-siv:"
	aessivPrefix     = "aes-siv:"
	fernetPrefix     = "fernet:"
	chachaPrefix     = "chacha:"
	xchachaPrefix    = "xchacha:"
)

// FromKey returns the value transformer that match the value format.
func FromKey(keyValue string) (value.Transformer, error) {
	var (
		transformer value.Transformer
		err         error
	)

	// Check arguments
	if keyValue == "" {
		return nil, fmt.Errorf("unable to select a value transformer with blank value")
	}

	// Build the value transformer according to used prefix.
	switch {
	case strings.HasPrefix(keyValue, secretboxPrefix):
		// Activate Nacl SecretBox transformer
		transformer, err = secretbox.Transformer(strings.TrimPrefix(keyValue, secretboxPrefix))
	case strings.HasPrefix(keyValue, aesgcmPrefix):
		// Activate AES-GCM transformer
		transformer, err = aead.AESGCM(strings.TrimPrefix(keyValue, aesgcmPrefix))
	case strings.HasPrefix(keyValue, aessivPrefix):
		// Activate AES-SIV transformer
		transformer, err = aead.AESSIV(strings.TrimPrefix(keyValue, aessivPrefix))
	case strings.HasPrefix(keyValue, aespmacsivPrefix):
		// Activate AES-PMAC-SIV transformer
		transformer, err = aead.AESPMACSIV(strings.TrimPrefix(keyValue, aespmacsivPrefix))
	case strings.HasPrefix(keyValue, chachaPrefix):
		// Activate ChaCha20Poly1305 transformer
		transformer, err = aead.Chacha20Poly1305(strings.TrimPrefix(keyValue, chachaPrefix))
	case strings.HasPrefix(keyValue, xchachaPrefix):
		// Activate XChaCha20Poly1305 transformer
		transformer, err = aead.XChacha20Poly1305(strings.TrimPrefix(keyValue, xchachaPrefix))
	case strings.HasPrefix(keyValue, fernetPrefix):
		// Activate Fernet transformer
		transformer, err = fernet.Transformer(strings.TrimPrefix(keyValue, fernetPrefix))
	default:
		// Fallback to fernet
		transformer, err = fernet.Transformer(keyValue)
	}

	// Check transformer initialization error
	if transformer == nil || err != nil {
		return nil, fmt.Errorf("unable to initialize value transformer: %w", err)
	}

	// No error
	return transformer, nil
}

// Must is used to panic when a transformer initialization failed.
func Must(t value.Transformer, err error) value.Transformer {
	if err != nil {
		panic(err)
	}
	if types.IsNil(t) {
		panic(errors.New("transformer is nil with a nil error"))
	}

	return t
}
