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

package patch

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func mustLoad(filePath string) io.Reader {
	f, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	return f
}

type readerTestCase struct {
	name    string
	args    io.Reader
	wantErr bool
}

func generateReaderTests(t *testing.T, rootPath, state string, wantErr bool) []readerTestCase {
	tests := []readerTestCase{}
	// Generate invalid test cases
	if err := filepath.Walk(filepath.Join(rootPath, state), func(path string, info os.FileInfo, errWalk error) error {
		if errWalk != nil {
			return errWalk
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != "yaml" {
			return nil
		}

		tests = append(tests, readerTestCase{
			name:    fmt.Sprintf("%s-%s", state, filepath.Base(info.Name())),
			args:    mustLoad(path),
			wantErr: wantErr,
		})
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	return tests
}

func TestYAML(t *testing.T) {
	tests := []readerTestCase{
		{
			name:    "nil",
			wantErr: true,
		},
	}

	// Generate invalid test cases
	tests = append(tests, generateReaderTests(t, "../../../test/fixtures/patch", "invalid", true)...)

	// Generate valid test cases
	tests = append(tests, generateReaderTests(t, "../../../test/fixtures/patch", "valid", false)...)

	// Execute them
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := YAML(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("YAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
