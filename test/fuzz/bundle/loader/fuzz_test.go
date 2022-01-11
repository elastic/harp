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

//go:build go1.18
// +build go1.18

package loader_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/elastic/harp/pkg/bundle"
)

func loadFromFile(t testing.TB, filename string) []byte {
	t.Helper()

	// Load sample bundle
	completeBundle, err := os.Open(filename)
	if err != nil {
		t.Fatalf("unable to load bundle content '%v': %v", filename, err)
	}

	// Load all content
	content, err := io.ReadAll(completeBundle)
	if err != nil {
		t.Fatalf("unable to load all bundle content '%v': %v", filename, err)
	}

	return content
}

func FuzzBundleLoader(f *testing.F) {
	f.Add(loadFromFile(f, "../../../fixtures/bundles/complete.bundle"))
	f.Add(loadFromFile(f, "../../../fixtures/bundles/empty.bundle"))

	f.Fuzz(func(t *testing.T, in []byte) {
		// Read from randomized data
		bundle.FromContainerReader(bytes.NewBuffer(in))
	})
}
