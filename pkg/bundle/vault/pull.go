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

package vault

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/vault/api"
	"golang.org/x/sync/errgroup"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle/vault/internal/operation"
	"github.com/elastic/harp/pkg/vault/kv"
	vpath "github.com/elastic/harp/pkg/vault/path"
)

// Pull all given path as a bundle.
func Pull(ctx context.Context, client *api.Client, paths []string, opts ...Option) (*bundlev1.Bundle, error) {
	// Check parameters
	if client == nil {
		return nil, fmt.Errorf("unable to process with nil client")
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no path given to pull")
	}

	// Default values
	var (
		defaultPrefix         = ""
		defaultPathInclusions = []*regexp.Regexp{}
		defaultPathExclusions = []*regexp.Regexp{}
		defaultWithMetadata   = false
	)

	// Create default option instance
	defaultOpts := &options{
		prefix:       defaultPrefix,
		exclusions:   defaultPathExclusions,
		includes:     defaultPathInclusions,
		withMetadata: defaultWithMetadata,
	}

	// Apply option functions
	for _, o := range opts {
		if err := o(defaultOpts); err != nil {
			return nil, err
		}
	}

	// Run the pull process
	b, err := runPull(ctx, client, paths, defaultOpts)
	if err != nil {
		return nil, fmt.Errorf("error occurs during pull process: %w", err)
	}

	// No error
	return b, nil
}

// runPull starts a multithreaded Vault secret puller.
//nolint:funlen // refactor
func runPull(ctx context.Context, client *api.Client, paths []string, opts *options) (*bundlev1.Bundle, error) {
	var res *bundlev1.Bundle

	// Initialize operation
	packageChan := make(chan *bundlev1.Package)

	// Prepare output
	g, gctx := errgroup.WithContext(ctx)

	// Preprocess paths
	if len(opts.exclusions) > 0 {
		paths = collect(paths, opts.exclusions, false)
	}
	if len(opts.includes) > 0 {
		paths = collect(paths, opts.includes, true)
	}

	// Fork consumer

	// Secret packages consumer
	g.Go(func() error {
		b := &bundlev1.Bundle{}

		// Wait for all packages
		for p := range packageChan {
			b.Packages = append(b.Packages, p)
		}

		// Assign result
		res = b

		// No error
		return nil
	})

	// Fork reader
	g.Go(func() error {
		defer close(packageChan)

		gReader, gReaderctx := errgroup.WithContext(gctx)

		// Wrap process in a builder to be able to pass p parameter
		exportBuilder := func(p string) func() error {
			return func() error {
				// Create dedicated service reader
				service, err := kv.New(client, p)
				if err != nil {
					return fmt.Errorf("unable to prepare vault reader for path '%s': %w", p, err)
				}

				// Create an exporter
				op := operation.Exporter(service, vpath.SanitizePath(p), packageChan, opts.withMetadata)

				// Run the job
				if err := op.Run(gReaderctx); err != nil {
					return fmt.Errorf("unable to export secret values for path `%s': %w", p, err)
				}

				// No error
				return nil
			}
		}

		// Generate producers
		for _, p := range paths {
			// For the process
			gReader.Go(exportBuilder(p))
		}

		// Wait for all producers to finish
		if err := gReader.Wait(); err != nil {
			return fmt.Errorf("unable to read secrets: %w", err)
		}

		// No error
		return nil
	})

	// Wait for completion
	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("unable to pull secrets: %w", err)
	}

	// Check bundle result
	if res == nil {
		return nil, fmt.Errorf("result bundle is nil")
	}

	// No error
	return res, nil
}
