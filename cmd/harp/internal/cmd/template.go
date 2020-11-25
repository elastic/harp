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
	"fmt"
	"io"
	"io/ioutil"

	"github.com/hashicorp/vault/api"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/bundle"
	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
	tplcmdutil "github.com/elastic/harp/pkg/template/cmdutil"
	"github.com/elastic/harp/pkg/template/engine"
	"github.com/elastic/harp/pkg/vault/kv"
)

var (
	templateInputPath     string
	templateOutputPath    string
	templateValueFiles    []string
	templateSecretLoaders []string
	templateValues        []string
	templateStringValues  []string
	templateFileValues    []string
	templateLeftDelims    string
	templateRightDelims   string
	templateAltDelims     bool
	templateRootPath      string
)

// -----------------------------------------------------------------------------

var templateCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "template",
		Aliases: []string{"t", "tpl"},
		Short:   "Read a template and execute it",
		Run:     runTemplate,
	}

	// Parameters
	cmd.Flags().StringVar(&templateInputPath, "in", "-", "Template input path ('-' for stdin or filename)")
	cmd.Flags().StringVar(&templateOutputPath, "out", "", "Output file ('-' for stdout or a filename)")
	cmd.Flags().StringVar(&templateRootPath, "root", "", "Defines file loader root base path")
	cmd.Flags().StringArrayVarP(&templateSecretLoaders, "secrets-from", "s", []string{"vault"}, "Specifies secret containers to load ('vault' for Vault loader or '-' for stdin or filename)")
	cmd.Flags().StringArrayVarP(&templateValueFiles, "values", "f", []string{}, "Specifies value files to load")
	cmd.Flags().StringArrayVar(&templateValues, "set", []string{}, "Specifies value (k=v)")
	cmd.Flags().StringArrayVar(&templateStringValues, "set-string", []string{}, "Specifies value (k=string)")
	cmd.Flags().StringArrayVar(&templateFileValues, "set-file", []string{}, "Specifies value (k=filepath)")
	cmd.Flags().StringVar(&templateLeftDelims, "left-delimiter", "{{", "Template left delimiter (default to '{{')")
	cmd.Flags().StringVar(&templateRightDelims, "right-delimiter", "}}", "Template right delimiter (default to '}}')")
	cmd.Flags().BoolVar(&templateAltDelims, "alt-delims", false, "Define '[[' and ']]' as template delimiters.")

	return cmd
}

func runTemplate(cmd *cobra.Command, args []string) {
	ctx, cancel := cmdutil.Context(cmd.Context(), "harp-template", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
	defer cancel()

	var (
		reader io.Reader
		err    error
	)

	// Create input reader
	reader, err = cmdutil.Reader(templateInputPath)
	if err != nil {
		log.For(ctx).Fatal("unable to open input template", zap.Error(err), zap.String("path", templateInputPath))
	}

	// Load values
	valueOpts := tplcmdutil.ValueOptions{
		ValueFiles:   templateValueFiles,
		Values:       templateValues,
		StringValues: templateStringValues,
		FileValues:   templateFileValues,
	}
	values, err := valueOpts.MergeValues()
	if err != nil {
		log.For(ctx).Fatal("unable to process values", zap.Error(err))
	}

	// Load files
	var files engine.Files
	if templateRootPath != "" {
		var errLoader error
		files, errLoader = tplcmdutil.Files(afero.NewOsFs(), templateRootPath)
		if errLoader != nil {
			log.For(ctx).Fatal("unable to process files", zap.Error(errLoader))
		}
	}

	// Drain reader
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		log.For(ctx).Fatal("unable to drain input template reader", zap.Error(err), zap.String("path", templateInputPath))
	}

	// If alternative delimiters is used
	if templateAltDelims {
		templateLeftDelims = "[["
		templateRightDelims = "]]"
	}

	// Process secret readers
	secretReaders := []engine.SecretReaderFunc{}
	for _, sr := range templateSecretLoaders {
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
		b, errBundle := bundle.Load(containerReader)
		if errBundle != nil {
			log.For(ctx).Fatal("unable to decode secret container", zap.Error(errBundle), zap.String("container-path", sr))
		}

		// Append secret loader
		secretReaders = append(secretReaders, bundle.SecretReader(b))
	}

	// Compile and execute template
	out, err := engine.RenderContext(engine.NewContext(
		engine.WithName(templateInputPath),
		engine.WithDelims(templateLeftDelims, templateRightDelims),
		engine.WithValues(values),
		engine.WithFiles(files),
		engine.WithSecretReaders(secretReaders...),
	), string(body))
	if err != nil {
		log.For(ctx).Fatal("unable to produce output content", zap.Error(err), zap.String("path", templateInputPath))
	}

	// Create output writer
	writer, err := cmdutil.Writer(templateOutputPath)
	if err != nil {
		log.For(ctx).Fatal("unable to create output writer", zap.Error(err), zap.String("path", templateOutputPath))
	}

	// Write rendered content
	fmt.Fprintf(writer, "%s", out)
}
