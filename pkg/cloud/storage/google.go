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

package storage

import (
	"context"
	"errors"
	"fmt"
	pathutil "path"

	"cloud.google.com/go/storage"
)

// GCS Backend object storage manager
func GCS(client *storage.Client, bucket, prefix string) Backend {
	return &gcsBackend{
		client: client,
		bucket: bucket,
		prefix: prefix,
	}
}

// -----------------------------------------------------------------------------

// googleGCSBackend is a storage backend for Google Cloud Storage
type gcsBackend struct {
	client *storage.Client
	bucket string
	prefix string
}

// GetObject retrieves an object from Google Cloud Storage bucket, at prefix
func (b *gcsBackend) GetObject(ctx context.Context, path string) (*Object, error) {
	// Check parameters
	if b.client == nil {
		return nil, errors.New("gcs: client is nil")
	}

	// Query gcs bucket
	objectHandle := b.client.Bucket(b.bucket).Object(pathutil.Join(b.prefix, path))
	if objectHandle == nil {
		return nil, errors.New("gcs: unable to retrieve object reference")
	}

	// Retrieve object attributes
	attrs, err := objectHandle.Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("gcs: unable to retrieve object attribute: %w", err)
	}

	// Prepare content reader
	rc, err := objectHandle.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("gcs: unable to initialie object reader: %w", err)
	}

	// Assemble response
	var object Object
	object.Path = path
	object.Content = rc
	object.LastModified = attrs.Updated

	// No error
	return &object, nil
}
