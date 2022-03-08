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
	"github.com/fatih/color"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Import fix all source code imports.
func Import() error {
	mg.Deps(CollectGoFiles)

	color.Cyan("## Process imports")

	for pth := range CollectedGoFiles {
		args := []string{"write", "-s", "Standard", "-s", "Prefix(golang.org/x/)", "-s", "Default", "-s", "Prefix(github.com/elastic)"}
		args = append(args, pth)

		if err := sh.RunV("gci", args...); err != nil {
			return err
		}
	}

	return nil
}
