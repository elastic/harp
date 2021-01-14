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

package gcs

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

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"

	"cloud.google.com/go/storage"

	cloudstorage "github.com/elastic/harp/pkg/cloud/storage"
	serverstorage "github.com/elastic/harp/pkg/server/storage"
)

type engine struct {
	bucketName string
	prefix     string
}

func build(u *url.URL) (serverstorage.Engine, error) {
	// Check arguments
	if u == nil {
		return nil, fmt.Errorf("unable to prepare gcs with nil url")
	}

	q := u.Query()

	// Build engine instance
	return &engine{
		bucketName: u.Hostname(),
		prefix:     q.Get("prefix"),
	}, nil
}

func init() {
	// Register to storage factory
	serverstorage.MustRegister("gcs", build)
}

// Get returns the file Reader
func (d *engine) Get(ctx context.Context, key string) ([]byte, error) {
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
	return ioutil.ReadAll(result.Content)
}
