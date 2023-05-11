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

package lint

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/xeipuuv/gojsonschema"
	"sigs.k8s.io/yaml"

	"github.com/elastic/harp/pkg/bundle"
	"github.com/elastic/harp/pkg/bundle/patch"
	"github.com/elastic/harp/pkg/bundle/ruleset"
	"github.com/elastic/harp/pkg/bundle/template"
	"github.com/elastic/harp/pkg/tasks"
)

// ValidateTask implements input linter task.
type ValidateTask struct {
	SourceReader tasks.ReaderProvider
	OutputWriter tasks.WriterProvider
	Schema       string
	SchemaOnly   bool
}

type fileSpec struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
}

var schemaRegistry = map[string]struct {
	Definition []byte
	LintFunc   func(io.Reader) ([]gojsonschema.ResultError, error)
}{
	"Bundle":         {Definition: bundle.JSONSchema(), LintFunc: bundle.Lint},
	"BundlePatch":    {Definition: patch.JSONSchema(), LintFunc: patch.Lint},
	"RuleSet":        {Definition: ruleset.JSONSchema(), LintFunc: ruleset.Lint},
	"BundleTemplate": {Definition: template.JSONSchema(), LintFunc: template.Lint},
}

// Run the task.
func (t *ValidateTask) Run(ctx context.Context) error {
	var (
		reader io.Reader
		err    error
	)

	// Create input reader
	reader, err = t.SourceReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to read input reader: %w", err)
	}

	// Drain input reader
	payload, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("unable to drain input reader: %w", err)
	}

	// Detect the appropriate schema
	if t.Schema == "" {
		// Decode as YAML any object
		var specBody fileSpec
		if errYaml := yaml.Unmarshal(payload, &specBody); errYaml != nil {
			return fmt.Errorf("unable to decode spec as YAML: %w", err)
		}

		// Check API Version
		if specBody.APIVersion != "harp.elastic.co/v1" {
			return fmt.Errorf("unsupported YAML file format '%s'", specBody.APIVersion)
		}

		// Assign detected kind
		t.Schema = specBody.Kind
	}

	// Select lint strategy
	s, ok := schemaRegistry[t.Schema]
	if !ok {
		return fmt.Errorf("unsupported schema definition for '%s'", t.Schema)
	}

	// Create output writer
	writer, err := t.OutputWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to open output bundle: %w", err)
	}

	// Display jsonschema
	if t.SchemaOnly {
		fmt.Fprintln(writer, string(s.Definition))
		return nil
	}

	// Execute the lint evaluation
	validationErrors, err := s.LintFunc(bytes.NewReader(payload))
	switch {
	case len(validationErrors) > 0:
		for _, e := range validationErrors {
			fmt.Fprintf(writer, " - %s\n", e.String())
		}
		return err
	case err != nil:
		return fmt.Errorf("unexpected validation error occurred: %w", err)
	default:
	}

	// No error
	return nil
}
