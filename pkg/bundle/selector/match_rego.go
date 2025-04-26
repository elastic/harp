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
	"context"
	"errors"
	"fmt"

	//nolint:staticcheck // TODO: deprecated usage. Requires an update.
	"github.com/open-policy-agent/opa/rego"
	"go.uber.org/zap"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/sdk/log"
)

// MatchRego returns a Rego package matcher specification.
func MatchRego(ctx context.Context, policy string) (Specification, error) {
	// Prepare query filter
	query, err := rego.New(
		rego.Query("data.harp.matched"),
		rego.Module("harp.rego", policy),
	).PrepareForEval(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to prepare for eval: %w", err)
	}

	// Wrap as a builder
	return &regoMatcher{
		ctx:   ctx,
		query: query,
	}, nil
}

type regoMatcher struct {
	ctx   context.Context
	query rego.PreparedEvalQuery
}

// IsSatisfiedBy returns specification satisfaction status
func (s *regoMatcher) IsSatisfiedBy(object interface{}) bool {
	// If object is a package
	if p, ok := object.(*bundlev1.Package); ok {
		// Evaluate filter compliance
		matched, err := s.regoEvaluate(s.ctx, s.query, p)
		if err != nil {
			log.For(s.ctx).Debug("rego evaluation failed", zap.Error(err))
			return false
		}

		return matched
	}

	return false
}

// -----------------------------------------------------------------------------

func (s *regoMatcher) regoEvaluate(ctx context.Context, query rego.PreparedEvalQuery, input interface{}) (bool, error) {
	// Evaluate the package with the policy
	results, err := query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return false, fmt.Errorf("unable to evaluate the policy: %w", err)
	} else if len(results) == 0 {
		// Handle undefined result.
		return false, nil
	}

	// Extract decision
	keep, ok := results[0].Expressions[0].Value.(bool)
	if !ok {
		// Handle unexpected result type.
		return false, errors.New("the policy must return boolean")
	}

	// No error
	return keep, nil
}
