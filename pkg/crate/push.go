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

package crate

import (
	"context"
	"errors"
	"fmt"

	"github.com/elastic/harp/pkg/sdk/types"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/oras"
	"oras.land/oras-go/pkg/target"
)

// Push the given image descriptor in the given repository.
func Push(ctx context.Context, registry target.Target, imageRef string, i *Image) (*ocispec.Descriptor, error) {
	// Check arguments
	if types.IsNil(registry) {
		return nil, errors.New("registry must not be nil")
	}
	if imageRef == "" {
		return nil, errors.New("image reference must not be blank")
	}
	if i == nil {
		return nil, errors.New("image must not be nil")
	}

	// Create in-memory image
	memoryStore := content.NewMemory()

	// Generate image
	manifest, manifestDesc, err := PrepareImage(memoryStore, i)
	if err != nil {
		return nil, fmt.Errorf("unable to generate manifest: %w", err)
	}

	// Add the manifest to store.
	if errManifest := memoryStore.StoreManifest(imageRef, *manifestDesc, manifest); errManifest != nil {
		return nil, fmt.Errorf("unable to register manifest: %w", errManifest)
	}

	// Pushing the image
	containerManifest, err := oras.Copy(ctx, memoryStore, imageRef, registry, "")
	if err != nil {
		return nil, fmt.Errorf("pushing manifest: %w", err)
	}

	// No error
	return &containerManifest, nil
}
