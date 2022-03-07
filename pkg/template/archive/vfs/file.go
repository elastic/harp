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

package vfs

import (
	"bytes"
	"io/fs"
	"time"
)

type tarGzFile struct {
	name     string
	size     int64
	contents *bytes.Buffer
}

// Stat returns a fs.FileInfo for the given file.
func (f *tarGzFile) Stat() (fs.FileInfo, error) {
	return f, nil
}

// Read reads the next len(p) bytes from the buffer or until the buffer is drained.
func (f *tarGzFile) Read(buf []byte) (int, error) {
	return f.contents.Read(buf)
}

// Close is a no-op.
func (f *tarGzFile) Close() error {
	return nil
}

// Implementation of fs.FileInfo for tarGzFile.

// Name returns the basename of the file.
func (f *tarGzFile) Name() string {
	return f.name
}

// Size returns the length in bytes for this file.
func (f *tarGzFile) Size() int64 {
	return f.size
}

// Mode returns the mode for this file.
func (*tarGzFile) Mode() fs.FileMode {
	return 0o444
}

// Mode returns the mtime for this file (always the Unix epoch).
func (*tarGzFile) ModTime() time.Time {
	return time.Unix(0, 0)
}

// IsDir returns whether this file is a directory (always false).
func (*tarGzFile) IsDir() bool {
	return false
}

// Sys returns nil.
func (*tarGzFile) Sys() interface{} {
	return nil
}
