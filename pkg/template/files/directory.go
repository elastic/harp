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

package files

import (
	"fmt"
	"io/fs"
	"path/filepath"
)

// DirLoader loads a chart from a directory
type DirLoader struct {
	filesystem fs.FS
	name       string
}

// Load loads the chart
func (l DirLoader) Load() ([]*BufferedFile, error) {
	return LoadDir(l.filesystem, l.name)
}

// LoadDir loads from a directory.
//
// This loads charts only from directories.
func LoadDir(filesystem fs.FS, dir string) ([]*BufferedFile, error) {
	// Check if path is valid
	if !fs.ValidPath(dir) {
		return nil, fmt.Errorf("'%s' is not a valid path", dir)
	}

	result := []*BufferedFile{}
	topdir := dir

	walk := func(name string, d fs.DirEntry, errWalk error) error {
		// Check walk error
		if errWalk != nil {
			return errWalk
		}

		// Compute relative path
		n, err := filepath.Rel(topdir, name)
		if err != nil {
			return fmt.Errorf("unable to compute relative path: %w", err)
		}
		if n == "" {
			return nil
		}

		// Normalize filepath
		n = filepath.ToSlash(n)

		// Ignore if it is a directory
		if d.IsDir() {
			return nil
		}

		// Irregular files include devices, sockets, and other uses of files that
		// are not regular files.
		if !d.Type().IsRegular() {
			return fmt.Errorf("cannot load irregular file %s as it has file mode type bits set", name)
		}

		// Read file content
		data, err := fs.ReadFile(filesystem, name)
		if err != nil {
			return fmt.Errorf("error reading %s: %w", name, err)
		}

		// Append to result
		result = append(result, &BufferedFile{Name: n, Data: data})

		// No error
		return nil
	}
	if err := fs.WalkDir(filesystem, topdir, walk); err != nil {
		return nil, fmt.Errorf("unable to walk directory '%s' : %w", topdir, err)
	}

	// No error
	return result, nil
}
