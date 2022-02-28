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

package oci

import (
	"context"
	"errors"
	"fmt"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/pkg/auth"
	"oras.land/oras-go/pkg/auth/docker"
	"oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/oras"
)

// Push the given image descriptor in the given repository.
func Push(ctx context.Context, repository string, i *Image) (*ocispec.Descriptor, error) {
	// Check arguments
	if repository == "" {
		return nil, errors.New("repository must not be blank")
	}
	if i == nil {
		return nil, errors.New("image must not be nil")
	}

	// Use docker client.
	cli, err := docker.NewClient()
	if err != nil {
		return nil, fmt.Errorf("docker client: %w", err)
	}

	// Prepare authenticated client.
	registry, err := cli.ResolverWithOpts(auth.WithResolverPlainHTTP())
	if err != nil {
		return nil, fmt.Errorf("docker resolver: %w", err)
	}

	// Create in-memory image
	memoryStore := content.NewMemory()
	descriptors := []ocispec.Descriptor{}

	// Add all containers
	for _, c := range i.Containers {
		// Create a layer for each sealed containers
		sealedContainerLayer, errLayer := AddSealedContainer(memoryStore, c)
		if errLayer != nil {
			return nil, fmt.Errorf("unable to add container layer: %w", errLayer)
		}

		// Add to manifest
		descriptors = append(descriptors, *sealedContainerLayer)
	}

	// Add all template archive
	for _, ta := range i.TemplateArchives {
		// Create a layer for each template archive
		templateLayer, errLayer := AddTemplateArchive(memoryStore, ta)
		if errLayer != nil {
			return nil, fmt.Errorf("unable to add template layer: %w", errLayer)
		}

		// Add to manifest
		descriptors = append(descriptors, *templateLayer)
	}

	// Generate manifest.
	manifest, manifestDesc, config, configDesc, err := content.GenerateManifestAndConfig(nil, nil, descriptors...)
	if err != nil {
		return nil, fmt.Errorf("unable to generate manifest: %w", err)
	}

	// Add the manifest to memory store.
	memoryStore.Set(configDesc, config)
	if errManifest := memoryStore.StoreManifest(repository, manifestDesc, manifest); errManifest != nil {
		return nil, fmt.Errorf("unable to register manifest: %w", errManifest)
	}

	// Pushing the image
	containerManifest, err := oras.Copy(ctx, memoryStore, repository, registry, "")
	if err != nil {
		return nil, fmt.Errorf("pushing manifest: %w", err)
	}

	// No error
	return &containerManifest, nil
}
