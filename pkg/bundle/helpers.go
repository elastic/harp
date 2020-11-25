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
	"errors"
	"fmt"
	"reflect"

	"github.com/gobwas/glob"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle/secret"
)

// KV describes map[string]interface{} alias
type KV map[string]interface{}

// Glob returns package objects that have name matching the given pattern.
func (kv KV) Glob(pattern string) KV {
	// Prepare Glob filter.
	g, err := glob.Compile(pattern, '/')
	if err != nil {
		g, _ = glob.Compile("**")
	}

	// Apply to collection
	nkv := KV{}
	for name, contents := range kv {
		if g.Match(name) {
			nkv[name] = contents
		}
	}

	return nkv
}

// Get returns a KV of the given package.
func (kv KV) Get(name string) interface{} {
	if v, ok := kv[name]; ok {
		return v
	}
	return KV{}
}

// AsMap returns a bundle as map
func AsMap(b *bundlev1.Bundle) (KV, error) {
	// Check input
	if b == nil {
		return nil, fmt.Errorf("unable to process nil bundle")
	}

	res := KV{}
	for _, p := range b.Packages {
		// Map package secrets
		secrets, err := AsSecretMap(p)
		if err != nil {
			return nil, err
		}

		// Assign result
		res[p.Name] = secrets
	}

	// No error
	return res, nil
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

// Paths returns bundle secret paths.
func Paths(b *bundlev1.Bundle) ([]string, error) {
	// Check input
	if b == nil {
		return nil, fmt.Errorf("unable to process nil bundle")
	}

	res := []string{}
	for _, p := range b.Packages {
		res = append(res, p.Name)
	}

	// No error
	return res, nil
}

// AnnotationOwner defines annotations owner contract
type AnnotationOwner interface {
	GetAnnotations() map[string]string
}

// LabelOwner defines label owner contract
type LabelOwner interface {
	GetLabels() map[string]string
}

func updateStringMap(obj interface{}, m map[string]string, fieldName, key, value string) {
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
	reflect.ValueOf(obj).Elem().FieldByName(fieldName).Set(reflect.ValueOf(m))
}

// Annotate a bundle object.
func Annotate(obj AnnotationOwner, key, value string) {
	updateStringMap(obj, obj.GetAnnotations(), "Annotations", key, value)
}

// Labelize a bundle object.
func Labelize(obj LabelOwner, key, value string) {
	updateStringMap(obj, obj.GetLabels(), "Labels", key, value)
}

// -----------------------------------------------------------------------------

// SecretReader is used by template engine to resolve secret from secret container.
func SecretReader(b *bundlev1.Bundle) func(path string) (map[string]interface{}, error) {
	return func(secretPath string) (map[string]interface{}, error) {
		return Read(b, secretPath)
	}
}

// -----------------------------------------------------------------------------

// AsSecretMap returns a KV map from given package.
func AsSecretMap(p *bundlev1.Package) (KV, error) {
	// Check arguments
	if p == nil {
		return nil, errors.New("unable to transform nil package")
	}

	secrets := KV{}
	for _, s := range p.Secrets.Data {
		// Unpack secret value
		var data interface{}
		if err := secret.Unpack(s.Value, &data); err != nil {
			return nil, fmt.Errorf("unable to unpack '%s' secret value: %w", p.Name, err)
		}

		// Assign result
		secrets[s.Key] = data
	}

	// No error
	return secrets, nil
}

// FromSecretMap returns the protobuf representation of secretMap.
func FromSecretMap(secretKv KV) ([]*bundlev1.KV, error) {
	secrets := []*bundlev1.KV{}

	// Prepare secret data
	for k, v := range secretKv {
		// Pack secret value
		packed, err := secret.Pack(v)
		if err != nil {
			return nil, fmt.Errorf("unable to pack secret value for `%s`: %w", k, err)
		}

		// Add to secret package
		secrets = append(secrets, &bundlev1.KV{
			Key:   k,
			Type:  fmt.Sprintf("%T", v),
			Value: packed,
		})
	}

	// No error
	return secrets, nil
}
