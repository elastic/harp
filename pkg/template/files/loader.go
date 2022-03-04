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
)

// BufferedFile represents an archive file buffered for later processing.
type BufferedFile struct {
	Name string
	Data []byte
}

// ContentLoader loads file content.
type ContentLoader interface {
	Load() ([]*BufferedFile, error)
}

// Loader returns a new BufferedFile list from given path name.
func Loader(filesystem fs.FS, name string) (ContentLoader, error) {
	// Check if it's a directory
	fi, err := fs.Stat(filesystem, name)
	if err != nil {
		return nil, fmt.Errorf("unable to get file info for '%s': %w", name, err)
	}

	// Is directory
	if fi.IsDir() {
		return &DirLoader{
			filesystem: filesystem,
			name:       name,
		}, nil
	}

	return nil, fmt.Errorf("only directory is supported as content loader")
}
