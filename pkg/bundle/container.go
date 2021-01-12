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

package bundle

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	containerv1 "github.com/elastic/harp/api/gen/go/harp/container/v1"
	"github.com/elastic/harp/pkg/container"
	"github.com/elastic/harp/pkg/sdk/types"
)

// Statistic hold bundle statistic information
type Statistic struct {
	PackageCount                 uint32
	CSOCompliantPackageNameCount uint32
	SecretCount                  uint32
}

const (
	bundleContentType    = "application/vnd.harp.v1.Bundle"
	gzipCompressionLevel = 9
)

// FromContainerReader returns a Bundle extracted from a secret container.
func FromContainerReader(r io.Reader) (*bundlev1.Bundle, error) {
	// Check parameters
	if types.IsNil(r) {
		return nil, fmt.Errorf("unable to process nil reader")
	}

	// Load secret container
	c, err := container.Load(r)
	if err != nil {
		return nil, fmt.Errorf("unable to load Bundle: %w", err)
	}

	// Delegate to bundle loader
	return FromContainer(c)
}

// ToContainerWriter returns a Bundle packaged as a secret container.
func ToContainerWriter(w io.Writer, b *bundlev1.Bundle) error {
	// Check parameters
	if types.IsNil(w) {
		return fmt.Errorf("unable to process nil writer")
	}
	if b == nil {
		return fmt.Errorf("unable to process nil bundle")
	}

	// Create a container
	c, err := ToContainer(b)
	if err != nil {
		return fmt.Errorf("unable to wrap bundle as a container: %w", err)
	}

	// Prepare secret container
	return container.Dump(w, c)
}

// FromContainer unwraps a Bundle from a secret container.
func FromContainer(c *containerv1.Container) (*bundlev1.Bundle, error) {
	// Check parameters
	if types.IsNil(c) {
		return nil, fmt.Errorf("unable to process nil container")
	}

	// Check headers
	if types.IsNil(c.Headers) {
		return nil, fmt.Errorf("unable to process nil container headers")
	}
	if c.Headers.ContentType != bundleContentType {
		return nil, fmt.Errorf("invalid content type for Bundle loader")
	}
	if c.Headers.ContentEncoding != "gzip" {
		return nil, fmt.Errorf("invalid content encoding for Bundle loader")
	}

	// Decompress bundle
	zr, err := gzip.NewReader(bytes.NewReader(c.Raw))
	if err != nil {
		return nil, fmt.Errorf("unable to initialize compression reader")
	}

	// Delegate to bundle loader
	return Load(zr)
}

// ToContainer wrpas a Bundle as a container object.
func ToContainer(b *bundlev1.Bundle) (*containerv1.Container, error) {
	if b == nil {
		return nil, fmt.Errorf("unable to process nil bundle")
	}

	// Dump bundle
	payload := &bytes.Buffer{}

	// Compress with gzip
	zw, errGz := gzip.NewWriterLevel(payload, gzipCompressionLevel)
	if errGz != nil {
		return nil, fmt.Errorf("unable to compress bundle content")
	}

	// Delegate to Bundle Writer
	if errDump := Dump(zw, b); errDump != nil {
		return nil, errDump
	}

	// Close gzip writer
	if errGz = zw.Close(); errGz != nil {
		return nil, fmt.Errorf("unable to close compression writer")
	}

	// Return container
	return &containerv1.Container{
		Headers: &containerv1.Header{
			ContentEncoding: "gzip",
			ContentType:     bundleContentType,
		},
		Raw: payload.Bytes(),
	}, nil
}
