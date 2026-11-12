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
	"io"
	"io/fs"

	"github.com/hashicorp/vault/api"

	"github.com/elastic/harp/pkg/bundle"
	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/fsutil"
	"github.com/elastic/harp/pkg/sdk/types"
	"github.com/elastic/harp/pkg/tasks"
	tplcmdutil "github.com/elastic/harp/pkg/template/cmdutil"
	"github.com/elastic/harp/pkg/template/engine"
	"github.com/elastic/harp/pkg/vault/kv"
)

// RenderTask implements single template rendering task.
type RenderTask struct {
	InputReader   tasks.ReaderProvider
	OutputWriter  tasks.WriterProvider
	ValueFiles    []string
	SecretLoaders []string
	Values        []string
	StringValues  []string
	FileValues    []string
	LeftDelims    string
	RightDelims   string
	AltDelims     bool
	RootPath      string
}

// Run the task.
func (t *RenderTask) Run(ctx context.Context) error {
	var (
		reader io.Reader
		err    error
	)

	// Create input reader
	reader, err = t.InputReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to create input reader: %w", err)
	}

	// Drain reader
	body, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("unable to drain input template reader: %w", err)
	}

	var fileRootFS fs.FS
	if t.RootPath != "" {
		var errRootFS error
		fileRootFS, errRootFS = fsutil.From(t.RootPath)
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

	// Compile and execute template
	out, err := engine.RenderContext(renderCtx, string(body))
	if err != nil {
		return fmt.Errorf("unable to produce output content: %w", err)
	}

	// Create output writer
	writer, err := t.OutputWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to create output writer: %w", err)
	}

	// Write rendered content
	_, _ = fmt.Fprintf(writer, "%s", out)

	// No error
	return nil
}

// -----------------------------------------------------------------------------

type renderContextConfig struct {
	ValueFiles    []string
	SecretLoaders []string
	Values        []string
	StringValues  []string
	FileValues    []string
	LeftDelims    string
	RightDelims   string
	AltDelims     bool
	FileRootPath  fs.FS
}

func prepareRenderContext(cfg *renderContextConfig) (engine.Context, error) {
	// Load values
	valueOpts := tplcmdutil.ValueOptions{
		ValueFiles:   cfg.ValueFiles,
		Values:       cfg.Values,
		StringValues: cfg.StringValues,
		FileValues:   cfg.FileValues,
	}
	values, err := valueOpts.MergeValues()
	if err != nil {
		return nil, fmt.Errorf("unable to process input values: %w", err)
	}

	// Load files
	var files engine.Files
	if !types.IsNil(cfg.FileRootPath) {
		var errFs error
		files, errFs = tplcmdutil.Files(cfg.FileRootPath, ".")
		if errFs != nil {
			return nil, fmt.Errorf("unable to process files: %w", errFs)
		}
	}

	// If alternative delimiters is used
	if cfg.AltDelims {
		cfg.LeftDelims = "[["
		cfg.RightDelims = "]]"
	}

	// Process secret readers
	secretReaders := []engine.SecretReaderFunc{}
	for _, sr := range cfg.SecretLoaders {
		if sr == "vault" {
			// Initialize Vault connection
			vaultClient, errVault := api.NewClient(api.DefaultConfig())
			if errVault != nil {
				return nil, fmt.Errorf("unable to initialize vault secret loader: %w", errVault)
			}

			secretReaders = append(secretReaders, kv.SecretGetter(vaultClient))
			continue
		}

		// Read container
		containerReader, errLoader := cmdutil.Reader(sr)
		if errLoader != nil {
			return nil, fmt.Errorf("unable to read secret container: %w", errLoader)
		}

		// Load container
		b, errBundle := bundle.FromContainerReader(containerReader)
		if errBundle != nil {
			return nil, fmt.Errorf("unable to decode secret container: %w", err)
		}

		// Append secret loader
		secretReaders = append(secretReaders, bundle.SecretReader(b))
	}

	// Create rendering context
	renderCtx := engine.NewContext(
		engine.WithName("template"),
		engine.WithDelims(cfg.LeftDelims, cfg.RightDelims),
		engine.WithValues(values),
		engine.WithFiles(files),
		engine.WithSecretReaders(secretReaders...),
	)

	// No error
	return renderCtx, nil
}
