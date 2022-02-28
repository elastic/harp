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
	"bytes"
	"context"
	"fmt"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/pkg/auth"
	"oras.land/oras-go/pkg/auth/docker"
	"oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/oras"

	containerv1 "github.com/elastic/harp/api/gen/go/harp/container/v1"
	"github.com/elastic/harp/pkg/container"
)

func Push(ctx context.Context, c *containerv1.Container, repository, path string) (*ocispec.Descriptor, error) {
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

	// Dump container
	var buf bytes.Buffer
	if errDump := container.Dump(&buf, c); errDump != nil {
		return nil, fmt.Errorf("unable to serialize container for OCI layer: %w", errDump)
	}

	// Create OCI layer
	memoryStore := content.NewMemory()
	sealedContainerLayer, err := memoryStore.Add(path, harpSealedContainerLayerMediaType, buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("building layers: %w", err)
	}

	// Generate manifest.
	manifest, manifestDesc, config, configDesc, err := content.GenerateManifestAndConfig(nil, nil, sealedContainerLayer)
	if err != nil {
		return nil, fmt.Errorf("unable to generate OCI manifest: %w", err)
	}

	// Add the manifest to memory store.
	memoryStore.Set(configDesc, config)
	if errManifest := memoryStore.StoreManifest(repository, manifestDesc, manifest); errManifest != nil {
		return nil, fmt.Errorf("unable to register OCI manifest: %w", errManifest)
	}

	// Pushing the image
	containerManifest, err := oras.Copy(ctx, memoryStore, repository, registry, "")
	if err != nil {
		return nil, fmt.Errorf("pushing manifest: %w", err)
	}

	// No error
	return &containerManifest, nil
}
