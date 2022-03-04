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
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/fs"

	"github.com/elastic/harp/pkg/container"
	"github.com/elastic/harp/pkg/crate/cratefile"
	schemav1 "github.com/elastic/harp/pkg/crate/schema/v1"
	"github.com/elastic/harp/pkg/template/archive"
)

const (
	maxContainerSize = 25 * 1024 * 1024
)

// Build a crate from the given specification.
func Build(rootFs fs.FS, spec *cratefile.Config) (*Image, error) {
	// Check arguments
	if spec == nil {
		return nil, errors.New("unable to build a crate with nil specification")
	}

	// Open container file
	cf, err := fs.ReadFile(rootFs, spec.Container.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to open container '%s': %w", spec.Container.Path, err)
	}

	// Try to load container
	c, err := container.Load(io.LimitReader(bytes.NewBuffer(cf), maxContainerSize))
	if err != nil {
		return nil, fmt.Errorf("unable to load input container '%s': %w", spec.Container.Path, err)
	}

	// Check container sealing status
	if !container.IsSealed(c) {
		// Seal with appropriate algorithm
		sc, err := container.Seal(rand.Reader, c, spec.Container.Identities...)
		if err != nil {
			return nil, fmt.Errorf("unable to seal container '%s': %w", spec.Container.Path, err)
		}

		// Replace container instance by the sealed one
		c = sc
	}

	// Prepare config manifest
	cfg := schemav1.NewConfig()
	cfg.SetContainers([]string{spec.Container.Name})

	// Preapre archives
	templateArchives := []*TemplateArchive{}

	// Create archives
	archiveNames := []string{}
	for _, arch := range spec.Archives {
		// Restrict sub filesystem
		subFs, errFs := fs.Sub(rootFs, arch.RootPath)
		if errFs != nil {
			return nil, fmt.Errorf("unable to restrict to sub-filesystem: %w", errFs)
		}

		// Create tar.gz archive
		var buf bytes.Buffer
		if err := archive.Create(subFs, &buf,
			archive.ExcludeFiles(arch.ExcludeGlob...),
			archive.IncludeFiles(arch.IncludeGlob...),
		); err != nil {
			return nil, fmt.Errorf("unable to create archive '%s': %w", arch.Name, err)
		}

		// Add archive format suffix
		name := fmt.Sprintf("%s.tar.gz", arch.Name)

		// Add to config manifest
		archiveNames = append(archiveNames, name)

		// Add layer.
		templateArchives = append(templateArchives, &TemplateArchive{
			Name:    name,
			Archive: buf.Bytes(),
		})
	}

	// Update config manifest
	cfg.SetTemplates(archiveNames)

	// Create
	res := &Image{
		Config: cfg,
		Containers: []*SealedContainer{
			{
				Name:      spec.Container.Name,
				Container: c,
			},
		},
		TemplateArchives: templateArchives,
	}

	// No error
	return res, nil
}

// -----------------------------------------------------------------------------
