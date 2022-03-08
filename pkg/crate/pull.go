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

package crate

import (
	"context"
	"errors"
	"fmt"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	orascontext "oras.land/oras-go/pkg/context"
	"oras.land/oras-go/pkg/oras"
	"oras.land/oras-go/pkg/target"

	"github.com/elastic/harp/pkg/sdk/types"
)

// Pull the given image descriptor from the given repository.
func Pull(ctx context.Context, from target.Target, imageRef string, to target.Target) (*ocispec.Descriptor, error) {
	// Check arguments
	if types.IsNil(from) {
		return nil, errors.New("from must not be nil")
	}
	if imageRef == "" {
		return nil, errors.New("image reference must not be blank")
	}

	// Pull the image
	containerManifest, err := oras.Copy(orascontext.WithLoggerDiscarded(ctx), from, imageRef, to, "", oras.WithAllowedMediaTypes([]string{
		harpConfigMediaType,
		harpContainerLayerMediaType,
		harpDataLayerMediaType,
	}))
	if err != nil {
		return nil, fmt.Errorf("unable to pull image: %w", err)
	}

	// No error
	return &containerManifest, err
}
