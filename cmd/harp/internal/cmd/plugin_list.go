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
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/elastic/harp/pkg/sdk/cmdutil"
)

// ValidPluginFilenamePrefixes defines harp plugin prefix to discover
// cli plugin.
var validPluginFilenamePrefixes = []string{"harp", "harp"}

// -----------------------------------------------------------------------------

// PluginListOptions describes `plugin list` command options.
type pluginListOptions struct {
	NameOnly    bool
	PluginPaths []string
	Verifier    pathVerifier
}

var pluginListCmd = func() *cobra.Command {
	o := &pluginListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List discovered cmd plugins",
		Run: func(cmd *cobra.Command, _ []string) {
			if err := o.Complete(cmd); err != nil {
				fmt.Fprintf(os.Stderr, "%v", err)
				os.Exit(-1)
				return
			}
			if err := o.Run(cmd); err != nil {
				fmt.Fprintf(os.Stderr, "%v", err)
				os.Exit(-1)
				return
			}
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&o.NameOnly, "name-only", o.NameOnly, "If true, display only the binary name of each plugin, rather than its full path")

	return cmd
}

func (o *pluginListOptions) Complete(cmd *cobra.Command) error {
	o.PluginPaths = filepath.SplitList(os.Getenv("PATH"))
	o.Verifier = &commandOverrideVerifier{
		root:        cmd.Root(),
		seenPlugins: make(map[string]string),
	}
	return nil
}

//nolint:gocyclo,gocognit // refactor imported code
func (o *pluginListOptions) Run(cmd *cobra.Command) error {
	_, cancel := cmdutil.Context(cmd.Context(), "harp-plugin-list", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
	defer cancel()

	var (
		pluginsFound   = false
		isFirstFile    = true
		pluginErrors   = []error{}
		pluginWarnings = 0
	)

	// Deduplicate plugin paths
	filteredPaths := uniquePathsList(o.PluginPaths)

	// For each path
	for _, dir := range filteredPaths {
		// Ignore empty dir
		if strings.TrimSpace(dir) == "" {
			continue
		}

		// Crawl each directory to identify readable ones
		files, err := os.ReadDir(dir)
		if err != nil {
			var pathErr *os.PathError
			if errors.As(err, &pathErr) {
				fmt.Fprintf(os.Stderr, "Unable read directory %q from your PATH: %v. Skipping...\n", dir, pathErr)
				continue
			}

			pluginErrors = append(pluginErrors, fmt.Errorf("error: unable to read directory %q in your PATH: %w", dir, err))
			continue
		}

		// Crawl each files
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			if !hasValidPrefix(f.Name(), validPluginFilenamePrefixes) {
				continue
			}

			// First file identified
			if isFirstFile {
				fmt.Fprintf(os.Stdout, "The following compatible plugins are available:\n\n")
				pluginsFound = true
				isFirstFile = false
			}

			// Display name according to flag
			pluginPath := f.Name()
			if !o.NameOnly {
				pluginPath = filepath.Join(dir, pluginPath)
			}

			fmt.Fprintf(os.Stdout, "%s\n", pluginPath)
			if errs := o.Verifier.Verify(filepath.Join(dir, f.Name())); len(errs) != 0 {
				for _, err := range errs {
					fmt.Fprintf(os.Stderr, "  - %s\n", err)
					pluginWarnings++
				}
			}
		}
	}

	if !pluginsFound {
		pluginErrors = append(pluginErrors, fmt.Errorf("error: unable to find any harp plugins in your PATH"))
	}

	if pluginWarnings > 0 {
		if pluginWarnings == 1 {
			pluginErrors = append(pluginErrors, fmt.Errorf("error: one plugin warning was found"))
		} else {
			pluginErrors = append(pluginErrors, fmt.Errorf("error: %v plugin warnings were found", pluginWarnings))
		}
	}
	if len(pluginErrors) > 0 {
		errs := bytes.NewBuffer(nil)
		for _, e := range pluginErrors {
			fmt.Fprintln(errs, e)
		}
		return fmt.Errorf("%s", errs.String())
	}

	return nil
}

// -----------------------------------------------------------------------------

// pathVerifier receives a path and determines if it is valid or not
type pathVerifier interface {
	// Verify determines if a given path is valid
	Verify(path string) []error
}

type commandOverrideVerifier struct {
	root        *cobra.Command
	seenPlugins map[string]string
}

// Verify implements PathVerifier and determines if a given path
// is valid depending on whether or not it overwrites an existing
// harp command path, or a previously seen plugin.
func (v *commandOverrideVerifier) Verify(path string) []error {
	if v.root == nil {
		return []error{fmt.Errorf("unable to verify path with nil root")}
	}

	// extract the plugin binary name
	segs := strings.Split(path, "/")
	binName := segs[len(segs)-1]

	cmdPath := strings.Split(binName, "-")
	if len(cmdPath) > 1 {
		// the first argument is always "harp" for a plugin binary
		cmdPath = cmdPath[1:]
	}

	errorList := []error{}

	if isExec, err := isExecutable(path); err == nil && !isExec {
		errorList = append(errorList, fmt.Errorf("warning: %s identified as a harp plugin, but it is not executable", path))
	} else if err != nil {
		errorList = append(errorList, fmt.Errorf("error: unable to identify %s as an executable file: %w", path, err))
	}

	if existingPath, ok := v.seenPlugins[binName]; ok {
		errorList = append(errorList, fmt.Errorf("warning: %s is overshadowed by a similarly named plugin: %s", path, existingPath))
	} else {
		v.seenPlugins[binName] = path
	}

	if cmd, _, err := v.root.Find(cmdPath); err == nil {
		errorList = append(errorList, fmt.Errorf("warning: %s overwrites existing command: %q", binName, cmd.CommandPath()))
	}

	return errorList
}

func isExecutable(fullPath string) (bool, error) {
	info, err := os.Stat(fullPath)
	if err != nil {
		return false, err
	}

	if m := info.Mode(); !m.IsDir() && m&0o111 != 0 {
		return true, nil
	}

	return false, nil
}

// uniquePathsList deduplicates a given slice of strings without
// sorting or otherwise altering its order in any way.
func uniquePathsList(paths []string) []string {
	seen := map[string]bool{}
	newPaths := []string{}
	for _, p := range paths {
		if seen[p] {
			continue
		}
		seen[p] = true
		newPaths = append(newPaths, p)
	}
	return newPaths
}

func hasValidPrefix(filePath string, validPrefixes []string) bool {
	for _, prefix := range validPrefixes {
		if !strings.HasPrefix(filePath, prefix+"-") {
			continue
		}
		return true
	}
	return false
}
