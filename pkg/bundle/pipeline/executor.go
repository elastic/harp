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
	"fmt"
	"os"

	"github.com/gosimple/slug"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/build/version"
	"github.com/elastic/harp/pkg/bundle"
	"github.com/elastic/harp/pkg/sdk/log"
)

// Run a processor.
func Run(ctx context.Context, name string, opts ...Option) error {
	// Initialize a running context to attach all goroutines
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Initialize logger
	log.Setup(ctx,
		&log.Options{
			Debug:    true,
			AppName:  slug.Make(name),
			AppID:    version.ID(),
			Version:  version.Version,
			Revision: version.Commit,
		},
	)

	// Read bundle from Stdin
	b, err := bundle.FromContainerReader(os.Stdin)
	if err != nil {
		return fmt.Errorf("unable to read bundle from stdin: %w", err)
	}

	const (
		defaultDisableOutput = false
	)

	v := &bundleVisitor{
		ctx: ctx,
		opts: &Options{
			disableOutput: defaultDisableOutput,
		},
		position: &defaultContext{},
	}

	// Loop through each option
	for _, opt := range opts {
		opt(v.opts)
	}

	// Apply remapping strategy
	v.VisitForFile(b)

	// Check error
	if err := v.Error(); err != nil {
		return fmt.Errorf("error during bundle processing: %w", err)
	}

	if !v.opts.disableOutput {
		// Write output bundle
		if err := bundle.ToContainerWriter(os.Stdout, b); err != nil {
			return fmt.Errorf("unable to dump remapped bundle content: %w", err)
		}
	}

	// No error
	return nil
}

// -----------------------------------------------------------------------------

// Context is used to pass current node location to processor
type defaultContext struct {
	File    *bundlev1.Bundle
	Package *bundlev1.Package
	Secret  *bundlev1.SecretChain
	KV      *bundlev1.KV
}

func (c *defaultContext) GetFile() *bundlev1.Bundle        { return c.File }
func (c *defaultContext) GetPackage() *bundlev1.Package    { return c.Package }
func (c *defaultContext) GetSecret() *bundlev1.SecretChain { return c.Secret }
func (c *defaultContext) GetKeyValue() *bundlev1.KV        { return c.KV }

// -----------------------------------------------------------------------------

type bundleVisitor struct {
	ctx      context.Context
	err      error
	opts     *Options
	position *defaultContext
}

func (bv *bundleVisitor) Error() error {
	return bv.err
}

func (bv *bundleVisitor) VisitForFile(obj *bundlev1.Bundle) {
	// Check argument
	if obj == nil {
		bv.err = fmt.Errorf("unable to process nil file")
		return
	}

	// Update position
	bv.position.File = obj

	// Crawl packages
	for _, p := range obj.Packages {
		bv.VisitForPackage(p)
	}

	// If processor given use it
	if bv.opts.fpf != nil {
		if bv.err = bv.opts.fpf(bv.position, obj); bv.err != nil {
			return
		}
	}
}

func (bv *bundleVisitor) VisitForPackage(obj *bundlev1.Package) {
	// Check argument
	if obj == nil {
		bv.err = fmt.Errorf("unable to process nil package")
		return
	}

	// Update position
	bv.position.Package = obj

	// If package has secrets
	if obj.Secrets != nil {
		bv.VisitForChain(obj.Secrets)
	}

	// If processor given use it
	if bv.opts.ppf != nil {
		if bv.err = bv.opts.ppf(bv.position, obj); bv.err != nil {
			return
		}
	}
}

func (bv *bundleVisitor) VisitForChain(obj *bundlev1.SecretChain) {
	// Check argument
	if obj == nil {
		bv.err = fmt.Errorf("unable to process nil secret chain")
		return
	}

	// Update position
	bv.position.Secret = obj

	// Crawl secret data
	for _, p := range obj.Data {
		bv.VisitForKV(p)
	}

	// If processor given use it
	if bv.opts.cpf != nil {
		if bv.err = bv.opts.cpf(bv.position, obj); bv.err != nil {
			return
		}
	}
}

func (bv *bundleVisitor) VisitForKV(obj *bundlev1.KV) {
	// Check argument
	if obj == nil {
		bv.err = fmt.Errorf("unable to process nil secret data")
		return
	}

	// Update position
	bv.position.KV = obj

	// If processor given use it
	if bv.opts.kpf != nil {
		if bv.err = bv.opts.kpf(bv.position, obj); bv.err != nil {
			return
		}
	}
}
