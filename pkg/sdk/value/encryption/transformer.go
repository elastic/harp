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
	"fmt"
	"strings"

	"github.com/elastic/harp/pkg/sdk/value"
	"github.com/elastic/harp/pkg/sdk/value/encryption/aes"
	"github.com/elastic/harp/pkg/sdk/value/encryption/fernet"
	"github.com/elastic/harp/pkg/sdk/value/encryption/secretbox"
)

const (
	secretboxPrefix = "secretbox:"
	aesgcmPrefix    = "aes-gcm:"
	fernetPrefix    = "fernet:"
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
	if strings.HasPrefix(keyValue, secretboxPrefix) {
		// Activate Nacl SecretBox transformer
		transformer, err = secretbox.Transformer(strings.TrimPrefix(keyValue, secretboxPrefix))
	} else if strings.HasPrefix(keyValue, aesgcmPrefix) {
		// Activate AES-GCM transformer
		transformer, err = aes.Transformer(strings.TrimPrefix(keyValue, aesgcmPrefix))
	} else if strings.HasPrefix(keyValue, fernetPrefix) {
		// Activate Fernet transformer
		transformer, err = fernet.Transformer(strings.TrimPrefix(keyValue, fernetPrefix))
	} else {
		// Fallback to fernet
		transformer, err = fernet.Transformer(keyValue)
	}

	// Check transformer initialization error
	if transformer == nil || err != nil {
		return nil, fmt.Errorf("unable to initialize value transformer: %v", err)
	}

	// No error
	return transformer, nil
}
