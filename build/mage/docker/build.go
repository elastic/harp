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
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"

	"github.com/elastic/harp/build/artifact"
	"github.com/elastic/harp/build/mage/git"
)

var dockerTemplate = strings.TrimSpace(`
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

		// Invoke docker commands
		err = sh.RunWith(
			map[string]string{
				"DOCKER_BUILDKIT": "1",
			},
			"/bin/sh", "-c",
			fmt.Sprintf("echo '%s' | base64 -D | docker build -t elastic/%s -f- --build-arg BUILD_DATE=%s --build-arg VERSION=%s --build-arg VCS_REF=%s .", base64.StdEncoding.EncodeToString(buf.Bytes()), cmd.Kebab(), time.Now().Format(time.RFC3339), git.Tag, git.Revision),
		)

		return err
	}
}
