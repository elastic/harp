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

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle/secret"
	"github.com/gobwas/glob"
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
