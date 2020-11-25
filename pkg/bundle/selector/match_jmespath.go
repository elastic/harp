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
	"encoding/json"

	"github.com/jmespath/go-jmespath"
	"google.golang.org/protobuf/encoding/protojson"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
)

// MatchJMESPath returns a JMESPatch package matcher specification.
func MatchJMESPath(exp *jmespath.JMESPath) Specification {
	return &jmesPathMatcher{
		exp: exp,
	}
}

type jmesPathMatcher struct {
	exp *jmespath.JMESPath
}

// IsSatisfiedBy returns specification satisfaction status
func (s *jmesPathMatcher) IsSatisfiedBy(object interface{}) bool {
	// If object is a package
	if p, ok := object.(*bundlev1.Package); ok {
		// Eliminate all package in case of nil query.
		if s.exp == nil {
			return false
		}

		// Rencode as json
		jsonRaw, err := protojson.Marshal(p)
		if err != nil {
			return false
		}

		var object map[string]interface{}
		if errJSON := json.Unmarshal(jsonRaw, &object); errJSON != nil {
			return false
		}

		// Check if query match results
		res, err := s.exp.Search(object)
		if err != nil {
			return false
		}

		// If result is a boolean
		if bRes, ok := res.(bool); ok {
			return bRes
		}
	}

	return false
}
