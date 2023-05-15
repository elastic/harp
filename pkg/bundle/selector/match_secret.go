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
	"regexp"

	"github.com/gobwas/glob"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
)

// MatchSecretStrict returns a secret key matcher specification with strict profile.
func MatchSecretStrict(value string) Specification {
	return &matchSecret{
		strict: value,
	}
}

// MatchSecretRegex returns a secret key matcher specification with regexp.
func MatchSecretRegex(regex *regexp.Regexp) Specification {
	return &matchSecret{
		regex: regex,
	}
}

// MatchSecretGlob returns a secret key matcher specification with glob query.
func MatchSecretGlob(pattern string) Specification {
	return &matchPath{
		g: glob.MustCompile(pattern),
	}
}

// matchSecret checks if secret key match the given string
type matchSecret struct {
	strict string
	regex  *regexp.Regexp
	g      glob.Glob
}

// IsSatisfiedBy returns specification satisfaction status
func (s *matchSecret) IsSatisfiedBy(object interface{}) bool {
	match := false

	// If object is a package
	if p, ok := object.(*bundlev1.Package); ok {
		// Ignore nil secret package
		if p.Secrets == nil {
			return false
		}

		for _, kv := range p.Secrets.Data {
			switch {
			case s.strict != "":
				return kv.Key == s.strict
			case s.regex != nil:
				return s.regex.MatchString(kv.Key)
			case s.g != nil:
				return s.g.Match(kv.Key)
			}
		}
	}

	return match
}
