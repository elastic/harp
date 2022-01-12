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
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	exec "golang.org/x/sys/execabs"

	"github.com/elastic/harp/build/mage/git"
)

var dockerToolTemplate = strings.TrimSpace(`
# syntax=docker/dockerfile:experimental

# Arguments
ARG BUILD_DATE={{.BuildDate}}
ARG VERSION={{.Version}}
ARG VCS_REF={{.VcsRef}}

# Builder arguments
ARG GOLANG_BASE_IMAGE={{.GolangImage}}
ARG GOLANG_VERSION={{.GolangVersion}}

## -------------------------------------------------------------------------------------------------

FROM ${GOLANG_BASE_IMAGE}

# Arguments
ARG BUILD_DATE={{.BuildDate}}
ARG VERSION={{.Version}}
ARG VCS_REF={{.VcsRef}}

# Builder argumentsgolang
ARG GOLANG_IMAGE={{.GolangImage}}
ARG GOLANG_VERSION={{.GolangVersion}}

LABEL \
    org.opencontainers.image.created=$BUILD_DATE \
    org.opencontainers.image.title="Harp SDK Environment (Go {{.GolangVersion}})" \
    org.opencontainers.image.description="Harp SDK Tools used to build harp and all related tools" \
    org.opencontainers.image.url="https://github.com/elastic/harp" \
    org.opencontainers.image.source="https://github.com/elastic/harp.git" \
    org.opencontainers.image.revision=$VCS_REF \
    org.opencontainers.image.vendor="Elastic" \
    org.opencontainers.image.version=$VERSION \
    org.opencontainers.image.licences="ASL2"

{{ if .OverrideGoBoringVersion }}
# Override goboring version
RUN wget https://storage.googleapis.com/go-boringcrypto/go{{ .GoBoringVersion }}.linux-amd64.tar.gz \
    && rm -rf /usr/local/go && tar -C /usr/local -xzf go{{ .GoBoringVersion }}.linux-amd64.tar.gz \
    rm go{{ .GoBoringVersion }}.linux-amd64.tar.gz
{{ end }}

# hadolint ignore=DL3008
RUN set -eux; \
    apt-get update -y && \
    apt-get install -y --no-install-recommends apt-utils bzr upx zip unzip;

RUN go version

# Create a non-root privilege account to build
RUN adduser --disabled-password --gecos "" -u 1000 golang && \
    mkdir -p "$GOPATH/src/workspace" && \
    chown -R golang:golang "$GOPATH/src/workspace" && \
    mkdir /home/golang/.ssh && \
    mkdir /var/ssh && \
    chown -R golang:golang /home/golang && \
    chown -R golang:golang /var/ssh && \
    chmod 700 /home/golang

# Force go modules
ENV GO111MODULE=on

# Disable go proxy
ENV GOPROXY=direct
ENV GOSUMDB=off

WORKDIR $GOPATH/src/workspace

# Prepare an unprivilegied user for run
RUN set -eux; \
    echo 'nobody:x:65534:65534:nobody:/:' > /tmp/passwd && \
    echo 'nobody:x:65534:' > /tmp/group && \
    mkdir /tmp/.config && \
    chown 65534:65534 /tmp/.config

# Drop privileges to build
USER golang
ENV USER golang

# Clean go mod cache
RUN set -eux; \
    go clean -modcache

# Checkout mage
RUN set -eux; \
    git clone https://github.com/magefile/mage .mage

# Go to tools
WORKDIR $GOPATH/src/workspace/.mage

# Install mage
RUN go run bootstrap.go

# Back to project root
WORKDIR $GOPATH/src/workspace

# Copy build tools
COPY --chown=golang:golang tools tools/

# Go to tools
WORKDIR $GOPATH/src/workspace/tools

# Install tools
RUN set -eux; \
    mage

# Set path for tools usages
ENV PATH=$GOPATH/src/workspace/tools/bin:$PATH
`)

// Tools build a docker container used for compilation.
func Tools() error {
	mg.Deps(git.CollectInfo)

	// Retrieve golang attributes
	golangBaseImage := golangImage
	if os.Getenv("GOLANG_BASE_IMAGE") != "" {
		golangBaseImage = os.Getenv("GOLANG_BASE_IMAGE")
	}
	golangVersion := golangVersion
	if os.Getenv("GOLANG_VERSION") != "" {
		golangVersion = os.Getenv("GOLANG_VERSION")
	}
	goBoringVersion := goBoringVersion
	overrideGoBoringVersion := false
	if os.Getenv("GOBORING_VERSION") != "" {
		goBoringVersion = os.Getenv("GOBORING_VERSION")
		overrideGoBoringVersion = true
	}

	buf, err := merge(dockerToolTemplate, map[string]interface{}{
		"BuildDate":               time.Now().Format(time.RFC3339),
		"Version":                 git.Tag,
		"VcsRef":                  git.Revision,
		"GolangImage":             golangBaseImage,
		"GolangVersion":           golangVersion,
		"OverrideGoBoringVersion": overrideGoBoringVersion,
		"GoBoringVersion":         goBoringVersion,
	})
	if err != nil {
		return err
	}

	// Check if we want to generate dockerfile output
	if os.Getenv("DOCKERFILE_ONLY") != "" {
		return os.WriteFile("Dockerfile.tools", buf.Bytes(), 0o600)
	}

	// Docker image name
	dockerImageName := toolImage
	if os.Getenv("DOCKER_IMAGE_NAME") != "" {
		dockerImageName = os.Getenv("DOCKER_IMAGE_NAME")
	}

	// Prepare command
	c := exec.Command("docker", "build",
		"-t", dockerImageName,
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

	return err
}
