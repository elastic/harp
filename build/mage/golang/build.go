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

package golang

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"

	"github.com/elastic/harp/build/mage/git"
)

type buildOpts struct {
	binaryName  string
	packageName string
	cgoEnabled  bool
	pieEnabled  bool
	goOS        string
	goArch      string
	goArm       string
}

// BuildOption is used to define function option pattern.
type BuildOption func(*buildOpts)

// -----------------------------------------------------------------------------

// WithCGO enables CGO compilation
func WithCGO() BuildOption {
	return func(opts *buildOpts) {
		opts.cgoEnabled = true
	}
}

// WithPIE enables Position Independent Executable compilation
func WithPIE() BuildOption {
	return func(opts *buildOpts) {
		opts.pieEnabled = true
	}
}

// GOOS sets the GOOS value during build
func GOOS(value string) BuildOption {
	return func(opts *buildOpts) {
		opts.goOS = value
	}
}

// GOARCH sets the GOARCH value during build
func GOARCH(value string) BuildOption {
	return func(opts *buildOpts) {
		opts.goArch = value
	}
}

// GOARM sets the GOARM value during build
func GOARM(value string) BuildOption {
	return func(opts *buildOpts) {
		opts.goArm = value
	}
}

// -----------------------------------------------------------------------------

// Build the given binary using the given package.
//nolint:funlen // To split
func Build(name, packageName, version string, opts ...BuildOption) func() error {
	const (
		defaultCgoEnabled = false
		defaultGoOs       = runtime.GOOS
		defaultGoArch     = runtime.GOARCH
		defaultGoArm      = ""
	)

	// Default build options
	defaultOpts := &buildOpts{
		binaryName:  name,
		packageName: packageName,
		cgoEnabled:  defaultCgoEnabled,
		goOS:        defaultGoOs,
		goArch:      defaultGoArch,
		goArm:       defaultGoArm,
	}

	// Apply options
	for _, o := range opts {
		o(defaultOpts)
	}

	return func() error {
		// Retrieve git info first
		mg.SerialDeps(git.CollectInfo)

		// Generate artifact name
		artifactName := fmt.Sprintf("%s-%s-%s%s", name, defaultOpts.goOS, defaultOpts.goArch, defaultOpts.goArm)

		// Compilation flags
		compilationFlags := []string{}

		// Check if CGO is enabled
		if defaultOpts.cgoEnabled {
			artifactName = fmt.Sprintf("%s-cgo", artifactName)
			compilationFlags = append(compilationFlags, "cgo")
		}

		// Enable PIE if requested
		buildMode := "-buildmode=exe"
		if defaultOpts.pieEnabled {
			buildMode = "-buildmode=pie"
			artifactName = fmt.Sprintf("%s-pie", artifactName)
			compilationFlags = append(compilationFlags, "pie")
		}

		// Check compilation flags
		strCompilationFlags := "defaults"
		if len(compilationFlags) > 0 {
			strCompilationFlags = strings.Join(compilationFlags, ",")
		}

		fmt.Printf(" > Building %s [%s] [os:%s arch:%s%s flags:%v tag:%v]\n", defaultOpts.binaryName, defaultOpts.packageName, defaultOpts.goOS, defaultOpts.goArch, defaultOpts.goArm, strCompilationFlags, version)

		// Inject version information
		varsSetByLinker := map[string]string{
			"github.com/elastic/harp/build/version.Version":          version,
			"github.com/elastic/harp/build/version.Revision":         git.Revision,
			"github.com/elastic/harp/build/version.Branch":           git.Branch,
			"github.com/elastic/harp/build/version.BuildUser":        os.Getenv("USER"),
			"github.com/elastic/harp/build/version.BuildDate":        time.Now().Format(time.RFC3339),
			"github.com/elastic/harp/build/version.GoVersion":        runtime.Version(),
			"github.com/elastic/harp/build/version.CompilationFlags": strCompilationFlags,
		}
		var linkerArgs []string
		for name, value := range varsSetByLinker {
			linkerArgs = append(linkerArgs, "-X", fmt.Sprintf("'%s=%s'", name, value))
		}

		// Strip and remove DWARF
		linkerArgs = append(linkerArgs, "-s", "-w")

		// Assemble ldflags
		ldflagsValue := strings.Join(linkerArgs, " ")

		// Build environment
		env := map[string]string{
			"GOOS":        defaultOpts.goOS,
			"GOARCH":      defaultOpts.goArch,
			"CGO_ENABLED": "0",
		}
		if defaultOpts.cgoEnabled {
			env["CGO_ENABLED"] = "1"
		}
		if defaultOpts.goArm != "" {
			env["GOARM"] = defaultOpts.goArm
		}

		// Generate output filename
		filename := fmt.Sprintf("bin/%s", artifactName)
		if defaultOpts.goOS == "windows" {
			filename = fmt.Sprintf("%s.exe", filename)
		}

		return sh.RunWith(env, "go", "build", buildMode, "-mod=readonly", "-ldflags", ldflagsValue, "-o", filename, packageName)
	}
}
