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
ARG BUILD_DATE
ARG VERSION
ARG VCS_REF

# Builder arguments
ARG GOLANG_IMAGE=golang
ARG GOLANG_VERSION=1.17

## -------------------------------------------------------------------------------------------------

FROM ${GOLANG_IMAGE}:${GOLANG_VERSION}

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

	buf, err := merge(dockerToolTemplate, nil)
	if err != nil {
		return err
	}

	// Check if we want to generate dockerfile output
	if os.Getenv("DOCKERFILE_ONLY") != "" {
		fmt.Fprintln(os.Stdout, buf.String())
		return nil
	}

	// Retrieve golang attributes
	golangImage := "golang"
	if os.Getenv("GOLANG_IMAGE") != "" {
		golangImage = os.Getenv("GOLANG_IMAGE")
	}
	golangVersion := "1.17"
	if os.Getenv("GOLANG_VERSION") != "" {
		golangVersion = os.Getenv("GOLANG_VERSION")
	}

	// Docker image name
	dockerImageName := "elastic/harp-tools"
	if os.Getenv("DOCKER_IMAGE_NAME") != "" {
		dockerImageName = os.Getenv("DOCKER_IMAGE_NAME")
	}

	// Prepare command
	//nolint:gosec // expected behavior
	c := exec.Command("docker", "build",
		"-t", dockerImageName,
		"-f", "-",
		"--build-arg", fmt.Sprintf("BUILD_DATE=%s", time.Now().Format(time.RFC3339)),
		"--build-arg", fmt.Sprintf("VERSION=%s", git.Tag),
		"--build-arg", fmt.Sprintf("VCS_REF=%s", git.Revision),
		"--build-arg", fmt.Sprintf("GOLANG_IMAGE=%s", golangImage),
		"--build-arg", fmt.Sprintf("GOLANG_VERSION=%s", golangVersion),
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
