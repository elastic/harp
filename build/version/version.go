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

package version

import (
	"encoding/json"
	"fmt"

	"github.com/dchest/uniuri"
)

// Build information. Populated at build-time.
var (
	Version          = "unknown"
	Revision         = "unknown"
	Branch           = "unknown"
	BuildUser        = "unknown"
	BuildDate        = "unknown"
	GoVersion        = "unknown"
	CompilationFlags = "unknown"
)

// Map provides the iterable version information.
var Map = map[string]string{
	"version":           Version,
	"revision":          Revision,
	"branch":            Branch,
	"build_user":        BuildUser,
	"build_date":        BuildDate,
	"go_version":        GoVersion,
	"compilation_flags": CompilationFlags,
}

// Full returns full composed version string
func Full() string {
	return fmt.Sprintf("%s [%s:%s] (Go: %s, Flags: %s, User: %s, Date: %s)", Version, Branch, Revision, GoVersion, CompilationFlags, BuildUser, BuildDate)
}

// JSON returns json representation of build info
func JSON() string {
	payload, err := json.Marshal(Map)
	if err != nil {
		panic(err)
	}

	return string(payload)
}

// ID returns an instance id
func ID() string {
	return uniuri.NewLen(64)
}
