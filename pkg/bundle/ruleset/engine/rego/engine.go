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

package rego

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/open-policy-agent/opa/rego"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle/ruleset/engine"
)

const (
	maxPolicySize = 5 * 1024 * 1025 // 5MB
)

func New(ctx context.Context, r io.Reader) (engine.PackageLinter, error) {
	// Read all policy content
	policy, err := io.ReadAll(io.LimitReader(r, maxPolicySize))
	if err != nil {
		return nil, fmt.Errorf("unable to read the policy content: %w", err)
	}

	// Parse and prepare the policy
	query, err := rego.New(
		rego.Query("data.harp.compliant"),
		rego.Module("harp.rego", string(policy)),
	).PrepareForEval(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to prepare for eval: %w", err)
	}

	// Return engine
	return &ruleEngine{
		query: query,
	}, nil
}

// -----------------------------------------------------------------------------

type ruleEngine struct {
	query rego.PreparedEvalQuery
}

func (re *ruleEngine) EvaluatePackage(ctx context.Context, p *bundlev1.Package) error {
	// Check arguments
	if p == nil {
		return errors.New("unable to evaluate nil package")
	}

	// Evaluation with the given package
	results, err := re.query.Eval(ctx, rego.EvalInput(p))
	if err != nil {
		return fmt.Errorf("unable to evaluate the policy: %w", err)
	} else if len(results) == 0 {
		// Handle undefined result.
		return nil
	}

	for _, result := range results {
		for _, expression := range result.Expressions {
			// Extract result
			compliant, ok := expression.Value.(bool)
			if !ok {
				// Handle unexpected result type.
				return errors.New("the policy must return boolean")
			}

			// Check package compliance
			if !compliant {
				return engine.ErrRuleNotValid
			}
		}
	}

	// Package validated
	return nil
}
