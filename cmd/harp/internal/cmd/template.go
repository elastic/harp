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

package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/bundle"
	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
	tplcmdutil "github.com/elastic/harp/pkg/template/cmdutil"
	"github.com/elastic/harp/pkg/template/engine"
	"github.com/elastic/harp/pkg/vault/kv"
)

type templateParams struct {
	InputPath     string
	OutputPath    string
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

// -----------------------------------------------------------------------------

var templateCmd = func() *cobra.Command {
	params := &templateParams{}

	cmd := &cobra.Command{
		Use:     "template",
		Aliases: []string{"t", "tpl"},
		Short:   "Read a template and execute it",
		Run: func(cmd *cobra.Command, args []string) {
			runTemplate(cmd.Context(), params)
		},
	}

	// Parameters
	cmd.Flags().StringVar(&params.InputPath, "in", "-", "Template input path ('-' for stdin or filename)")
	cmd.Flags().StringVar(&params.OutputPath, "out", "", "Output file ('-' for stdout or a filename)")
	cmd.Flags().StringVar(&params.RootPath, "root", "", "Defines file loader root base path")
	cmd.Flags().StringArrayVarP(&params.SecretLoaders, "secrets-from", "s", []string{"vault"}, "Specifies secret containers to load ('vault' for Vault loader or '-' for stdin or filename)")
	cmd.Flags().StringArrayVarP(&params.ValueFiles, "values", "f", []string{}, "Specifies value files to load")
	cmd.Flags().StringArrayVar(&params.Values, "set", []string{}, "Specifies value (k=v)")
	cmd.Flags().StringArrayVar(&params.StringValues, "set-string", []string{}, "Specifies value (k=string)")
	cmd.Flags().StringArrayVar(&params.FileValues, "set-file", []string{}, "Specifies value (k=filepath)")
	cmd.Flags().StringVar(&params.LeftDelims, "left-delimiter", "{{", "Template left delimiter (default to '{{')")
	cmd.Flags().StringVar(&params.RightDelims, "right-delimiter", "}}", "Template right delimiter (default to '}}')")
	cmd.Flags().BoolVar(&params.AltDelims, "alt-delims", false, "Define '[[' and ']]' as template delimiters.")

	return cmd
}

//nolint:funlen // to split
func runTemplate(ctx context.Context, params *templateParams) {
	ctx, cancel := cmdutil.Context(ctx, "harp-template", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
	defer cancel()

	var (
		reader io.Reader
		err    error
	)

	// Create input reader
	reader, err = cmdutil.Reader(params.InputPath)
	if err != nil {
		log.For(ctx).Fatal("unable to open input template", zap.Error(err), zap.String("path", params.InputPath))
	}

	// Load values
	valueOpts := tplcmdutil.ValueOptions{
		ValueFiles:   params.ValueFiles,
		Values:       params.Values,
		StringValues: params.StringValues,
		FileValues:   params.FileValues,
	}
	values, err := valueOpts.MergeValues()
	if err != nil {
		log.For(ctx).Fatal("unable to process values", zap.Error(err))
	}

	// Load files
	var files engine.Files
	if params.RootPath != "" {
		absRootPath, errAbs := filepath.Abs(params.RootPath)
		if errAbs != nil {
			log.For(ctx).Fatal("unable to get absolute template root path", zap.Error(errAbs))
		}

		files, errAbs = tplcmdutil.Files(os.DirFS(absRootPath), ".")
		if errAbs != nil {
			log.For(ctx).Fatal("unable to process files", zap.Error(errAbs))
		}
	}

	// Drain reader
	body, err := io.ReadAll(reader)
	if err != nil {
		log.For(ctx).Fatal("unable to drain input template reader", zap.Error(err), zap.String("path", params.InputPath))
	}

	// If alternative delimiters is used
	if params.AltDelims {
		params.LeftDelims = "[["
		params.RightDelims = "]]"
	}

	// Process secret readers
	secretReaders := []engine.SecretReaderFunc{}
	for _, sr := range params.SecretLoaders {
		if sr == "vault" {
			// Initialize Vault connection
			vaultClient, errVault := api.NewClient(api.DefaultConfig())
			if errVault != nil {
				log.For(ctx).Fatal("unable to initialize vault secret loader", zap.Error(errVault), zap.String("container-path", sr))
			}

			secretReaders = append(secretReaders, kv.SecretGetter(vaultClient))
			continue
		}

		// Read container
		containerReader, errLoader := cmdutil.Reader(sr)
		if errLoader != nil {
			log.For(ctx).Fatal("unable to read secret container", zap.Error(errLoader), zap.String("container-path", sr))
		}

		// Load container
		b, errBundle := bundle.FromContainerReader(containerReader)
		if errBundle != nil {
			log.For(ctx).Fatal("unable to decode secret container", zap.Error(errBundle), zap.String("container-path", sr))
		}

		// Append secret loader
		secretReaders = append(secretReaders, bundle.SecretReader(b))
	}

	// Compile and execute template
	out, err := engine.RenderContext(engine.NewContext(
		engine.WithName(params.InputPath),
		engine.WithDelims(params.LeftDelims, params.RightDelims),
		engine.WithValues(values),
		engine.WithFiles(files),
		engine.WithSecretReaders(secretReaders...),
	), string(body))
	if err != nil {
		log.For(ctx).Fatal("unable to produce output content", zap.Error(err), zap.String("path", params.InputPath))
	}

	// Create output writer
	writer, err := cmdutil.Writer(params.OutputPath)
	if err != nil {
		log.For(ctx).Fatal("unable to create output writer", zap.Error(err), zap.String("path", params.OutputPath))
	}

	// Write rendered content
	fmt.Fprintf(writer, "%s", out)
}
