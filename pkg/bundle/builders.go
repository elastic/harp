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
	"strings"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle/compare"
	"github.com/elastic/harp/pkg/bundle/secret"
	"github.com/elastic/harp/pkg/sdk/types"
	"google.golang.org/protobuf/encoding/protojson"
)

// FromDump creates a bundle from a JSON Dump.
func FromDump(r io.Reader) (*bundlev1.Bundle, error) {
	// Check parameters
	if types.IsNil(r) {
		return nil, fmt.Errorf("unable to process nil reader")
	}

	// Drain input content
	content, err := io.ReadAll(r)
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
		switch op.Type {
		case "package":
			// Ignore package operation
			continue
		case "secret":
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
		default:
			return nil, fmt.Errorf("unknown oplog type '%s'", op.Type)
		}
	}

	// Assign packages
	for _, p := range packageMap {
		b.Packages = append(b.Packages, p)
	}

	// No error
	return b, nil
}

// FromMap builds a secret container from map K/V.
func FromMap(input map[string]KV) (*bundlev1.Bundle, error) {
	// Check input
	if input == nil {
		return nil, fmt.Errorf("unable to process nil map")
	}

	res := &bundlev1.Bundle{
		Packages: []*bundlev1.Package{},
	}
	for packageName, secretKv := range input {
		// Prepare a package
		p := &bundlev1.Package{
			Name:    packageName,
			Secrets: &bundlev1.SecretChain{},
		}

		// Prepare secret data
		for k, v := range secretKv {
			// Pack secret value
			packed, err := secret.Pack(v)
			if err != nil {
				return nil, fmt.Errorf("unable to pack secret value for `%s`: %w", fmt.Sprintf("%s.%s", packageName, k), err)
			}

			// Add to secret package
			p.Secrets.Data = append(p.Secrets.Data, &bundlev1.KV{
				Key:   k,
				Type:  fmt.Sprintf("%T", v),
				Value: packed,
			})
		}

		// Add package to result
		res.Packages = append(res.Packages, p)
	}

	// No error
	return res, nil
}
