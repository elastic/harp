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

package selector

import (
	"errors"
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	celext "github.com/google/cel-go/ext"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle/ruleset/engine/cel/ext"
	"github.com/elastic/harp/pkg/sdk/log"
)

// MatchCEL returns a CEL package matcher specification.
func MatchCEL(expressions []string) (Specification, error) {
	// Check arguments
	if len(expressions) == 0 {
		return nil, errors.New("CEL expressions could not be empty for matcher")
	}

	// Prepare CEL Environment
	env, err := cel.NewEnv(
		cel.Types(&bundlev1.Bundle{}, &bundlev1.Package{}, &bundlev1.SecretChain{}, &bundlev1.KV{}),
		ext.Packages(),
		ext.Secrets(),
		celext.Strings(),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to prepare CEL engine environment: %w", err)
	}

	// Assemble the complete ruleset
	ruleset := make([]cel.Program, 0, len(expressions))
	for _, exp := range expressions {
		// Parse expression
		parsed, issues := env.Parse(exp)
		if issues != nil && issues.Err() != nil {
			return nil, fmt.Errorf("unable to parse '%s', go error: %w", exp, issues.Err())
		}

		// Extract AST
		ast, cerr := env.Check(parsed)
		if cerr != nil && cerr.Err() != nil {
			return nil, fmt.Errorf("invalid CEL expression: %w", cerr.Err())
		}

		// request matching is a boolean operation, so we don't really know
		// what to do if the expression returns a non-boolean type
		//nolint:staticcheck // TODO: refactor for deprecations
		if !proto.Equal(ast.ResultType(), decls.Bool) {
			return nil, fmt.Errorf("CEL rule engine expects return type of bool, not %s", ast.ResultType())
		}

		// Compile the program
		p, err := env.Program(ast)
		if err != nil {
			return nil, fmt.Errorf("error while creating CEL program: %w", err)
		}

		// Add to context
		ruleset = append(ruleset, p)
	}

	// Wrap as a builder
	return &celMatcher{
		cel:     env,
		ruleset: ruleset,
	}, nil
}

type celMatcher struct {
	cel     *cel.Env
	ruleset []cel.Program
}

// IsSatisfiedBy returns specification satisfaction status
func (s *celMatcher) IsSatisfiedBy(object interface{}) bool {
	// If object is a package
	if p, ok := object.(*bundlev1.Package); ok {
		// Evaluate filter compliance
		matched, err := s.celEvaluate(p)
		if err != nil {
			log.Bg().Debug("cel evaluation failed", zap.Error(err))
			return false
		}

		return matched
	}

	return false
}

// -----------------------------------------------------------------------------

func (s *celMatcher) celEvaluate(input *bundlev1.Package) (bool, error) {
	// Check arguments
	if input == nil {
		return false, errors.New("unable to evaluate nil package")
	}

	// Apply evaluation (implicit AND between rules)
	for _, exp := range s.ruleset {
		// Evaluate using the bundle context
		out, _, err := exp.Eval(map[string]interface{}{
			"p": input,
		})
		if err != nil {
			return false, fmt.Errorf("an error occurred during the rule evaluation: %w", err)
		}

		// Boolean rule returned false
		if out.Value() == false {
			return false, nil
		}
	}

	// No error
	return true, nil
}
