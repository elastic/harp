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

package pipeline

import (
	"context"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
)

// Processor declares a bundle processor contract
type Processor func(context.Context, *bundlev1.Bundle) error

// Context defines tree processing context.
type Context interface {
	GetFile() *bundlev1.Bundle
	GetPackage() *bundlev1.Package
	GetSecret() *bundlev1.SecretChain
	GetKeyValue() *bundlev1.KV
}

// -----------------------------------------------------------------------------

// FileProcessorFunc describes a file object processor contract.
type FileProcessorFunc func(Context, *bundlev1.Bundle) error

// PackageProcessorFunc describes a package object processor contract.
type PackageProcessorFunc func(Context, *bundlev1.Package) error

// ChainProcessorFunc describes a secret chain object processor contract.
type ChainProcessorFunc func(Context, *bundlev1.SecretChain) error

// KVProcessorFunc describes a kv object processor contract.
type KVProcessorFunc func(Context, *bundlev1.KV) error
