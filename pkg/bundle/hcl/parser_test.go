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

package hcl

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/sebdah/goldie"
	"github.com/stretchr/testify/require"
)

func init() {
	goldie.FixtureDir = "testdata"
	spew.Config.DisablePointerAddresses = true
}

func TestParseFile(t *testing.T) {
	f, err := os.Open("testdata")
	require.NoError(t, err)
	defer f.Close()

	fis, err := f.Readdir(-1)
	require.NoError(t, err)
	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}

		if filepath.Ext(fi.Name()) == ".golden" {
			continue
		}

		t.Run(fi.Name(), func(t *testing.T) {
			_, err := ParseFile(filepath.Join("testdata", fi.Name()))
			require.NoError(t, err)

			//goldie.Assert(t, fi.Name(), []byte(spew.Sdump(cfg)))
		})
	}
}
