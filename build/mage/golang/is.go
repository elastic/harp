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

package golang

import (
	"regexp"
	"runtime"

	semver "github.com/Masterminds/semver/v3"
	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/sdk/log"
)

var versionSemverRe = regexp.MustCompile("[0-9.]+")

// Is return true if current go version is included in given array.
func Is(constraints ...string) bool {
	// Extract version digit from go runtime version.
	v := versionSemverRe.FindString(runtime.Version())
	if v == "" {
		panic("unable to extract go runtime version")
	}

	// Parse golang version as semver
	sv := semver.MustParse(v)

	// Parse all constraints and check according to go version.
	for _, c := range constraints {
		constraint, err := semver.NewConstraint(c)
		if err != nil {
			log.Bg().Error("unable to parse version constraint", zap.String("constraint", c))
			return false
		}

		// Check version
		if constraint.Check(sv) {
			return true
		}
	}

	// No match found
	return false
}
