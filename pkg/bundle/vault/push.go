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

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle/vault/internal/operation"
)

// Push the given bundle in Hashicorp Vault.
func Push(ctx context.Context, b *bundlev1.Bundle, client *api.Client, opts ...Option) error {
	// Check parameters
	if b == nil {
		return fmt.Errorf("unable to process nil bundle")
	}
	if client == nil {
		return fmt.Errorf("unable to process nil vault client")
	}

	// Default values
	var (
		defaultPrefix         = ""
		defaultPathInclusions = []*regexp.Regexp{}
		defaultPathExclusions = []*regexp.Regexp{}
		defaultWithMetadata   = false
		defaultWorkerCount    = int64(4)
	)

	// Create default option instance
	defaultOpts := &options{
		prefix:       defaultPrefix,
		exclusions:   defaultPathExclusions,
		includes:     defaultPathInclusions,
		withMetadata: defaultWithMetadata,
		workerCount:  defaultWorkerCount,
	}

	// Apply option functions
	for _, o := range opts {
		if err := o(defaultOpts); err != nil {
			return fmt.Errorf("unable to apply option: %w", err)
		}
	}

	// No error
	return runPush(ctx, b, client, defaultOpts)
}

func runPush(ctx context.Context, b *bundlev1.Bundle, client *api.Client, opts *options) error {
	// Prepare bundle
	if len(opts.includes) > 0 {
		filteredPackages := []*bundlev1.Package{}
		for _, p := range b.Packages {
			if matchPathRule(p.Name, opts.exclusions) {
				filteredPackages = append(filteredPackages, p)
			}
		}
		b.Packages = filteredPackages
	}
	if len(opts.exclusions) > 0 {
		filteredPackages := []*bundlev1.Package{}
		for _, p := range b.Packages {
			if !matchPathRule(p.Name, opts.exclusions) {
				filteredPackages = append(filteredPackages, p)
			}
		}
		b.Packages = filteredPackages
	}

	// Initialize operation
	op := operation.Importer(client, b, opts.prefix, opts.withMetadata, opts.workerCount)

	// Run the vault operation
	if err := op.Run(ctx); err != nil {
		return fmt.Errorf("unable to push secret bundle: %w", err)
	}

	// No error
	return nil
}
