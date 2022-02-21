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

package compression

import (
	"compress/lzw"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/klauspost/compress/flate"
	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/s2"
	"github.com/klauspost/compress/zlib"
	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4"
	"github.com/ulikunitz/xz"
)

// -----------------------------------------------------------------------------

// NewReader returns a writer implementation according to given algorithm.
func NewReader(r io.Reader, algorithm string) (io.ReadCloser, error) {
	// Normalize input
	algorithm = strings.TrimSpace(strings.ToLower(algorithm))

	var (
		compressedReader io.ReadCloser
		readerErr        error
	)

	// Apply transformation
	switch algorithm {
	case "identity":
		compressedReader = ioutil.NopCloser(r)
	case "gzip":
		compressedReader, readerErr = gzip.NewReader(r)
	case "lzw", "lzw-lsb":
		compressedReader = lzw.NewReader(r, lzw.LSB, 8)
	case "lzw-msb":
		compressedReader = lzw.NewReader(r, lzw.MSB, 8)
	case "lz4":
		compressedReader = io.NopCloser(lz4.NewReader(r))
	case "s2", "snappy":
		compressedReader = io.NopCloser(s2.NewReader(r))
	case "zlib":
		compressedReader, readerErr = zlib.NewReader(r)
	case "flate", "deflate":
		compressedReader = flate.NewReader(r)
	case "lzma":
		reader, err := xz.NewReader(r)
		if err != nil {
			readerErr = err
		} else {
			compressedReader = io.NopCloser(reader)
		}
	case "zstd":
		reader, err := zstd.NewReader(r)
		if err != nil {
			readerErr = err
		} else {
			compressedReader = reader.IOReadCloser()
		}
	default:
		return nil, fmt.Errorf("unhandled compression algorithm '%s'", algorithm)
	}
	if readerErr != nil {
		return nil, fmt.Errorf("unable to initialize '%s' compressor: %w", algorithm, readerErr)
	}

	// No error
	return compressedReader, nil
}
