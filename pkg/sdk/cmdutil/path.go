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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ErrNoHome is raised when tilde expansion failed.
var ErrNoHome = errors.New("no home found")

// Expand a given path using `~` notation for HOMEDIR
func Expand(path string) (string, error) {
	// Check condition
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	// Retrieve HOMEDIR
	home, err := getHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to retrieve user home directory path: %w", err)
	}

	// Return result
	return home + path[1:], nil
}

func getHomeDir() (string, error) {
	home := ""

	switch runtime.GOOS {
	case "windows":
		// Retrieve windows specific env
		home = filepath.Join(os.Getenv("HomeDrive"), os.Getenv("HomePath"))
		if home == "" {
			home = os.Getenv("UserProfile")
		}

	default:
		home = os.Getenv("HOME")
	}

	// Homedir not evaluable ?
	if home == "" {
		return "", ErrNoHome
	}

	// Return result
	return home, nil
}
