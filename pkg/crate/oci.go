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
	"bytes"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/pkg/content"

	containerv1 "github.com/elastic/harp/api/gen/go/harp/container/v1"
	"github.com/elastic/harp/build/version"
	"github.com/elastic/harp/pkg/container"
	"github.com/elastic/harp/pkg/crate/schema"
	schemav1 "github.com/elastic/harp/pkg/crate/schema/v1"
	"github.com/elastic/harp/pkg/sdk/types"
)

// StoreSetter is the interface used to mock the image store.
type StoreSetter interface {
	Set(ocispec.Descriptor, []byte)
}

// StoreGetter is the interface used to mock the image store.
type StoreGetter interface {
	GetByName(name string) (ocispec.Descriptor, []byte, bool)
}

// PrepareImage is used to assemble the OCI image according to given specification.
func PrepareImage(store StoreSetter, image *Image) ([]byte, *ocispec.Descriptor, error) {
	// Check arguments
	if types.IsNil(store) {
		return nil, nil, errors.New("unable to prepare an image with nil storage")
	}
	if image == nil {
		return nil, nil, errors.New("given image is nil")
	}

	// Add config
	config, err := addConfig(store, image)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to add config layer: %w", err)
	}

	layers := []ocispec.Descriptor{}

	// Add all containers
	for _, c := range image.Containers {
		// Create a layer for each sealed containers
		sealedContainerLayer, errLayer := addSealedContainer(store, c)
		if errLayer != nil {
			return nil, nil, fmt.Errorf("unable to add container layer: %w", errLayer)
		}

		// Add to manifest
		layers = append(layers, *sealedContainerLayer)
	}

	// Add all template archive
	for _, ta := range image.TemplateArchives {
		// Create a layer for each template archive
		templateLayer, errLayer := addTemplateArchive(store, ta)
		if errLayer != nil {
			return nil, nil, fmt.Errorf("unable to add template layer: %w", errLayer)
		}

		// Add to manifest
		layers = append(layers, *templateLayer)
	}

	// Generate manifest.
	manifestBytes, manifest, errManifest := content.GenerateManifest(config, map[string]string{
		ocispec.AnnotationCreated: time.Now().UTC().Format(time.RFC3339),
		ocispec.AnnotationVersion: version.Version,
	}, layers...)
	if errManifest != nil {
		return nil, nil, fmt.Errorf("unable to generate manifest: %w", errManifest)
	}

	// No error
	return manifestBytes, &manifest, nil
}

// ExtractImage extracts from the given OCI storage the crate image descriptor.
func ExtractImage(store StoreGetter) (*Image, error) {
	// Check arguments
	if store == nil {
		return nil, errors.New("can't extract crate from a nil store")
	}

	// Retrieve config
	cfg, err := getConfig(store)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve config from crate: %w", err)
	}

	var containers []*SealedContainer
	// Retrieve container
	for _, layerName := range cfg.Containers() {
		c, err := getSealedContainer(store, layerName)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve container from crate: %w", err)
		}

		// Add to containers
		containers = append(containers, &SealedContainer{
			Name:      strings.TrimPrefix(layerName, "containers/"),
			Container: c,
		})
	}

	// Retrieve archives
	var archives []*TemplateArchive
	for _, layerName := range cfg.Templates() {
		c, err := getTemplateArchive(store, layerName)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve archive from crate: %w", err)
		}

		// Add to containers
		archives = append(archives, &TemplateArchive{
			Name:    strings.TrimPrefix(layerName, "templates/"),
			Archive: c,
		})
	}

	// No error
	return &Image{
		Config:           cfg,
		Containers:       containers,
		TemplateArchives: archives,
	}, nil
}

// -----------------------------------------------------------------------------

func getConfig(store StoreGetter) (schema.Config, error) {
	// Check arguments
	if store == nil {
		return nil, errors.New("can't extract crate from a nil store")
	}

	// Extract config from image
	configDesc, configRaw, hasConfigLayer := store.GetByName("_config.json")
	if !hasConfigLayer {
		return nil, errors.New("unable to lookup config layer from crate")
	}
	if configDesc.MediaType != harpConfigMediaType {
		return nil, errors.New("invalid config layer media type value")
	}

	// Decode config
	cfg, err := schemav1.ParseConfig(configRaw)
	if err != nil {
		return nil, fmt.Errorf("unable to decode config object: %w", err)
	}

	// No error
	return cfg, nil
}

// AddConfig register the OCI configuration layer to retrieve information about
// the image.
func addConfig(store StoreSetter, image *Image) (*ocispec.Descriptor, error) {
	// Check arguments
	if types.IsNil(store) {
		return nil, errors.New("unable to register sealed container with nil storage")
	}
	if image == nil {
		return nil, errors.New("given image is nil")
	}

	// Render config as JSON.
	configBytes, err := schema.RenderConfig(image.Config)
	if err != nil {
		return nil, err
	}

	// Prepare layer
	configDesc := ocispec.Descriptor{
		MediaType: harpConfigMediaType,
		Digest:    digest.FromBytes(configBytes),
		Size:      int64(len(configBytes)),
		Annotations: map[string]string{
			ocispec.AnnotationTitle: "_config.json",
		},
	}

	// Assign to image store
	store.Set(configDesc, configBytes)

	// No error
	return &configDesc, nil
}

func getSealedContainer(store StoreGetter, layerName string) (*containerv1.Container, error) {
	// Check arguments
	if store == nil {
		return nil, errors.New("can't extract container from a nil store")
	}

	// Extract config from image
	containerDesc, containerRaw, hasContainerLayer := store.GetByName(layerName)
	if !hasContainerLayer {
		return nil, fmt.Errorf("unable to lookup conainer '%s' layer from crate", layerName)
	}
	if containerDesc.MediaType != harpContainerLayerMediaType {
		return nil, fmt.Errorf("invalid container layer media type value for '%s'", layerName)
	}

	// Validate as container
	c, err := container.Load(bytes.NewReader(containerRaw))
	if err != nil {
		return nil, fmt.Errorf("unable to decode csealed container from layer '%s': %w", layerName, err)
	}

	// Must be sealed.
	if !container.IsSealed(c) {
		return nil, fmt.Errorf("container layer '%s' has an unsealed container", layerName)
	}

	// No error
	return c, nil
}

// AddSealedContainer registers a new layer to the current store for the given selaed container.
func addSealedContainer(store StoreSetter, c *SealedContainer) (*ocispec.Descriptor, error) {
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
	body := payload.Bytes()

	// Prepare a layer
	containerDesc := ocispec.Descriptor{
		MediaType: harpContainerLayerMediaType,
		Digest:    digest.FromBytes(body),
		Size:      int64(len(body)),
		Annotations: map[string]string{
			ocispec.AnnotationTitle: path.Join("containers", path.Clean(c.Name)),
		},
	}

	// Assign the store
	store.Set(containerDesc, body)

	// No error
	return &containerDesc, nil
}

func getTemplateArchive(store StoreGetter, layerName string) ([]byte, error) {
	// Check arguments
	if store == nil {
		return nil, errors.New("can't extract archive from a nil store")
	}

	// Extract archive from image
	archiveDesc, archiveRaw, hasArchiveLayer := store.GetByName(layerName)
	if !hasArchiveLayer {
		return nil, fmt.Errorf("unable to lookup archive '%s' layer from crate", layerName)
	}
	if archiveDesc.MediaType != harpDataLayerMediaType {
		return nil, fmt.Errorf("invalid archive layer media type value for '%s'", layerName)
	}

	// No error
	return archiveRaw, nil
}

// AddTemplateArchive registers a new layer to the current store for the given archive.
func addTemplateArchive(store StoreSetter, ta *TemplateArchive) (*ocispec.Descriptor, error) {
	// Check arguments
	if types.IsNil(store) {
		return nil, errors.New("unable to register sealed container with nil storage")
	}
	if ta == nil {
		return nil, errors.New("given template archive is nil")
	}

	// Prepare a layer
	containerDesc := ocispec.Descriptor{
		MediaType: harpDataLayerMediaType,
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
