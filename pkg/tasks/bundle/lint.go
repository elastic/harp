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

package bundle

import (
	"context"
	"errors"
	"fmt"

	"github.com/elastic/harp/pkg/bundle"
	"github.com/elastic/harp/pkg/bundle/ruleset"
	"github.com/elastic/harp/pkg/sdk/types"
	"github.com/elastic/harp/pkg/tasks"
)

// LintTask implements bundle linting task.
type LintTask struct {
	ContainerReader tasks.ReaderProvider
	RuleSetReader   tasks.ReaderProvider
}

// Run the task.
func (t *LintTask) Run(ctx context.Context) error {
	// Check arguments
	if types.IsNil(t.ContainerReader) {
		return errors.New("unable to run task with a nil containerReader provider")
	}
	if types.IsNil(t.RuleSetReader) {
		return errors.New("unable to run task with a nil ruleSetReader provider")
	}

	// Create input reader
	reader, err := t.ContainerReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to initialize bundle reader: %w", err)
	}

	// Create input reader
	rsReader, err := t.RuleSetReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to initialize ruleset reader: %w", err)
	}

	// Parse the input specification
	spec, err := ruleset.YAML(rsReader)
	if err != nil {
		return fmt.Errorf("unable to parse ruleset file: %w", err)
	}

	// Load bundle
	b, err := bundle.FromContainerReader(reader)
	if err != nil {
		return fmt.Errorf("unable to load bundle content: %w", err)
	}

	if err := ruleset.Evaluate(ctx, b, spec); err != nil {
		return fmt.Errorf("unable to validate given bundle: %w", err)
	}

	// No error
	return nil
}
