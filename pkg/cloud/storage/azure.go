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
	"time"

	msstorage "github.com/Azure/azure-sdk-for-go/storage"
)

// AzureBlob Backend object storage manager
func AzureBlob(client *msstorage.Client, bucket, prefix string) Backend {
	return &msAzureBlobBackend{
		client: client,
		bucket: bucket,
		prefix: prefix,
	}
}

// -----------------------------------------------------------------------------

// msAzureBlobBackend is a storage backend for Microsoft Azure Blob Storage
type msAzureBlobBackend struct {
	client *msstorage.Client
	bucket string
	prefix string
}

// GetObject retrieves an object from Microsoft Azure Blob Storage, at path
func (b *msAzureBlobBackend) GetObject(ctx context.Context, path string) (*Object, error) {
	// Check arguments
	if b.client == nil {
		return nil, errors.New("azure: unable to obtain a client reference")
	}

	// Retrieve blob service
	blobSrv := b.client.GetBlobService()

	// Retrieve container
	container := blobSrv.GetContainerReference(b.bucket)
	if container == nil {
		return nil, errors.New("azure: unable to obtain a container reference")
	}

	// Compute object path
	objectPath := pathutil.Join(b.prefix, path)

	// Check if object exists
	blobReference := container.GetBlobReference(objectPath)
	if blobReference == nil {
		return nil, fmt.Errorf("azure: unable to retrieve blob reference for '%s'", objectPath)
	}

	// Check existence
	exists, err := blobReference.Exists()
	if err != nil {
		return nil, fmt.Errorf("azure: unable to check blob existence for '%s': %w", objectPath, err)
	}
	if !exists {
		return nil, fmt.Errorf("azure: object '%s' does not exist", objectPath)
	}

	readCloser, err := blobReference.Get(nil)
	if err != nil {
		return nil, fmt.Errorf("azure: unbale to oper content reader for '%s': %w", objectPath, err)
	}

	// Assemble response
	var object Object
	object.Path = path
	object.Content = readCloser
	object.LastModified = time.Time(blobReference.Properties.LastModified)

	// No error
	return &object, nil
}
