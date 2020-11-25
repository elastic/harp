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
	"io"
	"os"

	"github.com/spf13/afero"
)

// Compile time type assertion check
var _ afero.File = (*secretFile)(nil)

type secretFile struct {
	name    string
	content io.Reader
	size    int64
}

func (f *secretFile) Name() string {
	return f.name
}

func (f *secretFile) Close() error {
	return nil
}

func (f *secretFile) Read(p []byte) (n int, err error) {
	return f.content.Read(p)
}

func (f *secretFile) Stat() (os.FileInfo, error) {
	return &secretFileInfo{
		name: f.Name(),
		size: f.size,
	}, nil
}

// -----------------------------------------------------------------------------

func (f *secretFile) ReadAt(p []byte, off int64) (n int, err error) {
	return 0, ErrNotSeekable
}

func (f *secretFile) Seek(offset int64, whence int) (int64, error) {
	return 0, ErrNotSeekable
}

func (f *secretFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, ErrDirectoryListingIsForbidden
}

func (f *secretFile) Readdirnames(n int) ([]string, error) {
	return nil, ErrDirectoryListingIsForbidden
}

func (f *secretFile) Write(p []byte) (n int, err error) {
	return 0, ErrReadOnly
}

func (f *secretFile) WriteAt(p []byte, off int64) (n int, err error) {
	return 0, ErrReadOnly
}

func (f *secretFile) Sync() error {
	return nil
}

func (f *secretFile) Truncate(size int64) error {
	return ErrReadOnly
}

func (f *secretFile) WriteString(s string) (ret int, err error) {
	return 0, ErrReadOnly
}
