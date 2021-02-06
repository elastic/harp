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

//+build mage

package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/elastic/harp/build/artifact"
	"github.com/elastic/harp/build/mage/git"
	"github.com/elastic/harp/build/mage/golang"
	"github.com/elastic/harp/build/mage/release"

	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
	"github.com/magefile/mage/mg"
)

var Default = Build

var descriptor = &artifact.Command{
	Package:     "github.com/elastic/harp",
	Module:      "cmd/harp-server",
	Name:        "Harp Server",
	Description: "Harp Crate Server",
}

// Build the artefact
func Build() error {
	banner := figure.NewFigure("Harp Server", "", true)
	banner.Print()

	fmt.Println("")
	color.Red("# Build Info ---------------------------------------------------------------")
	fmt.Printf("Go version : %s\n", runtime.Version())

	version, err := git.TagMatch("cmd/harp-server/v*")
	if err != nil {
		return err
	}

	fmt.Printf("Git tag    : %s\n", version)

	color.Red("# Pipeline -----------------------------------------------------------------")
	mg.SerialDeps(golang.Vendor, golang.License("../../"), Generate, golang.Lint("../../"), Test)

	color.Red("# Artifact(s) --------------------------------------------------------------")
	mg.Deps(Compile)

	return nil
}

// Generate code
func Generate() {
	color.Cyan("## Generate code")

	color.Blue("### Dispatchers")
	golang.Generate("HTTP", "github.com/elastic/harp/cmd/harp-server/internal/dispatchers/http")()
	golang.Generate("Vault", "github.com/elastic/harp/cmd/harp-server/internal/dispatchers/vault")()
	golang.Generate("gRPC", "github.com/elastic/harp/cmd/harp-server/internal/dispatchers/grpc")()
}

// Test application
func Test() {
	color.Cyan("## Tests")
	mg.SerialDeps(
		func() error {
			return golang.UnitTest("github.com/elastic/harp/cmd/harp-server/internal/...")()
		},
	)
}

// Compile artefacts
func Compile() error {
	// Extract git version
	version, err := git.TagMatch("cmd/harp/v*")
	if err != nil {
		return err
	}

	return golang.Build("harp-server", "github.com/elastic/harp/cmd/harp-server", version)()
}

// Release
func Release(ctx context.Context) error {
	color.Red(fmt.Sprintf("# Releasing (%s) ---------------------------------------------------------------", os.Getenv("RELEASE")))

	// Extract git version
	version, err := git.TagMatch("cmd/harp-server/v*")
	if err != nil {
		return err
	}

	color.Cyan("## Cross compiling artifact")

	mg.CtxDeps(ctx,
		func() error {
			return golang.Release(
				"harp-server",
				"github.com/elastic/harp/cmd/harp-server",
				version,
				golang.GOOS("darwin"), golang.GOARCH("amd64"),
			)()
		},
		func() error {
			if !golang.Is("go1.16rc1", "go1.16") {
				// Skip the build
				return nil
			}
			return golang.Release(
				"harp-server",
				"github.com/elastic/harp/cmd/harp-server",
				version,
				golang.GOOS("darwin"), golang.GOARCH("arm64"),
			)()
		},
		func() error {
			return golang.Release(
				"harp-server",
				"github.com/elastic/harp/cmd/harp-server",
				version,
				golang.GOOS("linux"), golang.GOARCH("amd64"),
			)()
		},
		func() error {
			return golang.Release(
				"harp-server",
				"github.com/elastic/harp/cmd/harp-server",
				version,
				golang.GOOS("linux"), golang.GOARCH("arm"), golang.GOARM("7"),
			)()
		},
		func() error {
			return golang.Release(
				"harp-server",
				"github.com/elastic/harp/cmd/harp-server",
				version,
				golang.GOOS("linux"), golang.GOARCH("arm64"),
			)()
		},
		func() error {
			return golang.Release(
				"harp-server",
				"github.com/elastic/harp/cmd/harp-server",
				version,
				golang.GOOS("windows"), golang.GOARCH("amd64"),
			)()
		},
	)

	return ctx.Err()
}

func Homebrew() error {
	return release.HomebrewFormula(descriptor)()
}
