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

package ruleset

import (
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/blake2b"

	"google.golang.org/protobuf/proto"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
)

// Validate bundle patch.
func Validate(spec *bundlev1.RuleSet) error {
	// Check if spec is nil
	if spec == nil {
		return fmt.Errorf("unable to validate bundle patch: patch is nil")
	}

	if spec.ApiVersion != "harp.elastic.co/v1" {
		return fmt.Errorf("apiVersion should be 'harp.elastic.co/v1'")
	}

	if spec.Kind != "RuleSet" {
		return fmt.Errorf("kind should be 'RuleSet'")
	}

	if spec.Meta == nil {
		return fmt.Errorf("meta should be 'nil'")
	}

	if spec.Spec == nil {
		return fmt.Errorf("spec should be 'nil'")
	}

	// No error
	return nil
}

// Checksum calculates the specification checksum.
func Checksum(spec *bundlev1.RuleSet) (string, error) {
	// Validate bundle template
	if err := Validate(spec); err != nil {
		return "", fmt.Errorf("unable to validate spec: %w", err)
	}

	// Encode spec as protobuf
	payload, err := proto.Marshal(spec)
	if err != nil {
		return "", fmt.Errorf("unable to encode bundle patch: %w", err)
	}

	// Calculate checksum
	checksum := blake2b.Sum256(payload)

	// No error
	return base64.RawURLEncoding.EncodeToString(checksum[:]), nil
}
