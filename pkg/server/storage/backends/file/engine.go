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

package file

import (
	"context"
	"fmt"
	"net/url"
	"regexp"

	"github.com/spf13/afero"

	"github.com/elastic/harp/pkg/server/storage"
)

type engine struct {
	u        *url.URL
	fs       afero.Fs
	basePath string
}

func build(u *url.URL) (storage.Engine, error) {
	// Prepare virtual filesystem
	var fs afero.Fs
	fs = afero.NewOsFs()
	fs = afero.NewReadOnlyFs(fs)
	fs = afero.NewBasePathFs(fs, u.Path)
	fs = afero.NewRegexpFs(fs, regexp.MustCompile(`\.(properties|conf|toml|xml|json|ya?ml|txt)$`))

	// Build engine instance
	return &engine{
		u:        u,
		basePath: u.Path,
		fs:       fs,
	}, nil
}

func init() {
	// Register to storage factory
	storage.MustRegister("file", build)
}

// -----------------------------------------------------------------------------

func (e *engine) Get(_ context.Context, id string) ([]byte, error) {
	// Open and read all file content
	out, err := afero.ReadFile(e.fs, id)
	if err != nil {
		return nil, fmt.Errorf("file: unable to read file content: %w", err)
	}

	// No error
	return out, nil
}
