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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/imdario/mergo"
	"github.com/jmespath/go-jmespath"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle/secret"
	"github.com/elastic/harp/pkg/bundle/selector"
	"github.com/elastic/harp/pkg/template/engine"
)

type ruleAction uint

const (
	packageUnchanged ruleAction = iota
	packageUpdated
	packagedRemoved
)

// -----------------------------------------------------------------------------

func executeRule(r *bundlev1.PatchRule, p *bundlev1.Package, values map[string]interface{}) (ruleAction, error) {
	// Check parameters
	if r == nil {
		return packageUnchanged, fmt.Errorf("cannot process nil rule")
	}
	if r.Package == nil {
		return packageUnchanged, fmt.Errorf("cannot process rule with nil package")
	}
	if p == nil {
		return packageUnchanged, fmt.Errorf("cannot process nil package")
	}

	// Compile selector
	s, err := compileSelector(r.Selector, values)
	if err != nil {
		return packageUnchanged, fmt.Errorf("unable to compile selector: %w", err)
	}

	// Package match selector specification
	if s.IsSatisfiedBy(p) {
		// Check removal request
		if r.Package.Remove {
			return packagedRemoved, nil
		}

		// Apply patch
		if err := applyPackagePatch(p, r.Package, values); err != nil {
			return packageUnchanged, fmt.Errorf("unable to apply patch to package `%s`: %w", p.Name, err)
		}

		// No error
		return packageUpdated, nil
	}

	// No error
	return packageUnchanged, nil
}

//nolint:gocyclo,funlen // to refactor
func compileSelector(s *bundlev1.PatchSelector, values map[string]interface{}) (selector.Specification, error) {
	// Check parameters
	if s == nil {
		return nil, fmt.Errorf("cannot process nil selector")
	}

	// Has matchPath selector
	if s.MatchPath != nil {
		switch {
		case s.MatchPath.Strict != "":
			// Evaluation with template engine first
			value, err := engine.Render(s.MatchPath.Strict, map[string]interface{}{
				"Values": values,
			})
			if err != nil {
				return nil, fmt.Errorf("unable to evaluate template before matchPath build: %w", err)
			}

			// Return specification
			return selector.MatchPathStrict(value), nil
		case s.MatchPath.Glob != "":
			// Evaluation with template engine first
			value, err := engine.Render(s.MatchPath.Glob, map[string]interface{}{
				"Values": values,
			})
			if err != nil {
				return nil, fmt.Errorf("unable to evaluate template before matchPath build: %w", err)
			}

			// Return specification
			return selector.MatchPathGlob(value)
		case s.MatchPath.Regex != "":
			// Evaluation with template engine first
			value, err := engine.Render(s.MatchPath.Regex, map[string]interface{}{
				"Values": values,
			})
			if err != nil {
				return nil, fmt.Errorf("unable to evaluate template before matchPath build: %w", err)
			}

			// Return specification
			return selector.MatchPathRegex(value)
		default:
			return nil, errors.New("no strict, glob or regexp defined for path matcher")
		}
	}

	// Has jmesPath selector
	if s.JmesPath != "" {
		// Compile query
		exp, err := jmespath.Compile(s.JmesPath)
		if err != nil {
			return nil, fmt.Errorf("unable to compile jmesPath expression `%s`: %w", s.JmesPath, err)
		}

		// Return specification
		return selector.MatchJMESPath(exp), nil
	}

	// Has matchSecret selector
	if s.MatchSecret != nil {
		switch {
		case s.MatchSecret.Strict != "":
			// Evaluation with template engine first
			value, err := engine.Render(s.MatchSecret.Strict, map[string]interface{}{
				"Values": values,
			})
			if err != nil {
				return nil, fmt.Errorf("unable to evaluate template before matchSecret build: %w", err)
			}

			// Return specification
			return selector.MatchSecretStrict(value), nil
		case s.MatchSecret.Glob != "":
			// Evaluation with template engine first
			value, err := engine.Render(s.MatchSecret.Glob, map[string]interface{}{
				"Values": values,
			})
			if err != nil {
				return nil, fmt.Errorf("unable to evaluate template before matchSecret build: %w", err)
			}

			// Return specification
			return selector.MatchSecretGlob(value), nil
		case s.MatchSecret.Regex != "":
			// Evaluation with template engine first
			value, err := engine.Render(s.MatchSecret.Regex, map[string]interface{}{
				"Values": values,
			})
			if err != nil {
				return nil, fmt.Errorf("unable to evaluate template before matchSecret build: %w", err)
			}

			// Compile regexp
			re, err := regexp.Compile(value)
			if err != nil {
				return nil, fmt.Errorf("unable to compile matchSecret regexp `%s`: %w", s.MatchPath.Regex, err)
			}

			// Return specification
			return selector.MatchSecretRegex(re), nil
		default:
			return nil, errors.New("no strict, glob or regexp defined for secret matcher")
		}
	}

	if s.RegoFile != "" {
		// Read policy file
		policyFile, err := os.ReadFile(s.RegoFile)
		if err != nil {
			return nil, fmt.Errorf("unable to open rego policy file: %w", err)
		}

		// Build the specification
		return selector.MatchRego(context.Background(), string(policyFile))
	}

	// Has rego policy
	if s.Rego != "" {
		// Return specification
		return selector.MatchRego(context.Background(), s.Rego)
	}

	// Has CEL expressions
	if len(s.Cel) > 0 {
		// Return specification
		return selector.MatchCEL(s.Cel)
	}

	// Fallback to default as error
	return nil, fmt.Errorf("no supported selector specified")
}

func applyPackagePatch(pkg *bundlev1.Package, p *bundlev1.PatchPackage, values map[string]interface{}) error {
	// Check parameters
	if pkg == nil {
		return fmt.Errorf("cannot process nil package")
	}
	if p == nil {
		return fmt.Errorf("cannot process nil patch")
	}

	// Patch convcerns path
	if p.Path != nil {
		var err error
		if pkg.Name, err = applyPackagePathPatch(pkg.Name, p.Path, values); err != nil {
			return fmt.Errorf("unable to process `%s` name operations: %w", pkg.Name, err)
		}
	}

	// Patch concerns annotations
	if p.Annotations != nil {
		if pkg.Annotations == nil {
			pkg.Annotations = map[string]string{}
		}
		if err := applyMapOperations(pkg.Annotations, p.Annotations, values); err != nil {
			return fmt.Errorf("unable to process `%s` annotations: %w", pkg.Name, err)
		}
	}

	// Patch concerns labels
	if p.Labels != nil {
		if pkg.Labels == nil {
			pkg.Labels = map[string]string{}
		}
		if err := applyMapOperations(pkg.Labels, p.Labels, values); err != nil {
			return fmt.Errorf("unable to process `%s` labels: %w", pkg.Name, err)
		}
	}

	// Patch concerns data
	if p.Data != nil {
		if pkg.Secrets == nil {
			pkg.Secrets = &bundlev1.SecretChain{}
		}
		if err := applySecretPatch(pkg.Secrets, p.Data, values); err != nil {
			return fmt.Errorf("unable to apply patch to secret data for package `%s`: %w", pkg.Name, err)
		}
	}

	return nil
}

//nolint:gocyclo // to refactor
func applySecretPatch(secrets *bundlev1.SecretChain, op *bundlev1.PatchSecret, values map[string]interface{}) error {
	// Check parameters
	if secrets == nil {
		return fmt.Errorf("cannot process nil secrets")
	}
	if op == nil {
		return fmt.Errorf("cannot process nil patch")
	}

	// Patch concerns annotations
	if op.Annotations != nil {
		if secrets.Annotations == nil {
			secrets.Annotations = map[string]string{}
		}
		if err := applyMapOperations(secrets.Annotations, op.Annotations, values); err != nil {
			return fmt.Errorf("unable to process annotations: %w", err)
		}
	}

	// Patch concerns labels
	if op.Labels != nil {
		if secrets.Labels == nil {
			secrets.Labels = map[string]string{}
		}
		if err := applyMapOperations(secrets.Labels, op.Labels, values); err != nil {
			return fmt.Errorf("unable to process labels: %w", err)
		}
	}

	// Check template
	if op.Template != "" {
		// Compile template
		rendered, err := engine.Render(op.Template, map[string]interface{}{
			"Values": values,
		})
		if err != nil {
			return fmt.Errorf("unable to compile secret template: %w", err)
		}

		// Unmarshall as kv
		var kv map[string]string
		if errJSON := json.Unmarshal([]byte(rendered), &kv); errJSON != nil {
			return fmt.Errorf("unable to valudate rendered secret template as a valid JSON: %w", errJSON)
		}

		// Update secret data
		if secrets.Data == nil {
			secrets.Data = make([]*bundlev1.KV, 0)
		}
		if secrets.Data, err = updateSecret(secrets.Data, kv); err != nil {
			return fmt.Errorf("unable to uppdate kv from template: %w", err)
		}
	}

	// Check K/V
	if op.Kv != nil {
		var err error
		if secrets.Data == nil {
			secrets.Data = make([]*bundlev1.KV, 0)
		}
		if secrets.Data, err = applySecretKVPatch(secrets.Data, op.Kv, values); err != nil {
			return fmt.Errorf("unable to process kv: %w", err)
		}
	}

	// No error
	return nil
}

//nolint:gocyclo // to refactor
func applySecretKVPatch(kv []*bundlev1.KV, op *bundlev1.PatchOperation, values map[string]interface{}) ([]*bundlev1.KV, error) {
	// Check parameters
	if kv == nil {
		return nil, fmt.Errorf("cannot process nil kv list")
	}
	if op == nil {
		return nil, fmt.Errorf("cannot process nil operation")
	}

	var out []*bundlev1.KV

	// Remove all keys
	if len(op.RemoveKeys) > 0 {
		for _, rx := range op.RemoveKeys {
			re, err := regexp.Compile(rx)
			if err != nil {
				return nil, fmt.Errorf("unable to compile regexp for key deletion '%s': %w", rx, err)
			}

			// Add to remove if match one expression
			for _, k := range kv {
				if k == nil {
					continue
				}
				if re.MatchString(k.Key) {
					if op.Remove == nil {
						op.Remove = make([]string, 0)
					}
					op.Remove = append(op.Remove, k.Key)
				}
			}
		}
	}

	// Remove secret
	if len(op.Remove) > 0 {
		// Overwrite secret list
		out = removeSecret(kv, op.Remove)
	}

	// Add
	if op.Add != nil {
		inMap, err := precompileMap(op.Add, values)
		if err != nil {
			return nil, fmt.Errorf("unable to compile add map templates: %w", err)
		}
		if out, err = addSecret(kv, inMap); err != nil {
			return nil, fmt.Errorf("unable to add secret: %w", err)
		}
	}

	// Update
	if op.Update != nil {
		inMap, err := precompileMap(op.Update, values)
		if err != nil {
			return nil, fmt.Errorf("unable to compile update map templates: %w", err)
		}
		if out, err = updateSecret(kv, inMap); err != nil {
			return nil, fmt.Errorf("unable to update secret: %w", err)
		}
	}

	// Replace secrets
	if op.ReplaceKeys != nil {
		inMap, err := precompileMap(op.ReplaceKeys, values)
		if err != nil {
			return nil, fmt.Errorf("unable to compile replaceKeys map templates: %w", err)
		}

		// Replace keys
		out = replaceSecret(kv, inMap)
	}

	// No error
	return out, nil
}

func replaceSecret(input []*bundlev1.KV, replaceMap map[string]string) []*bundlev1.KV {
	for i, s := range input {
		// Ignore nil
		if s == nil {
			continue
		}

		// Check if the key should be replaced
		if newKey, ok := replaceMap[s.Key]; ok {
			s.Key = newKey
		}

		// Update the reference
		input[i] = s
	}

	//
	return input
}

func removeSecret(input []*bundlev1.KV, removeList []string) []*bundlev1.KV {
	out := []*bundlev1.KV{}

	for _, s := range input {
		// Ignore nil
		if s == nil {
			continue
		}

		found := false
		for _, toRemove := range removeList {
			if s.Key == toRemove {
				found = true
			}
		}
		// If not in list
		if !found {
			out = append(out, s)
		}
	}

	return out
}

func addSecret(input []*bundlev1.KV, newSecrets map[string]string) ([]*bundlev1.KV, error) {
	// Secret to add
	keys := []string{}
	out := []*bundlev1.KV{}

	// Check overrides
	for k := range newSecrets {
		found := false
		for _, s := range input {
			// Ignore nil
			if s == nil {
				continue
			}

			if s.Key == k {
				found = true
			}
		}
		// If not found
		if !found {
			keys = append(keys, k)
		}
	}

	// Add all existing secrets
	out = append(out, input...)

	// Add non-override key as new secret only
	for _, k := range keys {
		payload, err := secret.Pack(newSecrets[k])
		if err != nil {
			return nil, fmt.Errorf("unable to pack secret: %w", err)
		}

		out = append(out, &bundlev1.KV{
			Key:   k,
			Type:  fmt.Sprintf("%T", newSecrets[k]),
			Value: payload,
		})
	}

	// No error
	return out, nil
}

func updateSecret(input []*bundlev1.KV, newSecrets map[string]string) ([]*bundlev1.KV, error) {
	// Secret to add
	out := []*bundlev1.KV{}

	for _, s := range input {
		// Ignore nil
		if s == nil {
			continue
		}

		// Check if concerned by updates
		v, ok := newSecrets[s.Key]
		if !ok {
			// Append to result
			out = append(out, s)

			// Skip modification
			continue
		}

		// Update with new value
		payload, err := secret.Pack(v)
		if err != nil {
			return nil, fmt.Errorf("unable to pack secret: %w", err)
		}

		// Append to result
		out = append(out, &bundlev1.KV{
			Key:   s.Key,
			Type:  fmt.Sprintf("%T", v),
			Value: payload,
		})
	}

	// No error
	return out, nil
}

//nolint:gocyclo // to refactor
func applyMapOperations(input map[string]string, op *bundlev1.PatchOperation, values map[string]interface{}) error {
	// Check parameters
	if input == nil {
		return fmt.Errorf("cannot process nil map")
	}
	if op == nil {
		return fmt.Errorf("cannot process nil operation")
	}

	// Process all operations
	// Remove all keys
	if len(op.RemoveKeys) > 0 {
		for _, rx := range op.RemoveKeys {
			re, err := regexp.Compile(rx)
			if err != nil {
				return fmt.Errorf("unable to compile regexp for key deletion '%s': %w", rx, err)
			}

			// Add to remove if match one expression
			for k := range input {
				if re.MatchString(k) {
					if op.Remove == nil {
						op.Remove = make([]string, 0)
					}
					op.Remove = append(op.Remove, k)
				}
			}
		}
	}
	if len(op.Remove) > 0 {
		for _, toRemove := range op.Remove {
			delete(input, toRemove)
		}
	}
	if op.Add != nil {
		inMap, err := precompileMap(op.Add, values)
		if err != nil {
			return fmt.Errorf("unable to compile add map templates: %w", err)
		}
		if err := mergo.Merge(&input, inMap); err != nil {
			return fmt.Errorf("unable to add attributes to object: %w", err)
		}
	}
	if op.Update != nil {
		inMap, err := precompileMap(op.Update, values)
		if err != nil {
			return fmt.Errorf("unable to compile update map templates: %w", err)
		}
		if err := mergo.Merge(&input, inMap, mergo.WithOverride); err != nil {
			return fmt.Errorf("unable to add attributes to object: %w", err)
		}
	}
	if op.ReplaceKeys != nil {
		inMap, err := precompileMap(op.ReplaceKeys, values)
		if err != nil {
			return fmt.Errorf("unable to compile replaceKeys map templates: %w", err)
		}
		for oldKey, newKey := range inMap {
			if v, ok := input[oldKey]; ok {
				input[newKey] = v
				delete(input, oldKey)
			}
		}
	}

	return nil
}

func precompileMap(input map[string]string, values map[string]interface{}) (map[string]string, error) {
	output := map[string]string{}

	for k, v := range input {
		// Compile key
		key, err := engine.Render(k, map[string]interface{}{
			"Values": values,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to compile key template `%s`: %w", k, err)
		}

		// Compile value
		val, err := engine.Render(v, map[string]interface{}{
			"Values": values,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to compile value template `%s`: %w", v, err)
		}

		// Assign to result
		if _, ok := output[key]; !ok {
			output[key] = val
		}
	}

	// No error
	return output, nil
}

func applyPackagePathPatch(path string, op *bundlev1.PatchPackagePath, values map[string]interface{}) (string, error) {
	// Check parameters
	if op == nil {
		return "", fmt.Errorf("cannot process nil operation")
	}

	// Apply template transformation
	out, err := engine.Render(op.Template, map[string]interface{}{
		"Values": values,
		"Path":   path,
	})
	if err != nil {
		return "", fmt.Errorf("unable to execute package name template of `%s`: %w", path, err)
	}

	// No error
	return out, nil
}
