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

package fsutil

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/sdk/log"
)

// Dump the given vfs to the outputpath
func Dump(srcFs fs.FS, outPath string) error {
	return fs.WalkDir(srcFs, ".", func(path string, d fs.DirEntry, errWalk error) error {
		// Raise immediately the error if any.
		if errWalk != nil {
			return fmt.Errorf("%s: %w", path, errWalk)
		}

		// Ignore directory
		if d.IsDir() {
			return nil
		}

		// Compute the target path
		targetPath := filepath.Join(outPath, path)

		// Extract relative directory
		relativeDir := filepath.Dir(targetPath)

		// Check folder hierarchy existence.
		if _, err := os.Stat(relativeDir); os.IsNotExist(err) {
			if err := os.MkdirAll(relativeDir, 0o750); err != nil {
				return fmt.Errorf("unable to create intermediate directories for path '%s': %w", relativeDir, err)
			}
		}

		// Create file
		//nolint:gosec // G304: targetPath is derived from srcFs path, not user input
		targetFile, err := os.Create(targetPath)
		if err != nil {
			return fmt.Errorf("unable to create the output file: %w", err)
		}

		// Open input file
		srcFile, err := srcFs.Open(path)
		if err != nil {
			return fmt.Errorf("unable to open source file: %w", err)
		}

		log.Bg().Debug("Copy file ...", zap.String("file", path))

		// Open the target file
		if _, err := io.Copy(targetFile, srcFile); err != nil {
			if !errors.Is(err, io.EOF) {
				return fmt.Errorf("unable to copy content from '%s' to '%s': %w", path, targetPath, err)
			}
		}

		// Close the file
		return srcFile.Close()
	})
}
