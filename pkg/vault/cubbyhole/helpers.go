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

package cubbyhole

import (
	"fmt"
	"time"

	"github.com/dchest/uniuri"

	"github.com/elastic/harp/pkg/vault/logical"
)

// addToCubbyhole inserts the secrets in a cubbyhole and returns a response-wrapping token
func addToCubbyhole(v logical.Logical, mountPath, secret string) (string, error) {
	// Generate a path
	secretPath := fmt.Sprintf("%s/harp/%s/%d", mountPath, uniuri.NewLen(64), time.Now().UnixNano())

	// Insert the secret
	_, err := v.Write(secretPath, map[string]interface{}{
		"s": secret,
	})
	if err != nil {
		return "", fmt.Errorf("unable to write secret on vault: %w", err)
	}

	// Read again to get a wrapped token
	s, err := v.Read(secretPath)
	if err != nil {
		return "", fmt.Errorf("unable to read the secret : %w", err)
	}

	// Return wrapping token
	return s.WrapInfo.Token, nil
}

// unWrap unwraps the received token and returns the secret as a string
func unWrap(v logical.Logical, token string) (string, error) {
	// Unwrap the given token
	s, err := v.Unwrap(token)
	if err != nil {
		return "", err
	}

	// Check if result has "s" attribute
	secretRaw, ok := s.Data["s"]
	if !ok {
		return "", fmt.Errorf("the returned secret is not well formatted")
	}

	// Check if secret is a string
	secret, ok := secretRaw.(string)
	if !ok {
		return "", fmt.Errorf("the returned secret is not a string")
	}

	// Return data
	return secret, nil
}
