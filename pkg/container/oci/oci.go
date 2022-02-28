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
	"errors"
	"fmt"
	"path"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/elastic/harp/pkg/container"
	"github.com/elastic/harp/pkg/sdk/types"
)

// StoreSetter is the interface used to mock the image store.
type StoreSetter interface {
	Set(ocispec.Descriptor, []byte)
}

// AddSealedContainer registers a new layer to the current store for the given selaed container.
func AddSealedContainer(store StoreSetter, c *SealedContainer) (*ocispec.Descriptor, error) {
	// Check arguments
	if types.IsNil(store) {
		return nil, errors.New("unable to register sealed container with nil storage")
	}
	if c == nil {
		return nil, errors.New("given container is nil")
	}
	if !container.IsSealed(c.Container) {
		return nil, errors.New("the given container must be sealed")
	}

	// Dump the container
	var payload bytes.Buffer
	if err := container.Dump(&payload, c.Container); err != nil {
		return nil, fmt.Errorf("unable to dump container: %w", err)
	}

	// Get layer content
	content := payload.Bytes()

	// Prepare a layer
	containerDesc := ocispec.Descriptor{
		MediaType: harpSealedContainerLayerMediaType,
		Digest:    digest.FromBytes(content),
		Size:      int64(len(content)),
		Annotations: map[string]string{
			ocispec.AnnotationTitle: path.Join("containers", path.Clean(c.Name)),
		},
	}

	// Assign the store
	store.Set(containerDesc, content)

	// No error
	return &containerDesc, nil
}

// AddTemplateArchive registers a new layer to the current store for the given archive.
func AddTemplateArchive(store StoreSetter, ta *TemplateArchive) (*ocispec.Descriptor, error) {
	// Check arguments
	if types.IsNil(store) {
		return nil, errors.New("unable to register sealed container with nil storage")
	}
	if ta == nil {
		return nil, errors.New("given templte archive is nil")
	}

	// Prepare a layer
	containerDesc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageLayerGzip,
		Digest:    digest.FromBytes(ta.Archive),
		Size:      int64(len(ta.Archive)),
		Annotations: map[string]string{
			ocispec.AnnotationTitle: path.Join("templates", path.Clean(ta.Name)),
		},
	}

	// Assign the store
	store.Set(containerDesc, ta.Archive)

	// No error
	return &containerDesc, nil
}
