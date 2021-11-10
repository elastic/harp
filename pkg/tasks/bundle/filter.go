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
	"regexp"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle"
	"github.com/elastic/harp/pkg/bundle/selector"
	"github.com/elastic/harp/pkg/sdk/types"
	"github.com/elastic/harp/pkg/tasks"
	"github.com/jmespath/go-jmespath"
)

// FilterTask implements secret container filtering task.
type FilterTask struct {
	ContainerReader tasks.ReaderProvider
	OutputWriter    tasks.WriterProvider
	ReverseLogic    bool
	KeepPaths       []string
	ExcludePaths    []string
	JMESPath        string
}

// Run the task.
func (t *FilterTask) Run(ctx context.Context) error {
	// Check arguments
	if types.IsNil(t.ContainerReader) {
		return errors.New("unable to run task with a nil containerReader provider")
	}
	if types.IsNil(t.OutputWriter) {
		return errors.New("unable to run task with a nil outputWriter provider")
	}

	// Create input reader
	reader, err := t.ContainerReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to open input bundle: %w", err)
	}

	// Load bundle
	b, err := bundle.FromContainerReader(reader)
	if err != nil {
		return fmt.Errorf("unable to load bundle content: %w", err)
	}

	var errFilter error

	if len(t.KeepPaths) > 0 {
		b.Packages, errFilter = t.keepFilter(b.Packages, t.KeepPaths, t.ReverseLogic)
		if errFilter != nil {
			return fmt.Errorf("unable to filter bundle packages: %w", errFilter)
		}
	}

	if len(t.ExcludePaths) > 0 {
		b.Packages, errFilter = t.excludeFilter(b.Packages, t.ExcludePaths, t.ReverseLogic)
		if errFilter != nil {
			return fmt.Errorf("unable to filter bundle packages: %w", errFilter)
		}
	}

	if t.JMESPath != "" {
		b.Packages, errFilter = t.jmespathFilter(b.Packages, t.JMESPath, t.ReverseLogic)
		if errFilter != nil {
			return fmt.Errorf("unable to filter bundle packages: %w", errFilter)
		}
	}

	// Create output writer
	writer, err := t.OutputWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to open output bundle: %w", err)
	}

	// Dump all content
	if err := bundle.ToContainerWriter(writer, b); err != nil {
		return fmt.Errorf("unable to dump bundle content: %w", err)
	}

	// No error
	return nil
}

// -----------------------------------------------------------------------------

func (t *FilterTask) keepFilter(in []*bundlev1.Package, paths []string, reverseLogic bool) ([]*bundlev1.Package, error) {
	// Check Arguments
	if len(in) == 0 {
		return in, nil
	}
	if len(paths) == 0 {
		return in, nil
	}

	pkgs := []*bundlev1.Package{}

	for _, includePath := range paths {
		includePathRegexp, errInclude := regexp.Compile(includePath)
		if errInclude != nil {
			return nil, fmt.Errorf("unable to compile keep regexp '%s': %w", includePath, errInclude)
		}

		for _, p := range in {
			matched := includePathRegexp.MatchString(p.Name)
			if matched && !reverseLogic || !matched && reverseLogic {
				pkgs = append(pkgs, p)
			}
		}
	}

	// No error
	return pkgs, nil
}

func (t *FilterTask) excludeFilter(in []*bundlev1.Package, paths []string, reverseLogic bool) ([]*bundlev1.Package, error) {
	// Check Arguments
	if len(in) == 0 {
		return in, nil
	}
	if len(paths) == 0 {
		return in, nil
	}

	pkgs := []*bundlev1.Package{}

	for _, excludePath := range paths {
		excludePathRegexp, errExclude := regexp.Compile(excludePath)
		if errExclude != nil {
			return nil, fmt.Errorf("unable to compile exclusion regexp '%s': %w", excludePath, errExclude)
		}

		for _, p := range in {
			matched := !excludePathRegexp.MatchString(p.Name)
			if matched && !reverseLogic || !matched && reverseLogic {
				pkgs = append(pkgs, p)
			}
		}
	}

	// No error
	return pkgs, nil
}

func (t *FilterTask) jmespathFilter(in []*bundlev1.Package, filter string, reverseLogic bool) ([]*bundlev1.Package, error) {
	// Check Arguments
	if len(in) == 0 {
		return in, nil
	}
	if filter == "" {
		return in, nil
	}

	pkgs := []*bundlev1.Package{}

	// Compile expression first
	exp, errJMESPath := jmespath.Compile(filter)
	if errJMESPath != nil {
		return nil, fmt.Errorf("unable to compile JMESPath filter '%s': %w", filter, errJMESPath)
	}

	// Initialize selector
	s := selector.MatchJMESPath(exp)

	// Apply package filtering
	for _, p := range in {
		matched := s.IsSatisfiedBy(p)
		if matched && !reverseLogic || !matched && reverseLogic {
			pkgs = append(pkgs, p)
		}
	}

	// No error
	return pkgs, nil
}
