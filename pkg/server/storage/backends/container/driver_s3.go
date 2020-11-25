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
	"strings"

	"github.com/aws/aws-sdk-go/service/s3/s3iface"

	"github.com/elastic/harp/pkg/cloud/storage"
)

type s3Loader struct {
	s3api      s3iface.S3API
	bucketName string
	prefix     string
}

// Reader returns the file Reader
func (d *s3Loader) Reader(ctx context.Context, key string) (io.ReadCloser, error) {
	// Check fields
	if d.s3api == nil {
		return nil, fmt.Errorf("s3 service is nil")
	}
	if d.bucketName == "" {
		return nil, fmt.Errorf("bucktName is blank")
	}

	// Clean key
	key = strings.TrimPrefix(key, fmt.Sprintf("/%s/", d.bucketName))

	// Retrieve using S3 storage backend
	result, err := storage.S3(d.s3api, d.bucketName, d.prefix).GetObject(ctx, key)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("s3: nil object returned")
	}

	// No error
	return result.Content, nil
}
