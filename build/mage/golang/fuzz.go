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
	"path"

	"github.com/magefile/mage/sh"
)

// -----------------------------------------------------------------------------

var fuzzDir = "../test-results/fuzz"

// FuzzBuild instrument the given package name for fuzzing tests.
func FuzzBuild(name, packageName string) func() error {
	return func() error {
		// Prepare output path
		outputPath := path.Join(fuzzDir, name)

		// Check output directory existence
		if !existDir(outputPath) {
			// Create output directory
			if err := os.MkdirAll(outputPath, 0o777); err != nil {
				return fmt.Errorf("unable to create fuzz output directory: %w", err)
			}
		}

		fmt.Fprintf(os.Stdout, " > Instrumenting %s [%s]\n", name, packageName)
		return sh.Run("go-fuzz-build", "-o", fmt.Sprintf("%s.zip", outputPath), packageName)
	}
}

// FuzzRun starts a fuzzing process
func FuzzRun(name string) func() error {
	return func() error {
		// Prepare output path
		outputPath := path.Join(fuzzDir, name)

		fmt.Fprintf(os.Stdout, " > Fuzzing %s\n", name)
		return sh.Run("go-fuzz", "-bin", fmt.Sprintf("%s.zip", outputPath), "-workdir", outputPath)
	}
}

func existDir(fpath string) bool {
	st, err := os.Stat(fpath)
	if err != nil {
		return false
	}
	return st.IsDir()
}
