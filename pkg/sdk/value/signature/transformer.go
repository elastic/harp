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

package signature

import (
	"errors"
	"fmt"
	"strings"

	"github.com/elastic/harp/pkg/sdk/types"
	"github.com/elastic/harp/pkg/sdk/value"
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

	// Extract prefix
	parts := strings.SplitN(keyValue, ":", 2)
	if len(parts) != 2 {
		// Fallback to fernet
		parts = []string{"fernet", keyValue}
	}

	// Clean prefix
	prefix := strings.ToLower(strings.TrimSpace(parts[0]))

	// Build the value transformer according to used prefix.
	tf, ok := registry[prefix]
	if !ok {
		return nil, fmt.Errorf("no transformer registered for '%s' as prefix", prefix)
	}

	// Build the transformer instance
	transformer, err = tf(keyValue)

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
