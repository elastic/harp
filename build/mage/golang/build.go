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
	"errors"
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
//
//nolint:funlen // to refactor
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

		// Check if fips is enabled
		buildTags := "-tags=!fips"
		if os.Getenv("HARP_BUILD_FIPS_MODE") == "1" {
			artifactName = fmt.Sprintf("%s-fips", artifactName)
			compilationFlags = append(compilationFlags, "fips")
			buildTags = "-tags=fips"
		}

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

		// Inject version information
		varsSetByLinker := map[string]string{
			"github.com/elastic/harp/build/version.Name":      name,
			"github.com/elastic/harp/build/version.AppName":   packageName,
			"github.com/elastic/harp/build/version.Version":   version,
			"github.com/elastic/harp/build/version.Commit":    git.Revision,
			"github.com/elastic/harp/build/version.Branch":    git.Branch,
			"github.com/elastic/harp/build/version.BuildDate": time.Now().Format(time.RFC3339),
			"github.com/elastic/harp/build/version.BuildTags": strCompilationFlags,
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

		// Create output directory
		if errMkDir := os.Mkdir("bin", 0o750); errMkDir != nil {
			if !errors.Is(errMkDir, os.ErrExist) {
				return fmt.Errorf("unable to create output directory: %w", errMkDir)
			}
		}

		// Generate output filename
		filename := fmt.Sprintf("bin/%s", artifactName)
		if defaultOpts.goOS == "windows" {
			filename = fmt.Sprintf("%s.exe", filename)
		}

		_, _ = fmt.Fprintf(os.Stdout, " > Generating SBOM %s [%s] [os:%s arch:%s%s flags:%v tag:%v]\n", defaultOpts.binaryName, defaultOpts.packageName, defaultOpts.goOS, defaultOpts.goArch, defaultOpts.goArm, strCompilationFlags, version)

		// Generate SBOM
		if err := sh.RunWith(env, "cyclonedx-gomod", "app", "-json", "-output", fmt.Sprintf("%s.sbom.json", filename), "-files", "-licenses", "-main", fmt.Sprintf("cmd/%s", defaultOpts.binaryName), "-packages"); err != nil {
			return fmt.Errorf("unable to generate SBOM for artifact: %w", err)
		}

		_, _ = fmt.Fprintf(os.Stdout, " > Building %s [%s] [os:%s arch:%s%s flags:%v tag:%v]\n", defaultOpts.binaryName, defaultOpts.packageName, defaultOpts.goOS, defaultOpts.goArch, defaultOpts.goArm, strCompilationFlags, version)

		// Compile
		return sh.RunWith(env, "go", "build", buildMode, buildTags, "-trimpath", "-mod=readonly", "-ldflags", ldflagsValue, "-o", filename, packageName)
	}
}
