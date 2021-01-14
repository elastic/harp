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

package s3

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"

	"github.com/elastic/harp/pkg/cloud/aws/session"
	cloudstorage "github.com/elastic/harp/pkg/cloud/storage"
	"github.com/elastic/harp/pkg/server/storage"
)

type engine struct {
	s3api      s3iface.S3API
	bucketName string
	basePath   string
}

func build(u *url.URL) (storage.Engine, error) {
	// Check arguments
	if u == nil {
		return nil, fmt.Errorf("unable to prepare s3 with nil url")
	}

	// Build session from url
	opts, err := session.FromURL(u.String())
	if err != nil {
		return nil, fmt.Errorf("unable to parse session URL: %w", err)
	}
	// Build AWS session
	sess, err := session.NewSession(opts)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize session: %w", err)
	}

	// Build engine instance
	return &engine{
		s3api:      s3.New(sess),
		bucketName: opts.BucketName,
	}, nil
}

func init() {
	// Register to storage factory
	storage.MustRegister("s3", build)
}

// -----------------------------------------------------------------------------

func (e *engine) Get(ctx context.Context, key string) ([]byte, error) {
	// Check fields
	if e.s3api == nil {
		return nil, fmt.Errorf("s3 service is nil")
	}
	if e.bucketName == "" {
		return nil, fmt.Errorf("bucketName is blank")
	}

	// Clean key
	key = strings.TrimPrefix(key, fmt.Sprintf("/%s/", e.bucketName))

	// Retrieve using S3 storage backend
	result, err := cloudstorage.S3(e.s3api, e.bucketName, e.basePath).GetObject(ctx, key)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("s3: nil object returned")
	}

	// No error
	return ioutil.ReadAll(result.Content)
}

// -----------------------------------------------------------------------------
