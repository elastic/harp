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
	"io"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/awnumar/memguard"
	"github.com/golang/protobuf/ptypes/wrappers"
	"gitlab.com/NebulousLabs/merkletree"
	"golang.org/x/crypto/blake2b"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle/compare"
	"github.com/elastic/harp/pkg/bundle/secret"
	csov1 "github.com/elastic/harp/pkg/cso/v1"
	"github.com/elastic/harp/pkg/sdk/security"
	"github.com/elastic/harp/pkg/sdk/types"
	"github.com/elastic/harp/pkg/sdk/value"
)

// Tree returns a merkle tree based on secrets hierarchy
func Tree(b *bundlev1.Bundle) (*merkletree.Tree, *Statistic, error) {
	// Check bundle
	if b == nil {
		return nil, nil, fmt.Errorf("unable to process nil bundle")
	}

	// Calculate merkle tree root
	h, err := blake2b.New512(nil)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to initialize hash function for merkle tree")
	}

	// Initialize merkle tree
	tree := merkletree.New(h)
	if err = tree.SetIndex(1); err != nil {
		return nil, nil, fmt.Errorf("unable to initialize merkle tree")
	}

	// Prepare statistics
	stats := &Statistic{
		SecretCount:                  0,
		PackageCount:                 0,
		CSOCompliantPackageNameCount: 0,
	}

	// Ensure packages order
	sort.SliceStable(b.Packages, func(i, j int) bool {
		return b.Packages[i].Name < b.Packages[j].Name
	})

	// All packages
	for _, p := range b.Packages {
		// Increment package count
		stats.PackageCount++

		// Check compliance with CSO
		if errValidate := csov1.Validate(p.Name); errValidate == nil {
			stats.CSOCompliantPackageNameCount++
		}

		// Prepare secret uri list
		uris := []string{}

		// Follow secret chain
		if p.Secrets != nil {
			for _, s := range p.Secrets.Data {
				// Increment secret count
				stats.SecretCount++

				// Build merkle tree leaf
				uris = append(uris, fmt.Sprintf("%s:%d:%s:%x", p.Name, p.Secrets.Version, s.Key, blake2b.Sum512(s.Value)))
			}

			// Sort them
			sort.Strings(uris)

			// Push sorted secret uri as proof
			for _, u := range uris {
				tree.Push([]byte(u))
			}
		}
	}

	// Return the tree
	return tree, stats, nil
}

// Load a file bundle from the buffer.
func Load(r io.Reader) (*bundlev1.Bundle, error) {
	// Check parameters
	if types.IsNil(r) {
		return nil, fmt.Errorf("unable to process nil reader")
	}

	decoded, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("unable to decompress bundle content")
	}

	// Deserialize protobuf payload
	bundle := &bundlev1.Bundle{}
	if err = proto.Unmarshal(decoded, bundle); err != nil {
		return nil, fmt.Errorf("unable to decode bundle content")
	}

	// Compute merkle tree root
	tree, _, err := Tree(bundle)
	if err != nil {
		return nil, fmt.Errorf("unable to compute merkle tree of bundle content: %w", err)
	}

	// Check if root match
	if !security.SecureCompare(bundle.MerkleTreeRoot, tree.Root()) {
		return nil, fmt.Errorf("invalid merkle tree root, bundle is corrupted")
	}

	// No error
	return bundle, nil
}

// Dump a file bundle to the writer.
func Dump(w io.Writer, b *bundlev1.Bundle) error {
	// Check parameters
	if types.IsNil(w) {
		return fmt.Errorf("unable to process nil writer")
	}
	if b == nil {
		return fmt.Errorf("unable to process nil bundle")
	}

	// Compute merkle tree
	tree, _, err := Tree(b)
	if err != nil {
		return fmt.Errorf("unable to compute merkle tree of bundle content: %w", err)
	}

	// Assign to bundle
	b.MerkleTreeRoot = tree.Root()

	// Serialize protobuf payload
	payload, err := proto.Marshal(b)
	if err != nil {
		return fmt.Errorf("unable to encode bundle content: %w", err)
	}

	// WWrite to writer
	if _, err = w.Write(payload); err != nil {
		return fmt.Errorf("unable to write serialized Bundle: %w", err)
	}

	// No error
	return nil
}

// Read a secret located at secretPath from the given bundle.
func Read(b *bundlev1.Bundle, secretPath string) (map[string]interface{}, error) {
	// Check bundle
	if b == nil {
		return nil, fmt.Errorf("unable to process nil bundle")
	}
	if secretPath == "" {
		return nil, fmt.Errorf("unable to process with empty path")
	}

	// Lookup secret package
	var found *bundlev1.Package
	for _, item := range b.Packages {
		if strings.EqualFold(item.Name, secretPath) {
			found = item
			break
		}
	}
	if found == nil {
		return nil, fmt.Errorf("unable to lookup secret with path '%s'", secretPath)
	}

	// Transform secret value
	result := map[string]interface{}{}
	for _, s := range found.Secrets.Data {
		// Unpack secret value
		var obj interface{}
		if err := secret.Unpack(s.Value, &obj); err != nil {
			return nil, fmt.Errorf("unable to unpack secret value for path '%s': %w", secretPath, err)
		}

		// Add to result
		result[s.Key] = obj
	}

	// No error
	return result, nil
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
func UnLock(ctx context.Context, b *bundlev1.Bundle, transformer value.Transformer) error {
	// Check bundle
	if b == nil {
		return fmt.Errorf("unable to process nil bundle")
	}
	if types.IsNil(transformer) {
		return fmt.Errorf("unable to process nil transformer")
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

		// Apply transformation
		out, err := transformer.From(ctx, p.Secrets.Locked.Value)
		if err != nil {
			return fmt.Errorf("unable to apply secret transformer: %w", err)
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

// JSON export given bundle as a JSON representation.
//nolint:interfacer // Tighly coupled with type
func JSON(w io.Writer, b *bundlev1.Bundle) error {
	// Check parameters
	if types.IsNil(w) {
		return fmt.Errorf("unable to process nil writer")
	}
	if b == nil {
		return fmt.Errorf("unable to process nil bundle")
	}

	// Clone bundle (we don't want to modify input bundle)
	cloned, ok := proto.Clone(b).(*bundlev1.Bundle)
	if !ok {
		return fmt.Errorf("the cloned bundle does not have a correct type: %T", cloned)
	}

	// Initialize marshaller
	m := &protojson.MarshalOptions{}

	// Decode packed values
	for _, p := range cloned.Packages {
		for _, s := range p.Secrets.Data {
			// Unpack secret value
			var data interface{}
			if err := secret.Unpack(s.Value, &data); err != nil {
				return fmt.Errorf("unable to unpack '%s' - '%s' secret value: %w", p.Name, s.Key, err)
			}

			// Re-encode as json
			payload, err := json.Marshal(data)
			if err != nil {
				return fmt.Errorf("unable to encode '%s' - '%s' secret value as json: %w", p.Name, s.Key, err)
			}

			// Replace current packed secret value by json encoded one.
			s.Value = payload
		}
	}

	// Marshal bundle
	out, err := m.Marshal(cloned)
	if err != nil {
		return fmt.Errorf("unable to produce JSON from bundle object: %w", err)
	}

	// Write to writer
	if _, err := fmt.Fprintf(w, "%s", string(out)); err != nil {
		return fmt.Errorf("unable to write JSON bundle: %w", err)
	}

	// No error
	return nil
}

// FromDump creates a bundle from a JSON Dump.
func FromDump(r io.Reader) (*bundlev1.Bundle, error) {
	// Check parameters
	if types.IsNil(r) {
		return nil, fmt.Errorf("unable to process nil reader")
	}

	// Drain input content
	content, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("unable to read input content: %w", err)
	}

	// Build the container from json
	var b bundlev1.Bundle
	if err = protojson.Unmarshal(content, &b); err != nil {
		return nil, fmt.Errorf("unable to decode JSON bundle: %w", err)
	}

	// Convert secret values to current value packing method.
	for _, p := range b.Packages {
		for _, s := range p.Secrets.Data {
			// Decode json encoded value
			var data interface{}
			if errJSON := json.Unmarshal(s.Value, &data); errJSON != nil {
				return nil, fmt.Errorf("unable to decode '%s' - '%s' secret value as json: %w", p.Name, s.Key, errJSON)
			}

			// Pack secret value
			payload, err := secret.Pack(data)
			if err != nil {
				return nil, fmt.Errorf("unable to pack '%s' - '%s' secret value: %w", p.Name, s.Key, err)
			}

			// Replace current json encoded secret value by packed one.
			s.Value = payload
		}
	}

	// No error
	return &b, nil
}

// FromOpLog convert oplog to a bundle.
func FromOpLog(oplog compare.OpLog) (*bundlev1.Bundle, error) {
	// Create an empty bundle.
	b := &bundlev1.Bundle{}

	packageMap := map[string]*bundlev1.Package{}

	// Generate patch rules
	for _, op := range oplog {
		if op.Type == "package" {
			// Ignore package operation
			continue
		}
		if op.Type == "secret" {
			pathParts := strings.SplitN(op.Path, "#", 2)
			pkg, ok := packageMap[pathParts[0]]
			if !ok {
				packageMap[pathParts[0]] = &bundlev1.Package{
					Name: pathParts[0],
					Secrets: &bundlev1.SecretChain{
						Data: []*bundlev1.KV{},
					},
				}
				pkg = packageMap[pathParts[0]]
			}

			// Process oplog event
			switch op.Operation {
			case compare.Add, compare.Replace:
				// Pack secret value
				payload, err := secret.Pack(op.Value)
				if err != nil {
					return nil, fmt.Errorf("unable to pack secret value for '%s' / '%s': %w", pathParts[0], pathParts[1], err)
				}

				// Assign secret data
				pkg.Secrets.Data = append(pkg.Secrets.Data, &bundlev1.KV{
					Key:   pathParts[1],
					Type:  "string",
					Value: payload,
				})
			case compare.Remove:
				// Ignore secret removal
			}
		}
	}

	// Assign packages
	for _, p := range packageMap {
		b.Packages = append(b.Packages, p)
	}

	// No error
	return b, nil
}
