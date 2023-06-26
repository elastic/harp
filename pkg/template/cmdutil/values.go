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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/flags/strvals"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/template/values"
)

// Inspired from Helm v3
// https://github.com/helm/helm/blob/master/pkg/cli/values/options.go

// ValueOptions represents value loader options.
type ValueOptions struct {
	ValueFiles   []string
	StringValues []string
	Values       []string
	FileValues   []string
}

// MergeValues merges values from files specified via -f/--values and directly
// via --set, --set-string, or --set-file, marshaling them to YAML
func (opts *ValueOptions) MergeValues() (map[string]interface{}, error) {
	base := map[string]interface{}{}

	// save the current directory and chdir back to it when done
	currentDirectory, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("unable to save current directory path: %w", err)
	}

	// User specified a values files via --values
	for _, filePath := range opts.ValueFiles {
		currentMap := map[string]interface{}{}

		// Process each file path
		if err := processFilePath(currentDirectory, filePath, &currentMap); err != nil {
			return nil, err
		}

		// Merge with the previous map
		base = mergeMaps(base, currentMap)
	}

	// User specified a value via --set
	for _, value := range opts.Values {
		if err := strvals.ParseInto(value, base); err != nil {
			return nil, fmt.Errorf("failed parsing --set data: %w", err)
		}
	}

	// User specified a value via --set-string
	for _, value := range opts.StringValues {
		if err := strvals.ParseIntoString(value, base); err != nil {
			return nil, fmt.Errorf("failed parsing --set-string data: %w", err)
		}
	}

	// User specified a value via --set-file
	for _, value := range opts.FileValues {
		reader := func(rs []rune) (interface{}, error) {
			b, err := os.ReadFile(string(rs))
			return string(b), err
		}
		if err := strvals.ParseIntoFile(value, base, reader); err != nil {
			return nil, fmt.Errorf("failed parsing --set-file data: %w", err)
		}
	}

	return base, nil
}

// -----------------------------------------------------------------------------

func processFilePath(currentDirectory, filePath string, result interface{}) error {
	defer func() {
		log.CheckErr("unable to reset to current working directory", os.Chdir(currentDirectory))
	}()

	// Check for type overrides
	parts := strings.Split(filePath, ":")

	filePath = parts[0]
	valuePrefix := ""
	inputType := ""

	if len(parts) > 1 {
		var err error

		// Expand if using homedir alias
		filePath, err = cmdutil.Expand(filePath)
		if err != nil {
			return fmt.Errorf("unable to expand homedir: %w", err)
		}

		// <type>:<path>
		inputType = parts[1]

		// Check prefix usage
		// <path>:<type>:<prefix>
		if len(parts) > 2 {
			valuePrefix = parts[2]
		}
	}

	// Retrieve file type from extension
	fileType := getFileType(filePath, inputType)

	// Retrieve appropriate parser
	p, err := values.GetParser(fileType)
	if err != nil {
		return fmt.Errorf("error occurred during parser instance retrieval for type '%s': %w", fileType, err)
	}

	// Drain file content
	_, err = os.Stat(filePath)
	if err != nil {
		return fmt.Errorf(
			"unable to os.Stat file name %s before attempting to build reader from current directory %s: error: %w",
			filePath,
			currentDirectory,
			err,
		)
	}

	reader, err := cmdutil.Reader(filePath)
	if err != nil {
		return fmt.Errorf("unable to build a reader from '%s' for current directory %s: %w",
			filePath,
			currentDirectory,
			err)
	}

	// Drain reader
	var contentBytes []byte
	contentBytes, err = io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("unable to drain all reader content from '%s': %w", filePath, err)
	}

	// Check prefix
	if valuePrefix != "" {
		// Parse with detected parser
		var fileContent interface{}
		if err := p.Unmarshal(contentBytes, &fileContent); err != nil {
			return fmt.Errorf("unable to unmarshal content from '%s' as '%s': %w", filePath, fileType, err)
		}

		// Re-encode JSON prepending the prefix
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(map[string]interface{}{
			valuePrefix: fileContent,
		}); err != nil {
			return fmt.Errorf("unable to re-encode as JSON with prefix '%s', content from '%s' as '%s': %w", valuePrefix, filePath, fileType, err)
		}

		// Send as result
		if err := json.NewDecoder(&buf).Decode(result); err != nil {
			return fmt.Errorf("unable to decode json content from '%s' parsed as '%s': %w", filePath, fileType, err)
		}
	} else if err := p.Unmarshal(contentBytes, result); err != nil {
		return fmt.Errorf("unable to unmarshal content from '%s' as '%s', you should use an explicit prefix: %w", filePath, fileType, err)
	}

	// No error
	return nil
}

func mergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = mergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}

func getFileType(fileName, input string) string {
	// Format override
	if input != "" {
		return input
	}

	// Stdin filename assumed as YAML
	if fileName == "-" {
		return "yaml"
	}

	// No extension return filename
	if filepath.Ext(fileName) == "" {
		return filepath.Base(fileName)
	}

	// Extract extension
	fileExtension := filepath.Ext(fileName)

	// Return extension whithout the '.'
	return fileExtension[1:]
}
