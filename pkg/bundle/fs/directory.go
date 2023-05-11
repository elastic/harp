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

//go:build go1.16

package fs

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"sync"
	"time"
)

type directory struct {
	sync.RWMutex

	name     string
	perm     os.FileMode
	modTime  time.Time
	children map[string]interface{}
}

// Compile time type assertion
var _ fs.ReadDirFile = (*directory)(nil)

// -----------------------------------------------------------------------------

func (d *directory) Stat() (fs.FileInfo, error) {
	return &fileInfo{
		name:    d.name,
		size:    1,
		modTime: d.modTime,
		mode:    d.perm | fs.ModeDir,
	}, nil
}

//nolint:revive
func (d *directory) Read(b []byte) (int, error) {
	return 0, errors.New("is a directory")
}

func (d *directory) Close() error {
	return nil
}

func (d *directory) ReadDir(n int) ([]fs.DirEntry, error) {
	// Lock for read
	d.RLock()
	defer d.RUnlock()

	// Retrieve children entry count
	childrenNames := []string{}
	for entryName := range d.children {
		childrenNames = append(childrenNames, entryName)
	}

	// Apply read limit
	if n <= 0 {
		n = len(childrenNames)
	}

	// Iterate on children entities
	out := []fs.DirEntry{}
	for i := 0; i < len(childrenNames) && i < n; i++ {
		name := childrenNames[i]
		h := d.children[name]

		switch item := h.(type) {
		case *directory:
			out = append(out, &dirEntry{
				fi: &fileInfo{
					name: item.name,
					size: 1,
					mode: item.perm | os.ModeDir,
				},
			})
		case *file:
			out = append(out, &dirEntry{
				fi: &fileInfo{
					name:    item.name,
					size:    item.size,
					modTime: item.modTime,
					mode:    item.mode,
				},
			})
		default:
			continue
		}
	}

	// Check directory entry exhaustion
	if n > len(childrenNames) {
		return out, io.EOF
	}

	// Check empty response
	if len(out) == 0 {
		return out, errors.New("directory has no entry")
	}

	// Return result
	return out, nil
}
