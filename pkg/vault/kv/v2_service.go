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
	"errors"
	"fmt"

	"github.com/hashicorp/vault/api"

	"github.com/elastic/harp/pkg/vault/logical"
	vpath "github.com/elastic/harp/pkg/vault/path"
)

type kvv2Backend struct {
	logical               logical.Logical
	mountPath             string
	customMetadataEnabled bool
}

// V2 returns a K/V v2 backend service instance.
func V2(l logical.Logical, mountPath string, customMetadataEnabled bool) Service {
	return &kvv2Backend{
		logical:               l,
		mountPath:             mountPath,
		customMetadataEnabled: customMetadataEnabled,
	}
}

// -----------------------------------------------------------------------------
func (s *kvv2Backend) List(ctx context.Context, path string) ([]string, error) {
	// Check arguments
	secretPath := vpath.SanitizePath(path)
	if secretPath == "" {
		return nil, fmt.Errorf("unable to query with empty path")
	}

	// Create logical client
	secret, err := s.logical.List(vpath.AddPrefixToVKVPath(secretPath, s.mountPath, "metadata"))
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
		out[i] = fmt.Sprintf("%v", r[i])
	}

	// No error
	return out, nil
}

func (s *kvv2Backend) Read(ctx context.Context, path string) (SecretData, SecretMetadata, error) {
	return s.ReadVersion(ctx, path, 0)
}

func (s *kvv2Backend) ReadVersion(ctx context.Context, path string, version uint32) (SecretData, SecretMetadata, error) {
	// Clean path first
	secretPath := vpath.SanitizePath(path)
	if secretPath == "" {
		return nil, nil, fmt.Errorf("unable to query with empty path")
	}

	var (
		secret *api.Secret
		err    error
	)

	// Create a logical client
	if version > 0 {
		// Prepare params
		versionParam := map[string][]string{
			"version": {fmt.Sprintf("%d", version)},
		}

		secret, err = s.logical.ReadWithData(vpath.AddPrefixToVKVPath(secretPath, s.mountPath, "data"), versionParam)
	} else {
		secret, err = s.logical.Read(vpath.AddPrefixToVKVPath(secretPath, s.mountPath, "data"))
	}
	if err != nil {
		return nil, nil, fmt.Errorf("unable to retrieve secret for path '%s': %w", path, err)
	}
	if secret == nil {
		return nil, nil, fmt.Errorf("unable to retrieve secret for path '%s': %w", path, ErrPathNotFound)
	}
	if secret.Data == nil {
		return nil, nil, fmt.Errorf("unable to retrieve secret for path '%s': %w", path, ErrNoData)
	}

	// Check v2 backend
	data, ok := secret.Data["data"]
	if !ok {
		return nil, nil, fmt.Errorf("unable to extract values for path '%s', secret backend supposed to be a v2 but it's not", path)
	}
	metadata, ok := secret.Data["metadata"].(map[string]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("unable to extract metadata for path '%s', secret backend supposed to be a v2 but it's not", path)
	}

	// Check data
	if data == nil {
		return nil, nil, ErrNoData
	}

	// Custom metadata enabled => retrieve secret meatadata.
	if s.customMetadataEnabled {
		rawMeta, err := s.logical.Read(vpath.AddPrefixToVKVPath(secretPath, s.mountPath, "metadata"))
		if err != nil {
			return nil, nil, fmt.Errorf("unable to extract secret metadata for path '%s': %w", path, err)
		}
		if rawMeta == nil {
			return nil, nil, fmt.Errorf("unable to retrieve secret metadata for path '%s': %w", path, ErrPathNotFound)
		}
		if rawMeta.Data == nil {
			return nil, nil, fmt.Errorf("unable to retrieve secret metadata for path '%s': %w", path, ErrNoData)
		}

		// Check if response contains custom_metadata
		if rawCustomMeta, ok := rawMeta.Data["custom_metadata"]; ok {
			if customMeta, ok := rawCustomMeta.(map[string]interface{}); ok {
				metadata["custom_metadata"] = customMeta
			}
		}
	}

	// Return secret value and no error
	return data.(map[string]interface{}), metadata, err
}

func (s *kvv2Backend) Write(ctx context.Context, path string, data SecretData) error {
	return s.WriteWithMeta(ctx, path, data, nil)
}

func (s *kvv2Backend) WriteWithMeta(ctx context.Context, path string, data SecretData, meta SecretMetadata) error {
	// Clean path first
	secretPath := vpath.SanitizePath(path)
	if secretPath == "" {
		return fmt.Errorf("unable to query with empty path")
	}

	// Custom metadata not enabled => store meatadata as secret data.
	if s.customMetadataEnabled {
		// Validate metadata
		if len(meta) > CustomMetadataKeyLimit {
			return errors.New("unable to store more than 64 custom metadata keys")
		}

		// Check key and value constraints
		for k, v := range meta {
			if len(k) > CustomMetadataKeySizeLimit {
				return fmt.Errorf("custom meta '%s' could not be stored, it must be less than 128 bytes", k)
			}
			raw, ok := v.(string)
			if !ok {
				return fmt.Errorf("custom meta '%s' must be a string", k)
			}
			if len(raw) > CustomMetadataValueSizeLimit {
				return fmt.Errorf("custom meta '%s' value is too large (%d), it must be less than 512 bytes", k, len(raw))
			}
		}
	} else if len(meta) > 0 {
		// Add metadata to data
		data[VaultMetadataDataKey] = meta
	}

	// Write data
	_, err := s.logical.Write(vpath.AddPrefixToVKVPath(secretPath, s.mountPath, "data"), map[string]interface{}{
		"data": data,
	})
	if err != nil {
		return fmt.Errorf("unable to write secret data for path '%s': %w", path, err)
	}

	// Write metadata
	if s.customMetadataEnabled && len(meta) > 0 {
		_, err := s.logical.Write(vpath.AddPrefixToVKVPath(secretPath, s.mountPath, "metadata"), map[string]interface{}{
			"custom_metadata": meta,
		})
		if err != nil {
			return fmt.Errorf("unable to write secret metadata for path '%s': %w", path, err)
		}
	}

	// No error
	return nil
}
