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

package template

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/psanford/memfs"

	"github.com/elastic/harp/pkg/sdk/fsutil"
	"github.com/elastic/harp/pkg/template/engine"
)

// FileSystemTask implements filesystem template rendering task.
type FileSystemTask struct {
	InputPath          string
	OutputPath         string
	ValueFiles         []string
	SecretLoaders      []string
	Values             []string
	StringValues       []string
	FileValues         []string
	LeftDelims         string
	RightDelims        string
	AltDelims          bool
	FileLoaderRootPath string
	DryRun             bool
}

// Run the task.
//
//nolint:revive // refactor use of ctx
func (t *FileSystemTask) Run(ctx context.Context) error {
	// Prepare input filesystem
	inFS, err := fsutil.From(t.InputPath)
	if err != nil {
		return fmt.Errorf("unable to prepare input filesystem: %w", err)
	}

	// Prepare embedded files
	var (
		fileRootFS fs.FS
	)
	if t.FileLoaderRootPath != "" {
		var errRootFS error
		fileRootFS, errRootFS = fsutil.From(t.FileLoaderRootPath)
		if errRootFS != nil {
			return fmt.Errorf("unable load files filesystem: %w", errRootFS)
		}
	}

	// Prepare render context
	renderCtx, err := prepareRenderContext(&renderContextConfig{
		ValueFiles:    t.ValueFiles,
		SecretLoaders: t.SecretLoaders,
		Values:        t.Values,
		StringValues:  t.StringValues,
		FileValues:    t.FileValues,
		LeftDelims:    t.LeftDelims,
		RightDelims:   t.RightDelims,
		AltDelims:     t.AltDelims,
		FileRootPath:  fileRootFS,
	})
	if err != nil {
		return fmt.Errorf("unable to prepare rendering context: %w", err)
	}

	// Memory filesystem
	outFs := memfs.New()

	// Generate all files from input filesystem.
	if err := fs.WalkDir(inFS, ".", func(path string, d fs.DirEntry, errWalk error) error {
		if errWalk != nil {
			return errWalk
		}
		if d.IsDir() {
			return outFs.MkdirAll(path, 0o755)
		}

		// Get file content.
		body, err := fs.ReadFile(inFS, path)
		if err != nil {
			return fmt.Errorf("unable to retrieve file content %q: %w", path, err)
		}

		// Compile and execute template
		out, err := engine.RenderContext(renderCtx, string(body))
		if err != nil {
			return fmt.Errorf("unable to produce output content for file %q: %w", path, err)
		}

		// Create output file.
		return outFs.WriteFile(path, []byte(out), 0o444)
	}); err != nil {
		return fmt.Errorf("unable to render filesytem: %w", err)
	}

	// Skip copy if no output is defined.
	if t.DryRun || t.OutputPath == "" {
		return nil
	}

	// Dump filesystem
	return fsutil.Dump(outFs, t.OutputPath)
}
