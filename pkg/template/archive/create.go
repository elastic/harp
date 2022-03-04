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

package archive

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"

	"github.com/gobwas/glob"
	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/sdk/types"
)

// Create an archive from given options to the given writer.
//nolint:gocyclo,funlen // to refactor
func Create(fileSystem fs.FS, w io.Writer, opts ...CreateOption) error {
	// Check arguments
	if types.IsNil(fileSystem) {
		return errors.New("fileSystem is nil")
	}
	if types.IsNil(w) {
		return errors.New("output writer is nil")
	}

	// Prepare arguments
	dopts := &createOptions{
		rootPath:     ".",
		includeGlobs: []string{"**"},
		excludeGlobs: []string{},
	}
	for _, o := range opts {
		o(dopts)
	}

	// Ensure that the root path is valid
	if !fs.ValidPath(dopts.rootPath) {
		return fmt.Errorf("root path '%s' is not a valid path", dopts.rootPath)
	}

	// Ensure the root actually exists before trying to tar it
	rootFi, err := fs.Stat(fileSystem, dopts.rootPath)
	if err != nil {
		return fmt.Errorf("unable to tar files: %w", err)
	}
	if !rootFi.Mode().IsRegular() {
		return errors.New("root path is not a regular file")
	}
	if !rootFi.IsDir() {
		return errors.New("root path must be a directory")
	}

	// Compile inclusion filters
	var includes []glob.Glob
	for _, f := range dopts.includeGlobs {
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
	for _, f := range dopts.excludeGlobs {
		// Try to compile glob filter.
		filter, err := glob.Compile(f)
		if err != nil {
			return fmt.Errorf("unable to compile glob filter '%s' for exclusion: %w", f, err)
		}

		// Add to explusion filters.
		excludes = append(excludes, filter)
	}

	// Create writer chain.
	zr := gzip.NewWriter(w)
	tw := tar.NewWriter(zr)

	// walk through every file in the folder
	if errWalk := fs.WalkDir(fileSystem, dopts.rootPath, func(file string, dirEntry fs.DirEntry, errIn error) error {
		// return on any error
		if errIn != nil {
			return errIn
		}

		// ignore invalid file path
		if !fs.ValidPath(file) {
			log.Bg().Debug("ignoring invalid path file ...", zap.String("file", file))
			return nil
		}

		// Ignore non regular files
		if !dirEntry.Type().IsRegular() {
			log.Bg().Debug("ignoring irregular file ...", zap.String("file", file))
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
			log.Bg().Debug("ignoring file ...", zap.String("file", file))
			return nil
		}

		// Get FileInfo
		fi, err := dirEntry.Info()
		if err != nil {
			return fmt.Errorf("unable to retrieve fileInfo for '%s': %w", file, err)
		}

		// generate tar header
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return fmt.Errorf("unable to create TAR File header: %w", err)
		}

		// must provide real name
		header.Name = file

		// write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// if not a dir, write file content
		if !fi.IsDir() {
			data, err := fs.ReadFile(fileSystem, file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, bytes.NewReader(data)); err != nil {
				return err
			}
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
