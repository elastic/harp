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

package docker

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	exec "golang.org/x/sys/execabs"

	semver "github.com/Masterminds/semver/v3"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"

	"github.com/elastic/harp/build/artifact"
	"github.com/elastic/harp/build/mage/git"
)

var dockerReleaseTemplate = strings.TrimSpace(`
# syntax=docker/dockerfile:experimental

# Arguments
ARG BUILD_DATE={{.BuildDate}}
ARG VERSION={{.Version}}
ARG VCS_REF={{.VcsRef}}

# Builder arguments
ARG TOOLS_IMAGE={{.ToolImageName}}
ARG RELEASE={{.Release}}

FROM $TOOLS_IMAGE as compiler

ARG RELEASE={{.Release}}

# Back to project root
WORKDIR $GOPATH/src/workspace

# Get dependencies
COPY go.mod .
RUN go mod download

# Copy project go module
COPY --chown=golang:golang . .

{{ if .Cmd.HasModule }}
# Go to cmd
WORKDIR $GOPATH/src/workspace/{{ .Cmd.Module }}
{{ else }}
# Stay at root path
WORKDIR $GOPATH/src/workspace
{{ end }}

# Clean existing binaries
RUN set -eux; \
	rm -f bin/*

# Update vendor
RUN set -eux; \
	go mod vendor

# Set the target release version
ENV RELEASE=$RELEASE

# Build final target
RUN set -eux; \
	mage release

## -------------------------------------------------------------------------------------------------

# Assemble all binary files
FROM alpine:latest AS assembler

WORKDIR /app
{{ if .Cmd.HasModule }}
COPY --from=compiler /go/src/workspace/{{.Cmd.Module}}/bin/* /app/
{{ else }}
COPY --from=compiler /go/src/workspace/bin/* /app/
{{ end }}

## -------------------------------------------------------------------------------------------------

# hadolint ignore=DL3007
FROM alpine:latest

# Arguments
ARG BUILD_DATE={{.BuildDate}}
ARG VERSION={{.Version}}
ARG VCS_REF={{.VcsRef}}
ARG RELEASE={{.Release}}

# Metadata
LABEL \
	org.opencontainers.image.created=$BUILD_DATE \
    org.opencontainers.image.title="{{.Name}}" \
    org.opencontainers.image.description="{{.Cmd.Description}}" \
    org.opencontainers.image.url="https://{{.Cmd.Package}}" \
    org.opencontainers.image.source="https://{{.Cmd.Package}}.git" \
    org.opencontainers.image.revision=$VCS_REF \
    org.opencontainers.image.vendor="Elastic" \
    org.opencontainers.image.version=$RELEASE \
	org.opencontainers.image.licences="ASL2"

WORKDIR /app

COPY --from=assembler /app/* /app/
`)

// Release uses docker pipeline to generate all artifacts.
func Release(cmd *artifact.Command) func() error {
	return func() error {
		mg.Deps(git.CollectInfo)

		// Docker image name
		toolImageName := toolImage
		if os.Getenv("TOOL_IMAGE_NAME") != "" {
			toolImageName = os.Getenv("TOOL_IMAGE_NAME")
		}

		// Extract release
		release := os.Getenv("RELEASE")
		relVer, err := semver.StrictNewVersion(release)
		if err != nil {
			return fmt.Errorf("invalid semver syntax for release: %w", err)
		}

		buf, err := merge(dockerReleaseTemplate, map[string]interface{}{
			"Name":          "Harp CLI Artifacts",
			"ToolImageName": toolImageName,
			"BuildDate":     time.Now().Format(time.RFC3339),
			"Version":       git.Tag,
			"VcsRef":        git.Revision,
			"Cmd":           cmd,
			"Release":       release,
		})
		if err != nil {
			return err
		}

		// Check if we want to generate dockerfile output
		if os.Getenv("DOCKERFILE_ONLY") != "" {
			return os.WriteFile("Dockerfile.release", buf.Bytes(), 0o600)
		}

		// Prepare command
		//nolint:gosec // Expected behavior
		c := exec.Command("docker", "build",
			"-t", fmt.Sprintf("elastic/%s:artifacts-%s", cmd.Kebab(), relVer.String()),
			"-f", "-",
			".",
		)

		// Prepare environment
		c.Env = os.Environ()
		c.Env = append(c.Env, "DOCKER_BUILDKIT=1")
		c.Stderr = os.Stderr
		c.Stdout = os.Stdout
		c.Stdin = buf // Pass Dockerfile as stdin

		// Execute
		err = c.Run()
		if err != nil {
			return fmt.Errorf("unable to run command: %w", err)
		}

		// Check execution error
		if !sh.CmdRan(err) {
			return fmt.Errorf("running '%s' failed with exit code %d", c.String(), sh.ExitStatus(err))
		}

		// No error
		return nil
	}
}

// -----------------------------------------------------------------------------

func merge(t string, m interface{}) (*bytes.Buffer, error) {
	// Compile template
	dockerFileTmpl, err := template.New("Dockerfile").Parse(t)
	if err != nil {
		return nil, fmt.Errorf("unable to compile dockerfile template: %w", err)
	}

	// Merge data
	var buf bytes.Buffer
	if errTmpl := dockerFileTmpl.Execute(&buf, m); errTmpl != nil {
		return nil, errTmpl
	}

	// Return buffer without error
	return &buf, nil
}
