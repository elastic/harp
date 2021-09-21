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

package kv

import (
	"context"
	"errors"
)

var (
	// ErrPathNotFound is raised when given secret path doesn't exists.
	ErrPathNotFound = errors.New("path not found")
	// ErrNoData is raised when gievn secret path doesn't contains data.
	ErrNoData = errors.New("no data")
	// ErrCustomMetadataDisabled is raised when trying to write a custom
	// metadata with globally disabled feature.
	ErrCustomMetadataDisabled = errors.New("custom metadata is disabled")
)

// VaultMetadataDataKey represents the secret data key used to store
// metadata.
var VaultMetadataDataKey = "www.vaultproject.io/kv/metadata"

const (
	// CustomMetadataKeyLimit defines the key count limit for custom metadata.
	CustomMetadataKeyLimit = 64
	// CustomMetadataKeySizeLimit defines the key size limit in bytes for
	// custom metadata.
	CustomMetadataKeySizeLimit = 128
	// CustomMetadataValueSizeLimit defines the value size limit in bytes for
	// custom metadata.
	CustomMetadataValueSizeLimit = 512
)

// SecretData is a secret body
type SecretData map[string]interface{}

// SecretMetadata is secret data attached metadata
type SecretMetadata map[string]interface{}

// SecretLister repesents secret key listing feature contract.
type SecretLister interface {
	List(ctx context.Context, path string) ([]string, error)
}

// SecretReader represents secret reader feature contract.
type SecretReader interface {
	Read(ctx context.Context, path string) (SecretData, SecretMetadata, error)
	ReadVersion(ctx context.Context, path string, version uint32) (SecretData, SecretMetadata, error)
}

// SecretWriter represents secret writer feature contract.
type SecretWriter interface {
	Write(ctx context.Context, path string, secrets SecretData) error
	WriteWithMeta(ctx context.Context, path string, secrets SecretData, meta SecretMetadata) error
}

// Service declares vault service contract.
type Service interface {
	SecretLister
	SecretReader
	SecretWriter
}
