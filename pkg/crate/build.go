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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gobwas/glob"
	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/container"
	"github.com/elastic/harp/pkg/crate/cratefile"
	schemav1 "github.com/elastic/harp/pkg/crate/schema/v1"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/sdk/types"
)

const (
	maxContainerSize = 25 * 1024 * 1024
)

// Build a crate from the given specification.
func Build(spec *cratefile.Config) (*Image, error) {
	// Check arguments
	if spec == nil {
		return nil, errors.New("unable to build a crate with nil specification")
	}

	// Open container file
	cf, err := os.Open(spec.Container.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to open container '%s': %w", spec.Container.Path, err)
	}

	// Try to load container
	c, err := container.Load(io.LimitReader(cf, maxContainerSize))
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
	for i, archive := range spec.Archives {
		// Create tar.gz archive
		var buf bytes.Buffer
		if err := createArchive(&spec.Archives[i], &buf); err != nil {
			return nil, fmt.Errorf("unable to create archive '%s': %w", archive.Name, err)
		}

		// Add archive format suffix
		name := fmt.Sprintf("%s.tar.gz", archive.Name)

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

//nolint:gocyclo,funlen // to refactor
func createArchive(archive *cratefile.Archive, w io.Writer) error {
	// Check arguments
	if types.IsNil(w) {
		return errors.New("output writer is nil")
	}
	if archive == nil {
		return errors.New("archive is nil")
	}

	// Ensure the root actually exists before trying to tar it
	if _, err := os.Stat(archive.RootPath); err != nil {
		return fmt.Errorf("unable to tar files: %w", err)
	}

	// Create writer chain.
	zr := gzip.NewWriter(w)
	tw := tar.NewWriter(zr)

	// Compile inclusion filters
	var includes []glob.Glob
	for _, f := range archive.IncludeGlob {
		// Try to compile glob filter.
		filter, err := glob.Compile(f)
		if err != nil {
			return fmt.Errorf("unable to compile glob filter '%s' for inclusion: %w", f, err)
		}

		// Add to explusion filters.
		includes = append(includes, filter)
	}

	// Compile explusion filters
	var excludes []glob.Glob
	for _, f := range archive.ExcludeGlob {
		// Try to compile glob filter.
		filter, err := glob.Compile(f)
		if err != nil {
			return fmt.Errorf("unable to compile glob filter '%s' for exclusion: %w", f, err)
		}

		// Add to explusion filters.
		excludes = append(excludes, filter)
	}

	// walk through every file in the folder
	if errWalk := filepath.Walk(archive.RootPath, func(file string, fi os.FileInfo, errIn error) error {
		// return on any error
		if errIn != nil {
			return errIn
		}
		// Ignore non regular files
		if !fi.Mode().IsRegular() {
			return nil
		}

		// Process inclusions
		keep := false
		for _, f := range includes {
			if f.Match(file) {
				keep = true
			}
		}
		for _, f := range excludes {
			if f.Match(file) {
				keep = false
			}
		}
		if !keep {
			// Ignore this file.
			log.Bg().Debug("ignoreing file ...", zap.String("file", file))
			return nil
		}

		// generate tar header
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		// must provide real name
		header.Name, _ = filepath.Rel(archive.RootPath, file)

		// write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// if not a dir, write file content
		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, data); err != nil {
				return err
			}

			// Manual close to prevent to wait all files to be processed
			// to close all.
			log.SafeClose(data, "unble to close file", zap.String("file", file))
		}

		// No error
		return nil
	}); errWalk != nil {
		return fmt.Errorf("fail to walk folders for archive compression: %w", errWalk)
	}

	// produce tar
	if err := tw.Close(); err != nil {
		return err
	}

	// produce gzip
	if err := zr.Close(); err != nil {
		return err
	}

	// No error
	return nil
}
