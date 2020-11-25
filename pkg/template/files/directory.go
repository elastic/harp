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
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// DirLoader loads a chart from a directory
type DirLoader struct {
	fs   afero.Fs
	name string
}

// Load loads the chart
func (l DirLoader) Load() ([]*BufferedFile, error) {
	return LoadDir(l.fs, l.name)
}

// LoadDir loads from a directory.
//
// This loads charts only from directories.
func LoadDir(fs afero.Fs, dir string) ([]*BufferedFile, error) {
	// Retrieve absolute path
	topdir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	result := []*BufferedFile{}
	topdir += string(filepath.Separator)

	walk := func(name string, fi os.FileInfo, errWalk error) error {
		// Check walk error
		if errWalk != nil {
			return errWalk
		}

		n := strings.TrimPrefix(name, topdir)
		if n == "" {
			return nil
		}

		// Normalize filepath
		n = filepath.ToSlash(n)

		// Ignore if it is a directory
		if fi.IsDir() {
			return nil
		}

		// Irregular files include devices, sockets, and other uses of files that
		// are not regular files.
		if !fi.Mode().IsRegular() {
			return fmt.Errorf("cannot load irregular file %s as it has file mode type bits set", name)
		}

		// Read file content
		data, err := afero.ReadFile(fs, name)
		if err != nil {
			return fmt.Errorf("error reading %s: %v", name, err)
		}

		// Append to result
		result = append(result, &BufferedFile{Name: n, Data: data})

		// No error
		return nil
	}
	if err := afero.Walk(fs, topdir, walk); err != nil {
		return nil, err
	}

	// No error
	return result, nil
}
