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
	"crypto/sha256"
	"fmt"

	"github.com/fatih/color"
	"github.com/magefile/mage/sh"
)

// UnitTest run go test
func UnitTest(packageName string) func() error {
	return func() error {
		color.Yellow("> Unit testing [%s]", packageName)
		if err := sh.Run("mkdir", "-p", "test-results/junit"); err != nil {
			return err
		}

		return sh.RunV("gotestsum", "--junitfile", fmt.Sprintf("test-results/junit/unit-%x.xml", sha256.Sum256([]byte(packageName))), "--", "-short", "-race", "-cover", packageName)
	}
}

// IntegrationTest run go test
func IntegrationTest(packageName string) func() error {
	return func() error {
		color.Yellow("> Integration testing [%s]", packageName)
		if err := sh.Run("mkdir", "-p", "test-results/junit"); err != nil {
			return err
		}

		return sh.RunV("gotestsum", "--junitfile", fmt.Sprintf("test-results/junit/integration-%x.xml", sha256.Sum256([]byte(packageName))), "--", "-tags=integration", "-short", "-race", "-cover", packageName)
	}
}
