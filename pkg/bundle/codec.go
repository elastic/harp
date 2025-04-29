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
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle/secret"
	"github.com/elastic/harp/pkg/sdk/security"
	"github.com/elastic/harp/pkg/sdk/types"
)

// Load a file bundle from the buffer.
func Load(r io.Reader) (*bundlev1.Bundle, error) {
	// Check parameters
	if types.IsNil(r) {
		return nil, fmt.Errorf("unable to process nil reader")
	}

	decoded, err := io.ReadAll(r)
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

	// Sort packages
	sort.SliceStable(b.Packages, func(i, j int) bool {
		return b.Packages[i].Name < b.Packages[j].Name
	})

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

// AsProtoJSON export given bundle as a JSON representation.
//
//nolint:interfacer // Tighly coupled with type
func AsProtoJSON(w io.Writer, b *bundlev1.Bundle) error {
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

// AsMap returns a bundle as map
func AsMap(b *bundlev1.Bundle) (KV, error) {
	// Check input
	if b == nil {
		return nil, fmt.Errorf("unable to process nil bundle")
	}

	res := KV{}
	for _, p := range b.Packages {
		// Check if secret is locked
		if p.Secrets.Locked != nil {
			// Encode value
			res[p.Name] = KV{
				"@type": packageEncryptedValueType,
				"value": p.Secrets.Locked.Value,
			}
			continue
		}

		// Map package secrets
		secrets, err := AsSecretMap(p)
		if err != nil {
			return nil, fmt.Errorf("unable to pack secrets as a map: %w", err)
		}

		// Assign result
		res[p.Name] = secrets
	}

	// No error
	return res, nil
}

// AsMetadataMap exports given bundle metadata as a map.
func AsMetadataMap(b *bundlev1.Bundle) (KV, error) {
	// Check input
	if b == nil {
		return nil, fmt.Errorf("unable to process nil bundle")
	}

	metaMap := KV{}

	// Check if bundle as metadata
	if len(b.Annotations) > 0 {
		metaMap[bundleAnnotationsKey] = b.Annotations
	}
	if len(b.Labels) > 0 {
		metaMap[bundleLabelsKey] = b.Labels
	}

	// Export bundle metadata
	for _, p := range b.Packages {
		metadata := KV{}
		// Has annotations
		if len(p.Annotations) > 0 {
			// Assign json
			metadata[packageAnnotations] = p.Annotations
		}
		// Has labels
		if len(p.Labels) > 0 {
			// Assign json
			metadata[packageLabels] = p.Labels
		}

		// Assign to package
		metaMap[p.Name] = metadata
	}

	// No error
	return metaMap, nil
}
