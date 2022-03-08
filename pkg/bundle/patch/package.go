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

package patch

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strings"

	"golang.org/x/crypto/blake2b"

	"google.golang.org/protobuf/proto"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle"
	"github.com/elastic/harp/pkg/sdk/types"
)

// Validate bundle patch.
func Validate(spec *bundlev1.Patch) error {
	// Check if spec is nil
	if spec == nil {
		return fmt.Errorf("unable to validate bundle patch: patch is nil")
	}

	if spec.ApiVersion != "harp.elastic.co/v1" {
		return fmt.Errorf("apiVersion should be 'harp.elastic.co/v1'")
	}

	if spec.Kind != "BundlePatch" {
		return fmt.Errorf("kind should be 'BundlePatch'")
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

// Checksum calculates the bundle patch checksum.
func Checksum(spec *bundlev1.Patch) (string, error) {
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

// Apply given patch to the given bundle.
//nolint:interfacer,gocyclo,funlen // Explicit type restriction
func Apply(spec *bundlev1.Patch, b *bundlev1.Bundle, values map[string]interface{}, o ...OptionFunc) (*bundlev1.Bundle, error) {
	// Validate spec
	if err := Validate(spec); err != nil {
		return nil, fmt.Errorf("unable to validate spec: %w", err)
	}
	if b == nil {
		return nil, fmt.Errorf("cannot process nil bundle")
	}

	// Prepare selectors
	if len(spec.Spec.Rules) == 0 {
		return nil, fmt.Errorf("empty bundle patch")
	}

	// Copy bundle
	bCopy, ok := proto.Clone(b).(*bundlev1.Bundle)
	if !ok {
		return nil, fmt.Errorf("the cloned bundle does not have the expected type: %T", bCopy)
	}
	if bCopy.Packages == nil {
		bCopy.Packages = []*bundlev1.Package{}
	}

	// Default evaluation options
	dopts := &options{
		stopAtRuleID:      "",
		stopAtRuleIndex:   -1,
		ignoreRuleIDs:     []string{},
		ignoreRuleIndexes: []int{},
	}

	// Apply functions
	for _, opt := range o {
		opt(dopts)
	}

	// Process all creation rule first
	for i, r := range spec.Spec.Rules {
		// Ignore nil rule
		if r == nil {
			continue
		}

		// Ignore non creation rules and non strict matcher
		if !r.Package.Create || r.Selector.MatchPath.Strict == "" {
			continue
		}
		if shouldIgnoreThisRule(i, r.Id, dopts) {
			continue
		}
		if shouldStopAtThisRule(i, r.Id, dopts) {
			break
		}

		// Create a package
		p := &bundlev1.Package{
			Name: r.Selector.MatchPath.Strict,
		}

		_, err := executeRule(r, p, values)
		if err != nil {
			return nil, fmt.Errorf("unable to execute rule index %d: %w", i, err)
		}

		// Add created package
		bCopy.Packages = append(bCopy.Packages, p)
	}

	for ri, r := range spec.Spec.Rules {
		// Ignore nil rule
		if r == nil {
			continue
		}
		if shouldIgnoreThisRule(ri, r.Id, dopts) {
			continue
		}
		if shouldStopAtThisRule(ri, r.Id, dopts) {
			break
		}

		// Process all packages
		for i, p := range bCopy.Packages {
			action, err := executeRule(r, p, values)
			if err != nil {
				return nil, fmt.Errorf("unable to execute rule index %d: %w", ri, err)
			}

			switch action {
			case packagedRemoved:
				bCopy.Packages = append(bCopy.Packages[:i], bCopy.Packages[i+1:]...)
			case packageUpdated:
				if WithAnnotations(spec) {
					// Add annotations to mark package as patched.
					bundle.Annotate(p, "patched", "true")
					bundle.Annotate(p, spec.Meta.Name, "true")
				}
				bCopy.Packages[i] = p
			case packageUnchanged:
				// No changes
			default:
			}
		}
	}

	// Sort packages
	sort.SliceStable(bCopy.Packages, func(i, j int) bool {
		return bCopy.Packages[i].Name < bCopy.Packages[j].Name
	})

	// No error
	return bCopy, nil
}

func shouldStopAtThisRule(idx int, id string, opts *options) bool {
	// Stop at index
	if opts.stopAtRuleIndex > 0 && idx >= opts.stopAtRuleIndex {
		return true
	}
	// Stop at rule id
	if opts.stopAtRuleID != "" && strings.EqualFold(id, opts.stopAtRuleID) {
		return true
	}

	return false
}

func shouldIgnoreThisRule(idx int, id string, opts *options) bool {
	// Ignore using index
	if len(opts.ignoreRuleIndexes) > 0 {
		for _, v := range opts.ignoreRuleIndexes {
			if v == idx {
				return true
			}
		}
	}

	// Ignore using id
	if len(opts.ignoreRuleIDs) > 0 {
		return types.StringArray(opts.ignoreRuleIDs).Contains(id)
	}

	return false
}
