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

package ruleset

import (
	"context"
	"fmt"

	"sigs.k8s.io/yaml"

	"github.com/elastic/harp/pkg/bundle"
	"github.com/elastic/harp/pkg/bundle/ruleset"
	"github.com/elastic/harp/pkg/tasks"
)

// FromBundleTask implements RuleSet generation from a bundle.
type FromBundleTask struct {
	ContainerReader tasks.ReaderProvider
	OutputWriter    tasks.WriterProvider
}

// Run the task.
func (t *FromBundleTask) Run(ctx context.Context) error {
	// Create input reader
	reader, err := t.ContainerReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to initialize bundle reader: %w", err)
	}

	// Load bundle
	b, err := bundle.FromContainerReader(reader)
	if err != nil {
		return fmt.Errorf("unable to load bundle content: %w", err)
	}

	// Generate ruleset
	rs, err := ruleset.FromBundle(b)
	if err != nil {
		return fmt.Errorf("unable to generate RuleSet from given bundle: %w", err)
	}

	// Marshal as YAML
	ruleSetSpec, err := yaml.Marshal(rs)
	if err != nil {
		return fmt.Errorf("unable to generate YAML descriptor: %w", err)
	}

	// Create output writer
	writer, err := t.OutputWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to initialize output writer: %w", err)
	}

	// Write output
	fmt.Fprintln(writer, string(ruleSetSpec))

	// No error
	return nil
}
