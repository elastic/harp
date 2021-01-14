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

package azblob

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

	storage "github.com/Azure/azure-sdk-for-go/storage"

	cloudstorage "github.com/elastic/harp/pkg/cloud/storage"
	serverstorage "github.com/elastic/harp/pkg/server/storage"
)

type engine struct {
	client     *storage.Client
	bucketName string
	prefix     string
}

func build(u *url.URL) (serverstorage.Engine, error) {
	// Check arguments
	if u == nil {
		return nil, fmt.Errorf("unable to prepare azblob with nil url")
	}

	q := u.Query()

	azureConnString := os.Getenv("AZURE_CONNECTION_STRING")
	if azureConnString == "" {
		return nil, errors.New("AZURE_CONNECTION_STRING env. variable must be set for azblob backend")
	}

	// Create an Azure Stroage client
	client, err := storage.NewClientFromConnectionString(azureConnString)
	if err != nil {
		return nil, fmt.Errorf("azblob: unable to initialize storage client: %w", err)
	}

	// Build engine instance
	return &engine{
		client:     &client,
		bucketName: u.Hostname(),
		prefix:     q.Get("prefix"),
	}, nil
}

func init() {
	// Register to storage factory
	serverstorage.MustRegister("azblob", build)
}

// Reader returns the file Reader
func (d *engine) Get(ctx context.Context, key string) ([]byte, error) {
	// Create an Azure Stroage client
	if d.client == nil {
		return nil, fmt.Errorf("azblob: unable proceed with nil client")
	}

	// Retrieve using Azure storage backend
	result, err := cloudstorage.AzureBlob(d.client, d.bucketName, d.prefix).GetObject(ctx, key)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("azblob: nil object returned")
	}

	// No error
	return ioutil.ReadAll(result.Content)
}
