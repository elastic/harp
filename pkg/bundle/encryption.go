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

package bundle

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/awnumar/memguard"
	"github.com/golang/protobuf/ptypes/wrappers"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle/secret"
	"github.com/elastic/harp/pkg/sdk/types"
	"github.com/elastic/harp/pkg/sdk/value"
)

// PartialLock apply conditional transformer according to applicable annotation
// on the given package.
// The annotation is referring to a key alias provided.
func PartialLock(ctx context.Context, b *bundlev1.Bundle, transformerMap map[string]value.Transformer, skipUnresolved bool) error {
	// Check bundle
	if b == nil {
		return fmt.Errorf("unable to process nil bundle")
	}
	if transformerMap == nil {
		return fmt.Errorf("unable to process nil transformer map")
	}

	// For each packages
	for _, p := range b.Packages {
		// Check annotation usage
		keyAlias, hasKeyAlias := p.Annotations[packageEncryptionAnnotation]
		if !hasKeyAlias {
			// Skip package processing
			continue
		}

		// Check key alias declaration
		transformer, hasTransformer := transformerMap[keyAlias]
		if !hasTransformer {
			if skipUnresolved {
				// Skip unresolved transformer alias.
				continue
			}
			return fmt.Errorf("package encryption annotation found, but no key alias for '%s' provided", keyAlias)
		}
		if types.IsNil(transformer) {
			return fmt.Errorf("key alias '%s' refers to a nil transformer", keyAlias)
		}

		// Convert secret as a map
		secrets := map[string]interface{}{}
		for _, s := range p.Secrets.Data {
			var out interface{}
			if err := secret.Unpack(s.Value, &out); err != nil {
				return fmt.Errorf("unable to load secret value, corrupted bundle: %w", err)
			}

			// Assign to secret map
			secrets[s.Key] = out
		}

		// Export secrets as JSON
		content, err := json.Marshal(secrets)
		if err != nil {
			return fmt.Errorf("unable to extract secret map as json")
		}

		// Apply transformer
		out, err := transformer.To(ctx, content)
		if err != nil {
			return fmt.Errorf("unable to apply secret transformer: %w", err)
		}

		// Cleanup
		memguard.WipeBytes(content)
		p.Secrets.Data = nil

		// Assign locked secret
		p.Secrets.Locked = &wrappers.BytesValue{
			Value: out,
		}
	}

	// No error
	return nil
}

// Lock apply transformer function to all secret values and set as locked.
func Lock(ctx context.Context, b *bundlev1.Bundle, transformer value.Transformer) error {
	// Check bundle
	if b == nil {
		return fmt.Errorf("unable to process nil bundle")
	}
	if types.IsNil(transformer) {
		return fmt.Errorf("unable to process nil transformer")
	}

	// For each packages
	for _, p := range b.Packages {
		// Convert secret as a map
		secrets := map[string]interface{}{}
		for _, s := range p.Secrets.Data {
			var out interface{}
			if err := secret.Unpack(s.Value, &out); err != nil {
				return fmt.Errorf("unable to load secret value, corrupted bundle: %w", err)
			}

			// Assign to secret map
			secrets[s.Key] = out
		}

		// Export secrets as JSON
		content, err := json.Marshal(secrets)
		if err != nil {
			return fmt.Errorf("unable to extract secret map as json")
		}

		// Apply transformer
		out, err := transformer.To(ctx, content)
		if err != nil {
			return fmt.Errorf("unable to apply secret transformer: %w", err)
		}

		// Cleanup
		memguard.WipeBytes(content)
		p.Secrets.Data = nil

		// Assign locked secret
		p.Secrets.Locked = &wrappers.BytesValue{
			Value: out,
		}
	}

	// No error
	return nil
}

// UnLock apply transformer function to all secret values and set as unlocked.
func UnLock(ctx context.Context, b *bundlev1.Bundle, transformers []value.Transformer, skipNotDecryptable bool) error {
	// Check bundle
	if b == nil {
		return fmt.Errorf("unable to process nil bundle")
	}
	if len(transformers) == 0 {
		return fmt.Errorf("unable to process empty transformer list")
	}

	// For each packages
	for _, p := range b.Packages {
		// Skip not locked package
		if p.Secrets.Locked == nil {
			continue
		}
		if len(p.Secrets.Locked.Value) == 0 {
			continue
		}

		// Try all transformers
		var (
			out          []byte
			errTransform error
		)
	LOOP:
		for _, t := range transformers {
			// Apply transformation
			out, errTransform = t.From(ctx, p.Secrets.Locked.Value)
			switch {
			case errTransform != nil:
				// Try next transformer
				continue
			default:
				break LOOP
			}
		}
		if errTransform != nil {
			if skipNotDecryptable {
				// Skip not decrypted secrets.
				continue
			}
			return fmt.Errorf("unable to transform '%s': %w", p.Name, errTransform)
		}

		// Unpack secrets
		raw := map[string]interface{}{}
		if err := json.Unmarshal(out, &raw); err != nil {
			return fmt.Errorf("unable to unpack locked secret: %w", err)
		}

		// Prepare secrets collection
		secrets := []*bundlev1.KV{}
		for key, value := range raw {
			// Pack secret value
			s, err := secret.Pack(value)
			if err != nil {
				return fmt.Errorf("unable to pack as secret bundle: %w", err)
			}

			// Add to secret collection
			secrets = append(secrets, &bundlev1.KV{
				Key:   key,
				Type:  fmt.Sprintf("%T", value),
				Value: s,
			})
		}

		// Cleanup
		memguard.WipeBytes(p.Secrets.Locked.Value)
		p.Secrets.Locked = nil

		// Assign unlocked secrets
		p.Secrets.Data = secrets
	}

	// No error
	return nil
}
