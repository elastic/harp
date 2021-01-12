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

package container

import (
	"context"
	"errors"
	"fmt"
	"io"

	"cloud.google.com/go/storage"

	cloudstorage "github.com/elastic/harp/pkg/cloud/storage"
)

type gcsLoader struct {
	bucketName string
	prefix     string
}

// Reader returns the file Reader
func (d *gcsLoader) Reader(ctx context.Context, key string) (io.ReadCloser, error) {
	// Create a Google Storage client
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("gcs: unable to initialize storage client: %w", err)
	}

	// Retrieve using S3 storage backend
	result, err := cloudstorage.GCS(client, d.bucketName, d.prefix).GetObject(ctx, key)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("gcs: nil object returned")
	}

	// No error
	return result.Content, nil
}
