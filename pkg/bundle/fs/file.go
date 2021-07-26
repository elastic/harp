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
// +build go1.16

package fs

import (
	"io"
	"io/fs"
	"os"
	"time"

	"github.com/awnumar/memguard"
)

type file struct {
	modTime    time.Time
	name       string
	bodyReader io.Reader
	size       int64
	content    *memguard.Enclave
	mode       os.FileMode
	closed     bool
}

// Compile time type assertion
var _ fs.File = (*file)(nil)

// -----------------------------------------------------------------------------

func (f *file) Stat() (fs.FileInfo, error) {
	// Check file state
	if f.closed {
		return nil, fs.ErrClosed
	}

	// Return file information
	return &fileInfo{
		name:    f.name,
		size:    f.size,
		modTime: f.modTime,
		mode:    f.mode,
	}, nil
}

func (f *file) Read(b []byte) (int, error) {
	// Check file state
	if f.closed || f.bodyReader == nil {
		return 0, fs.ErrClosed
	}

	// Delegate to reader
	return f.bodyReader.Read(b)
}

func (f *file) Close() error {
	if f.closed {
		return fs.ErrClosed
	}
	f.closed = true
	f.bodyReader = nil
	return nil
}
