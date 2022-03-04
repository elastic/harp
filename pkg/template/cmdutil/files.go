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

package cmdutil

import (
	"fmt"
	"io/fs"

	"github.com/elastic/harp/pkg/sdk/types"
	"github.com/elastic/harp/pkg/template/engine"
	"github.com/elastic/harp/pkg/template/files"
)

// Files returns template files.
func Files(fileSystem fs.FS, basePath string) (engine.Files, error) {
	// Check arguments
	if types.IsNil(fileSystem) {
		return nil, fmt.Errorf("unable to load files without a default filesystem to use")
	}

	// Get appropriate loader
	loader, err := files.Loader(fileSystem, basePath)
	if err != nil {
		return nil, fmt.Errorf("unable to process files: %w", err)
	}

	// Crawl and load file content
	fileList, err := loader.Load()
	if err != nil {
		return nil, fmt.Errorf("unable to load files: %w", err)
	}

	// Wrap as template files
	templateFiles := engine.NewFiles(fileList)

	// No error
	return templateFiles, nil
}
