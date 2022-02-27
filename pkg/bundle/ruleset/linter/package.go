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

package linter

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gobwas/glob"
	"golang.org/x/crypto/blake2b"
	"google.golang.org/protobuf/proto"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle/ruleset/linter/engine"
	"github.com/elastic/harp/pkg/bundle/ruleset/linter/engine/cel"
	"github.com/elastic/harp/pkg/bundle/ruleset/linter/engine/rego"
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

// Evaluate given bundl using the loaded ruleset.
//nolint:gocyclo // to refactor
func Evaluate(ctx context.Context, b *bundlev1.Bundle, spec *bundlev1.RuleSet) error {
	// Validate spec
	if err := Validate(spec); err != nil {
		return fmt.Errorf("unable to validate spec: %w", err)
	}
	if b == nil {
		return fmt.Errorf("cannot process nil bundle")
	}

	// Prepare selectors
	if len(spec.Spec.Rules) == 0 {
		return fmt.Errorf("empty ruleset")
	}

	// Process each rule
	for _, r := range spec.Spec.Rules {
		// Complie path matcher
		pathMatcher, err := glob.Compile(r.Path)
		if err != nil {
			return fmt.Errorf("unable to compile path matcher: %w", err)
		}

		var (
			vm    engine.PackageLinter
			vmErr error
		)

		switch {
		case len(r.Constraints) > 0:
			// Compile constraints
			vm, vmErr = cel.New(r.Constraints)
		case r.RegoFile != "":
			// Open policy file
			f, err := os.Open(r.RegoFile)
			if err != nil {
				return fmt.Errorf("unable to open rego policy file: %w", err)
			}

			// Create a evaluation context
			vm, vmErr = rego.New(ctx, f)
		case r.Rego != "":
			// Create a evaluation context
			vm, vmErr = rego.New(ctx, strings.NewReader(r.Rego))
		default:
			return errors.New("one of 'constraints', 'rego' or 'rego_file' property must be defined")
		}
		if vmErr != nil {
			return fmt.Errorf("unable to prepare evaluation context: %w", vmErr)
		}

		// A rule must match at least one time.
		matchOnce := false

		// For each package
		for _, p := range b.Packages {
			if p == nil {
				// Ignore nil package
				continue
			}

			// If package match the path filter.
			if pathMatcher.Match(p.Name) {
				matchOnce = true

				errEval := vm.EvaluatePackage(ctx, p)
				if errEval != nil {
					if errors.Is(errEval, engine.ErrRuleNotValid) {
						return fmt.Errorf("package '%s' doesn't validate rule '%s'", p.Name, r.Name)
					}
					return fmt.Errorf("unexpected error occurred during constraints evaluation: %w", errEval)
				}
			}
		}

		// Check matching constraint
		if !matchOnce {
			return fmt.Errorf("rule '%s' didn't match any packages", r.Name)
		}
	}

	// No error
	return nil
}
