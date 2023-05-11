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
	"archive/tar"
	"io"
	"io/fs"
	"time"
)

type tarEntry struct {
	h *tar.Header
	b []byte

	// If the file is a directory, the following property will contain
	// owned files.
	entries []fs.DirEntry
}

// Implementation of fs.DirEntry for tarEntry.
var _ fs.DirEntry = (*tarEntry)(nil)

// Name returns the basename of the file.
func (f *tarEntry) Name() string {
	return f.h.FileInfo().Name()
}

// Type returns file mode.
func (f *tarEntry) Type() fs.FileMode {
	return f.h.FileInfo().Mode() & fs.ModeType
}

// IsDir returns whether this file is a directory (always false).
func (f *tarEntry) IsDir() bool {
	return f.h.FileInfo().IsDir()
}

// Info returns file info.
func (f *tarEntry) Info() (fs.FileInfo, error) {
	return f.h.FileInfo(), nil
}

type tarFile struct {
	tarEntry
	r          io.ReadSeeker
	readDirPos int
}

// Implementation of fs.File for tarFile.
var _ fs.File = (*tarFile)(nil)

// Stat returns a fs.FileInfo for the given file.
func (f *tarFile) Stat() (fs.FileInfo, error) {
	return f.h.FileInfo(), nil
}

// Read reads the next len(p) bytes from the buffer or until the buffer is drained.
func (f *tarFile) Read(buf []byte) (int, error) {
	if f.IsDir() {
		return 0, &fs.PathError{Op: "read", Path: f.Name(), Err: fs.ErrInvalid}
	}

	return f.r.Read(buf)
}

// Close is a no-op.
func (f *tarFile) Close() error {
	return nil
}

// Implementation of io.Seeker for tarFile.
var _ io.Seeker = (*tarFile)(nil)

func (f *tarFile) Seek(offset int64, whence int) (int64, error) {
	if f.IsDir() {
		return 0, &fs.PathError{Op: "seek", Path: f.Name(), Err: fs.ErrInvalid}
	}

	return f.r.Seek(offset, whence)
}

// Implementation of fs.ReadDirFile for tarFile.
var _ fs.ReadDirFile = (*tarFile)(nil)

func (f *tarFile) ReadDir(n int) ([]fs.DirEntry, error) {
	if !f.IsDir() {
		return nil, &fs.PathError{Op: "readdir", Path: f.Name(), Err: fs.ErrInvalid}
	}

	if f.readDirPos >= len(f.entries) {
		if n <= 0 {
			return nil, nil
		}
		return nil, io.EOF
	}

	var entries []fs.DirEntry

	if n > 0 && f.readDirPos+n <= len(f.entries) {
		entries = f.entries[f.readDirPos : f.readDirPos+n]
		f.readDirPos += n
	} else {
		entries = f.entries[f.readDirPos:]
		f.readDirPos += len(entries)
	}

	return entries, nil
}

// Implmentation of fs.FileInfo for tarFile.
var _ fs.FileInfo = (*tarFile)(nil)

// Size returns the length in bytes for this file.
func (f *tarFile) Size() int64 {
	return f.h.Size
}

// Mode returns the mode for this file.
func (f *tarFile) Mode() fs.FileMode {
	return f.h.FileInfo().Mode()
}

// Mode returns the mtime for this file (always the Unix epoch).
func (f *tarFile) ModTime() time.Time {
	return f.h.ModTime
}

// Sys returns nil.
func (f *tarFile) Sys() interface{} {
	return nil
}

// -----------------------------------------------------------------------------

type rootFile struct{}

var _ fs.File = (*rootFile)(nil)

func (rf *rootFile) Stat() (fs.FileInfo, error) {
	return rf, nil
}

func (*rootFile) Read([]byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: ".", Err: fs.ErrInvalid}
}

func (*rootFile) Close() error {
	return nil
}

var _ fs.FileInfo = (*rootFile)(nil)

func (rf *rootFile) Name() string {
	return "."
}

func (rf *rootFile) Size() int64 {
	return 0
}

func (rf *rootFile) Mode() fs.FileMode {
	return fs.ModeDir | 0o755
}

func (rf *rootFile) ModTime() time.Time {
	return time.Time{}
}

func (rf *rootFile) IsDir() bool {
	return true
}

func (rf *rootFile) Sys() interface{} {
	return nil
}
