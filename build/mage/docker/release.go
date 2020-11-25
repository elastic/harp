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
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"

	"github.com/elastic/harp/build/artifact"
	"github.com/elastic/harp/build/mage/git"
)

var dockerReleaseTemplate = strings.TrimSpace(`
# syntax=docker/dockerfile:experimental

# Arguments
ARG BUILD_DATE
ARG VERSION
ARG VCS_REF
ARG RELEASE

FROM elastic/harp-tools as compiler

ARG RELEASE

# Back to project root
WORKDIR $GOPATH/src/workspace

# Get dependencies
COPY go.mod .
RUN go mod download

# Copy project go module
COPY --chown=golang:golang . .

{{ if .HasModule }}
# Go to cmd
WORKDIR $GOPATH/src/workspace/{{ .Module }}
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

# Build final target
RUN set -eux; \
	mage release

## -------------------------------------------------------------------------------------------------

# Compress to binary file
FROM alpine:latest AS compressor

RUN apk add --no-cache upx
WORKDIR /app
{{ if .HasModule }}
COPY --from=compiler /go/src/workspace/{{.Module}}/bin/* /app/
{{ else }}
COPY --from=compiler /go/src/workspace/bin/* /app/
{{ end }}
RUN upx --overlay=strip -9 *

## -------------------------------------------------------------------------------------------------

# hadolint ignore=DL3007
FROM alpine:latest

# Arguments
ARG BUILD_DATE
ARG VCS_REF
ARG RELEASE

# Metadata
LABEL \
    org.label-schema.build-date=$BUILD_DATE \
    org.label-schema.name="{{.Name}}" \
    org.label-schema.description="{{.Description}}" \
    org.label-schema.url="https://{{.Package}}" \
    org.label-schema.vcs-url="https://{{.Package}}.git" \
    org.label-schema.vcs-ref=$VCS_REF \
    org.label-schema.vendor="Elastic" \
    org.label-schema.version=$RELEASE \
    org.label-schema.schema-version="1.0"

WORKDIR /app

COPY --from=compressor /app/* /app/
`)

// Release uses docker pipeline to generate all artifacts.
func Release(cmd *artifact.Command) func() error {
	return func() error {
		mg.Deps(git.CollectInfo)

		buf, err := merge(dockerReleaseTemplate, cmd)
		if err != nil {
			return err
		}

		// Extract release
		release := os.Getenv("RELEASE")

		// Invoke docker commands
		err = sh.RunWith(
			map[string]string{
				"DOCKER_BUILDKIT": "1",
			},
			"/bin/sh", "-c",
			fmt.Sprintf("echo '%s' | base64 -D | docker build -t elastic/%s:artifacts-%s -f- --build-arg BUILD_DATE=%s --build-arg VERSION=%s --build-arg VCS_REF=%s --build-arg RELEASE=%s --cache-from=elastic/harp-tools --cache-from=elastic/%s:artifacts-%s .", base64.StdEncoding.EncodeToString(buf.Bytes()), cmd.Kebab(), release, time.Now().Format(time.RFC3339), git.Tag, git.Revision, release, cmd.Kebab(), release),
		)

		return err
	}
}

// -----------------------------------------------------------------------------

func merge(t string, cmd *artifact.Command) (*bytes.Buffer, error) {
	// Compile template
	dockerFileTmpl, err := template.New("Dockerfile").Parse(t)
	if err != nil {
		return nil, err
	}

	// Merge data
	var buf bytes.Buffer
	if errTmpl := dockerFileTmpl.Execute(&buf, cmd); errTmpl != nil {
		return nil, errTmpl
	}

	// Return buffer without error
	return &buf, nil
}
