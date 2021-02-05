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

package kv

import (
	"context"
	"fmt"

	"github.com/elastic/harp/pkg/vault/logical"
	vpath "github.com/elastic/harp/pkg/vault/path"
)

type kvv1Backend struct {
	logical   logical.Logical
	mountPath string
}

// V1 returns a K/V v1 backend service instance.
func V1(l logical.Logical, mountPath string) Service {
	return &kvv1Backend{
		logical:   l,
		mountPath: mountPath,
	}
}

// -----------------------------------------------------------------------------
func (s *kvv1Backend) List(ctx context.Context, path string) ([]string, error) {
	// Clean path first
	secretPath := vpath.SanitizePath(path)
	if secretPath == "" {
		return nil, fmt.Errorf("unable to query with empty path")
	}

	// Create logical client
	secret, err := s.logical.List(secretPath)
	if err != nil {
		return nil, fmt.Errorf("unable to list secret keys: %w", err)
	}
	if secret == nil {
		// Path is a leaf
		return nil, nil
	}
	if secret.Data == nil {
		return nil, fmt.Errorf("invalid secret response")
	}

	// Check required property
	k, ok := secret.Data["keys"]
	if !ok || k == nil {
		return nil, fmt.Errorf("invalid response missing 'keys' property")
	}

	// Check value type
	r, ok := k.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response 'keys' is not a list (%T)", k)
	}

	// Convert list of interface to list of string
	out := make([]string, len(r))
	for i := range r {
		out[i] = r[i].(string)
	}

	// No error
	return out, nil
}

func (s *kvv1Backend) Read(ctx context.Context, path string) (SecretData, SecretMetadata, error) {
	// Clean path first
	secretPath := vpath.SanitizePath(path)
	if secretPath == "" {
		return nil, nil, fmt.Errorf("unable to query with empty path")
	}

	// Create a logical client
	secret, err := s.logical.Read(secretPath)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to retrieve secret for path '%s': %w", path, err)
	}
	if secret == nil {
		return nil, nil, fmt.Errorf("unable to retrieve secret for path '%s': %w", path, ErrPathNotFound)
	}
	if secret.Data == nil {
		return nil, nil, fmt.Errorf("unable to retrieve secret for path '%s': %w", path, ErrNoData)
	}

	// Return secret value and no error
	return secret.Data, nil, err
}

func (s *kvv1Backend) ReadVersion(ctx context.Context, path string, version uint32) (SecretData, SecretMetadata, error) {
	return s.Read(ctx, path)
}

func (s *kvv1Backend) Write(ctx context.Context, path string, data SecretData) error {
	// Clean path first
	secretPath := vpath.SanitizePath(path)
	if secretPath == "" {
		return fmt.Errorf("unable to query with empty path")
	}

	// Create a logical client
	_, err := s.logical.Write(secretPath, data)
	if err != nil {
		return fmt.Errorf("unable to write secret data for path '%s': %w", path, err)
	}

	return nil
}
