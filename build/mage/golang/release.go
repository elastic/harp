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
	"fmt"
	"os"

	"github.com/magefile/mage/mg"

	"github.com/elastic/harp/build/mage/git"
)

// -----------------------------------------------------------------------------

// Release build and generate a final release artifact.
func Release(name, packageName, version string, opts ...BuildOption) func() error {
	return func() error {
		mg.SerialDeps(git.CollectInfo)

		// Retrieve release from ENV
		releaseVersion := os.Getenv("RELEASE")
		if releaseVersion == "" {
			return fmt.Errorf("RELEASE environment variable is missing")
		}

		// Release must be done on main branch only
		if git.Branch != "main" && os.Getenv("RELEASE_FORCE") == "" {
			return fmt.Errorf("a release must be build on 'main' branch only")
		}

		// Build the artifact
		if err := Build(
			name,
			packageName,
			version,
			opts...,
		)(); err != nil {
			return err
		}

		// No error
		return nil
	}
}
