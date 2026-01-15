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
	"bufio"
	"bytes"
	"os"
	"strings"

	exec "golang.org/x/sys/execabs"
)

// PathSeparatorString models the os.PathSeparator as a string.
var PathSeparatorString = string(os.PathSeparator)

// AllPackagesPath denotes all Go packages in a project.
var AllPackagesPath = "." + PathSeparatorString + "..."

// AllCommandsPath denotes all Go application packages in this project.
var AllCommandsPath = strings.Join([]string{".", "cmd", "..."}, PathSeparatorString)

// GoListSourceFilesTemplate provides a standard Go template for querying
// a project's Go source file paths.
var GoListSourceFilesTemplate = "{{$p := .}}{{range $f := .GoFiles}}{{$p.Dir}}/{{$f}}\n{{end}}"

// GoListTestFilesTemplate provides a standard Go template for querying
// a project's Go test file paths.
var GoListTestFilesTemplate = "{{$p := .}}{{range $f := .XTestGoFiles}}{{$p.Dir}}/{{$f}}\n{{end}}"

// CollectedGoFiles represents source and test Go files in a project.
// Populdated with CollectGoFiles().
var CollectedGoFiles = make(map[string]bool)

// CollectedGoSourceFiles represents the set of Go source files in a project.
// Populated with CollectGoFiles().
var CollectedGoSourceFiles = make(map[string]bool)

// CollectedGoTestFiles represents the set of Go test files in a project.
// Populdated with CollectGoFiles().
var CollectedGoTestFiles = make(map[string]bool)

// CollectGoFiles populates CollectedGoFiles, CollectedGoSourceFiles, and CollectedGoTestFiles.
//
// Vendored files are ignored.
func CollectGoFiles() error {
	var sourceOut bytes.Buffer
	var testOut bytes.Buffer

	//nolint:gosec // G204: Arguments are package constants, not user input
	cmdSource := exec.Command(
		"go",
		"list",
		"-f",
		GoListSourceFilesTemplate,
		AllPackagesPath,
	)
	cmdSource.Stdout = &sourceOut
	cmdSource.Stderr = os.Stderr

	if err := cmdSource.Run(); err != nil {
		return err
	}

	scannerSource := bufio.NewScanner(&sourceOut)

	for scannerSource.Scan() {
		pth := scannerSource.Text()

		CollectedGoFiles[pth] = true
		CollectedGoSourceFiles[pth] = true
	}

	//nolint:gosec // G204: Arguments are package constants, not user input
	cmdTest := exec.Command(
		"go",
		"list",
		"-f",
		GoListTestFilesTemplate,
		AllPackagesPath,
	)
	cmdTest.Stdout = &testOut
	cmdTest.Stderr = os.Stderr

	if err := cmdTest.Run(); err != nil {
		return err
	}

	scannerTest := bufio.NewScanner(&testOut)

	for scannerTest.Scan() {
		pth := scannerTest.Text()

		CollectedGoFiles[pth] = true
		CollectedGoTestFiles[pth] = true
	}

	return nil
}
