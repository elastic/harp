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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

// S3 Backend object storage manager
func S3(client s3iface.S3API, bucket, prefix string) Backend {
	return &s3Backend{
		client: client,
		bucket: bucket,
		prefix: prefix,
	}
}

// -----------------------------------------------------------------------------

// s3Backend is a storage backend for Amazon S3
type s3Backend struct {
	client s3iface.S3API
	bucket string
	prefix string
}

// GetObject retrieves an object from Amazon S3 bucket, at prefix
func (b *s3Backend) GetObject(ctx context.Context, path string) (*Object, error) {
	// Check parameters
	if b.client == nil {
		return nil, errors.New("s3: client is nil")
	}

	// Prepare request
	input := &s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(pathutil.Join(b.prefix, path)),
	}

	// Get object from bucket
	result, err := b.client.GetObjectWithContext(ctx, input)
	if err != nil {
		var aerr awserr.Error
		if errors.As(err, &aerr) {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				return nil, fmt.Errorf("s3: object not found '%s'", *input.Key)
			default:
				return nil, fmt.Errorf("s3: unable to retrieve object: %w", aerr)
			}
		}
		return nil, fmt.Errorf("s3: unable to process request: %w", err)
	}

	// Assemble response
	var object Object
	object.Path = path
	object.Content = result.Body
	object.LastModified = *result.LastModified

	// No error
	return &object, nil
}
