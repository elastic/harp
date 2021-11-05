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

	"github.com/elastic/harp/build/artifact"
	"github.com/elastic/harp/build/mage/git"
)

var dockerTemplate = strings.TrimSpace(`
# syntax=docker/dockerfile:experimental

# Arguments
ARG BUILD_DATE
ARG VERSION
ARG VCS_REF

FROM elastic/harp-tools as compiler

# Back to project root
WORKDIR $GOPATH/src/workspace

# Copy go.mod
COPY --chown=golang:golang go.mod .
COPY --chown=golang:golang go.sum .
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
    mage

# Compress binaries
RUN set -eux; \
    upx -9 bin/* && \
    chmod +x bin/*

## -------------------------------------------------------------------------------------------------

# hadolint ignore=DL3007
FROM gcr.io/distroless/static:latest

# Arguments
ARG BUILD_DATE
ARG VERSION
ARG VCS_REF

# Metadata
LABEL \
    org.label-schema.build-date=$BUILD_DATE \
    org.label-schema.name="{{.Name}}" \
    org.label-schema.description="{{.Description}}" \
    org.label-schema.url="https://{{.Package}}" \
    org.label-schema.vcs-url="https://{{.Package}}.git" \
    org.label-schema.vcs-ref=$VCS_REF \
    org.label-schema.vendor="Elastic" \
    org.label-schema.version=$VERSION \
    org.label-schema.schema-version="1.0"

{{ if .HasModule }}
COPY --from=compiler /go/src/workspace/{{.Module}}/bin/{{.Kebab}}-linux-amd64 /usr/bin/{{.Kebab}}
{{ else }}
COPY --from=compiler /go/src/workspace/bin/{{.Kebab}}-linux-amd64 /usr/bin/{{.Kebab}}
{{ end }}

COPY --from=compiler /tmp/group /tmp/passwd /etc/
COPY --from=compiler --chown=65534:65534 /tmp/.config /

USER nobody:nobody
WORKDIR /

ENTRYPOINT [ "/usr/bin/{{.Kebab}}" ]
CMD ["--help"]
`)

// Build a docker container for given command.
func Build(cmd *artifact.Command) func() error {
	return func() error {
		mg.Deps(git.CollectInfo)

		buf, err := merge(dockerTemplate, cmd)
		if err != nil {
			return err
		}

		// Prepare command
		//nolint:gosec // expected behavior
		c := exec.Command("docker", "build",
			"-t", fmt.Sprintf("elastic/%s", cmd.Kebab()),
			"-f", "-",
			"--build-arg", fmt.Sprintf("BUILD_DATE=%s", time.Now().Format(time.RFC3339)),
			"--build-arg", fmt.Sprintf("VERSION=%s", git.Tag),
			"--build-arg", fmt.Sprintf("VCS_REF=%s", git.Revision),
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
