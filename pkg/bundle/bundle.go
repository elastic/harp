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
	"fmt"
	"reflect"
	"sort"

	"gitlab.com/NebulousLabs/merkletree"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	csov1 "github.com/elastic/harp/pkg/cso/v1"

	"golang.org/x/crypto/blake2b"
)

// Annotate a bundle object.
func Annotate(obj AnnotationOwner, key, value string) {
	updateAnnotations(obj, obj.GetAnnotations(), key, value)
}

// Labelize a bundle object.
func Labelize(obj LabelOwner, key, value string) {
	updateLabels(obj, obj.GetLabels(), key, value)
}

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

// Paths returns ordered bundle secret paths.
func Paths(b *bundlev1.Bundle) ([]string, error) {
	// Check input
	if b == nil {
		return nil, fmt.Errorf("unable to process nil bundle")
	}

	// Get all paths
	res := []string{}
	for _, p := range b.Packages {
		res = append(res, p.Name)
	}

	// Sort result
	sort.SliceStable(res, func(i, j int) bool {
		return res[i] < res[j]
	})

	// No error
	return res, nil
}

// SecretReader is used by template engine to resolve secret from secret container.
func SecretReader(b *bundlev1.Bundle) func(path string) (map[string]interface{}, error) {
	return func(secretPath string) (map[string]interface{}, error) {
		return Read(b, secretPath)
	}
}

// -----------------------------------------------------------------------------
func updateAnnotations(obj interface{}, m map[string]string, key, value string) {
	// Check allocation
	if m == nil {
		m = map[string]string{}
	}

	// Check if map key is already assigned
	if _, ok := m[key]; ok {
		return
	}

	// Assign value
	m[key] = value

	// Reaffect map to owner
	// Really not fan of this ... but protobuf doesn't generate setters for go
	mv := reflect.ValueOf(m)
	f := reflect.ValueOf(obj).Elem().FieldByName("Annotations")
	if f.IsValid() && f.CanSet() && f.Kind() == mv.Kind() {
		f.Set(mv)
	}
}

func updateLabels(obj interface{}, m map[string]string, key, value string) {
	// Check allocation
	if m == nil {
		m = map[string]string{}
	}

	// Check if map key is already assigned
	if _, ok := m[key]; ok {
		return
	}

	// Assign value
	m[key] = value

	// Reaffect map to owner
	// Really not fan of this ... but protobuf doesn't generate setters for go
	mv := reflect.ValueOf(m)
	f := reflect.ValueOf(obj).Elem().FieldByName("Labels")
	if f.IsValid() && f.CanSet() && f.Kind() == mv.Kind() {
		f.Set(mv)
	}
}
