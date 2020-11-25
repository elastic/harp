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

package from

import (
	"context"
	"fmt"
	"io"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle"
	"github.com/elastic/harp/pkg/bundle/template"
	"github.com/elastic/harp/pkg/bundle/template/visitor/secretbuilder"
	"github.com/elastic/harp/pkg/tasks"
	"github.com/elastic/harp/pkg/template/engine"
)

// BundleTemplateTask implements secret-container generation from BundleTemplate
// manifest.
type BundleTemplateTask struct {
	TemplateReader  tasks.ReaderProvider
	OutputWriter    tasks.WriterProvider
	TemplateContext engine.Context
}

// Run the task.
func (t *BundleTemplateTask) Run(ctx context.Context) error {
	var (
		reader io.Reader
		writer io.Writer
		spec   *bundlev1.Template
		err    error
	)

	// Create input reader
	reader, err = t.TemplateReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to open input bundle template: %w", err)
	}

	// Parse the input specification
	spec, err = template.YAML(reader)
	if err != nil {
		return fmt.Errorf("unable to parse template: %w", err)
	}

	// Initialize output
	b := &bundlev1.Bundle{
		Template: spec,
	}

	// Initialize a bundle creator
	v := secretbuilder.New(b, t.TemplateContext)

	// Execute the template to generate an output bundle
	if err = template.Execute(spec, v); err != nil {
		return fmt.Errorf("unable to generate output bundle from template: %w", err)
	}

	// Create output writer
	writer, err = t.OutputWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to open output bundle: %w", err)
	}

	// Dump all content
	if err = bundle.ToContainerWriter(writer, b); err != nil {
		return fmt.Errorf("unable to dump bundle content: %w", err)
	}

	// No error
	return nil
}
