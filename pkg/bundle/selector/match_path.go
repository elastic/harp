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
	"fmt"
	"regexp"

	"github.com/gobwas/glob"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
)

// MatchPathStrict returns a path matcher specification with strict profile.
func MatchPathStrict(value string) Specification {
	return &matchPath{
		strict: value,
	}
}

// MatchPathRegex returns a path matcher specification with regexp.
func MatchPathRegex(pattern string) (Specification, error) {
	// Compile and check filter
	m, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("unable to compile regex filter: %w", err)
	}

	// No error
	return &matchPath{
		regex: m,
	}, nil
}

// MatchPathGlob returns a path matcher specification with glob query.
func MatchPathGlob(pattern string) (Specification, error) {
	// Compile and check filter
	m, err := glob.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("unable to compile glob filter: %w", err)
	}

	// No error
	return &matchPath{
		g: m,
	}, nil
}

// MatchPath checks if secret path match the given string
type matchPath struct {
	strict string
	regex  *regexp.Regexp
	g      glob.Glob
}

// IsSatisfiedBy returns specification satisfaction status
func (s *matchPath) IsSatisfiedBy(object interface{}) bool {
	// If object is a package
	if p, ok := object.(*bundlev1.Package); ok {
		switch {
		case s.strict != "":
			return p.Name == s.strict
		case s.regex != nil:
			return s.regex.MatchString(p.Name)
		case s.g != nil:
			return s.g.Match(p.Name)
		}
	}

	return false
}
