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

package targzfs

import (
	"bytes"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/gobwas/glob"
)

var (
	// Block decompression if the TAR archive is larger than 25MB.
	maxDecompressedSize = int64(25 * 1024 * 1024)
	// Block decompression if the archive has more than 10k files.
	maxFileCount = 10000
)

type tarGzFs struct {
	files       map[string]*tarEntry
	rootEntries []fs.DirEntry
	rootEntry   *tarEntry
}

var _ fs.FS = (*tarGzFs)(nil)

// Open opens the named file.
func (gzfs *tarGzFs) Open(name string) (fs.File, error) {
	// Shortcut if the file is '.'
	if name == "." {
		if gzfs.rootEntries == nil {
			return &rootFile{}, nil
		}
		return &tarFile{
			tarEntry:   *gzfs.rootEntry,
			r:          bytes.NewReader(gzfs.rootEntry.b),
			readDirPos: 0,
		}, nil
	}

	// Lookup file.
	f, err := gzfs.get(name, "open")
	if err != nil {
		return nil, err
	}

	// Wrapped file content
	return &tarFile{
		tarEntry:   *f,
		r:          bytes.NewReader(f.b),
		readDirPos: 0,
	}, nil
}

var _ fs.ReadDirFS = (*tarGzFs)(nil)

// ReadDir is used to enumerate all files from a directory.
func (gzfs *tarGzFs) ReadDir(name string) ([]fs.DirEntry, error) {
	// Shortcut if the file is '.'
	if name == "." {
		return gzfs.rootEntries, nil
	}

	// Lookup file.
	e, err := gzfs.get(name, "readdir")
	if err != nil {
		return nil, err
	}

	// Only directory should be used.
	if !e.IsDir() {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrInvalid}
	}

	// Sort results by name.
	sort.Slice(e.entries, func(i, j int) bool {
		return e.entries[i].Name() < e.entries[j].Name()
	})

	// Return file entries.
	return e.entries, nil
}

var _ fs.ReadFileFS = (*tarGzFs)(nil)

// ReadFile is used to retrieve directly the file content.
func (gzfs *tarGzFs) ReadFile(name string) ([]byte, error) {
	// Shortcut if the file is '.'
	if name == "." {
		return nil, &fs.PathError{Op: "readfile", Path: name, Err: fs.ErrInvalid}
	}

	// Lookup file.
	e, err := gzfs.get(name, "readfile")
	if err != nil {
		return nil, err
	}

	// Entry must be a file
	if e.IsDir() {
		return nil, &fs.PathError{Op: "readfile", Path: name, Err: fs.ErrInvalid}
	}

	// Copy content
	buf := make([]byte, len(e.b))
	copy(buf, e.b)

	// No error
	return buf, nil
}

var _ fs.StatFS = (*tarGzFs)(nil)

// Stat query the in-memory file system to get file info.
func (gzfs *tarGzFs) Stat(name string) (fs.FileInfo, error) {
	// Shortcut if the file is '.'
	if name == "." {
		if gzfs.rootEntry == nil {
			return &rootFile{}, nil
		}

		// Return root fileinfo
		return gzfs.rootEntry.Info()
	}

	// Lookup file.
	e, err := gzfs.get(name, "stat")
	if err != nil {
		return nil, err
	}

	// Return fileinfo
	return e.h.FileInfo(), nil
}

var _ fs.GlobFS = (*tarGzFs)(nil)

func (gzfs *tarGzFs) Glob(pattern string) (matches []string, _ error) {
	// Compile pattern
	g, err := glob.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("unable to compile glob pattern: %w", err)
	}

	// Iterate over file names
	for name := range gzfs.files {
		// Check if pattern match the file name
		if g.Match(name) {
			matches = append(matches, name)
		}
	}

	// Return results
	return
}

var _ fs.SubFS = (*tarGzFs)(nil)

func (gzfs *tarGzFs) Sub(dir string) (fs.FS, error) {
	if dir == "." {
		return gzfs, nil
	}

	// Lookup directory
	e, err := gzfs.get(dir, "sub")
	if err != nil {
		return nil, err
	}

	// Must be a directory
	if !e.IsDir() {
		return nil, &fs.PathError{Op: "sub", Path: dir, Err: fs.ErrInvalid}
	}

	// Create a sub-filesystem
	subfs := &tarGzFs{
		files:       make(map[string]*tarEntry),
		rootEntries: e.entries,
		rootEntry:   e,
	}

	// Copy files and remove directory prefix.
	prefix := dir + "/"
	for name, file := range gzfs.files {
		if strings.HasPrefix(name, prefix) {
			subfs.files[strings.TrimPrefix(name, prefix)] = file
		}
	}

	// No error
	return subfs, nil
}

// -----------------------------------------------------------------------------

func (gzfs *tarGzFs) get(name, op string) (*tarEntry, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: op, Path: name, Err: fs.ErrInvalid}
	}

	// Lookup file
	e, ok := gzfs.files[name]
	if !ok {
		return nil, &fs.PathError{Op: op, Path: name, Err: fs.ErrNotExist}
	}

	return e, nil
}
