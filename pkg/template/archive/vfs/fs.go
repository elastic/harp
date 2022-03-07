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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
)

var (
	// Block decompression if the TAR archive is larger than 25MB.
	maxDecompressedSize = int64(25 * 1024 * 1024)
	// Block decompression if the archive has more than 10k files.
	maxFileCount = 10000
)

type tarGzFs struct {
	files map[string][]byte
}

// FromArchive exposes the contents of the given reader (which is a .tar.gz file)
// as an fs.FS.
func FromArchive(r io.Reader) (fs.FS, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("unable to open .tar.gz file: %w", err)
	}

	// Retrieve TAR content from GZIP
	var (
		tarContents      bytes.Buffer
		tarContentLength = int64(0)
	)

	// Chunked read with hard limit to prevent/reduce zipbomb vulnerability
	// exploitation.
	for {
		written, err := io.CopyN(&tarContents, gz, 1024)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		// Add to length
		tarContentLength += written

		// Check max size
		if tarContentLength > maxDecompressedSize {
			return nil, errors.New("the archive contains a too large content (>25MB)")
		}
	}

	// Close the gzip decompressor
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("unable to close gzip reader: %w", err)
	}

	// TAR format reader
	tarReader := tar.NewReader(bytes.NewBuffer(tarContents.Bytes()))

	// Prepare in-memory filesystem.
	var ret tarGzFs
	ret.files = make(map[string][]byte)
	for {
		// Iterate on each file entry
		hdr, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("unable to read .tar.gz entry: %w", err)
		}

		// Load content in memory
		var fileContents bytes.Buffer
		if _, err := io.Copy(&fileContents, io.LimitReader(tarReader, maxDecompressedSize)); err != nil {
			return nil, fmt.Errorf("unable to read .tar.gz entry: %w", err)
		}

		// Append to in-memory map
		ret.files[hdr.Name] = fileContents.Bytes()

		// Check file count limit
		if len(ret.files) > maxFileCount {
			return nil, errors.New("interrupted extraction, too many files in the archive")
		}
	}

	// No error
	return &ret, nil
}

// Open opens the named file.
func (gzfs *tarGzFs) Open(name string) (fs.File, error) {
	contents, ok := gzfs.files[name]
	if !ok {
		return nil, os.ErrNotExist
	}

	// Wrapped file content
	return &tarGzFile{
		name:     name,
		contents: bytes.NewBuffer(contents),
		size:     int64(len(contents)),
	}, nil
}
