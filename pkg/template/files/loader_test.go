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

package files

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

func TestLoadDir(t *testing.T) {
	// Get basepath
	basePath, err := filepath.Abs("../../../test")
	if err != nil {
		t.Fatalf("Failed to load testdata: %s", err)
	}

	// Initialize afero
	fs := afero.NewReadOnlyFs(afero.NewOsFs())

	l, err := Loader(fs, basePath)
	if err != nil {
		t.Fatalf("Failed to load testdata: %s", err)
	}
	c, err := l.Load()
	if err != nil {
		t.Fatalf("Failed to load testdata: %s", err)
	}
	if len(c) == 0 {
		t.Fatalf("Failed to load all test files")
	}
}
