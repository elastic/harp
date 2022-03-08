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

package crate

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/target"

	"github.com/elastic/harp/pkg/crate"
	"github.com/elastic/harp/pkg/sdk/log"
)

// CopyTask implements secret-container pulling process to and OCI compatible registry.
type CopyTask struct {
	Source                  string
	SourceRef               string
	SourceRegistryOpts      content.RegistryOptions
	Destination             string
	DestinationRef          string
	DestinationRegistryOpts content.RegistryOptions
}

// Run the task.
func (t *CopyTask) Run(ctx context.Context) error {
	// Create source resolver
	from, err := getResolver(t.Source, t.SourceRegistryOpts)
	if err != nil {
		return fmt.Errorf("unable to prepare source resolver: %w", err)
	}

	// Create destination resolver
	to, err := getResolver(t.Destination, t.DestinationRegistryOpts)
	if err != nil {
		return fmt.Errorf("unable to prepare destination resolver: %w", err)
	}

	// Pull the container
	m, err := crate.Pull(ctx, from, t.SourceRef, to)
	if err != nil {
		return fmt.Errorf("unable to push container to registry: %w", err)
	}

	log.For(ctx).Info("Crate '%s' succcessfully pulled !", zap.String("digest", m.Digest.Hex()))

	// No error
	return nil
}

// -----------------------------------------------------------------------------

func getResolver(targetContent string, registryOpts content.RegistryOptions) (target.Target, error) {
	var (
		out target.Target
		err error
	)

	// Split target content
	parts := strings.SplitN(targetContent, ":", 2)

	// Build appropriate target instance
	switch parts[0] {
	case "files":
		out = content.NewFile(parts[1])
	case "registry":
		out, err = content.NewRegistry(registryOpts)
		if err != nil {
			return nil, fmt.Errorf("could not create registry resolver: %w", err)
		}
	case "oci":
		out, err = content.NewOCI(parts[1])
		if err != nil {
			return nil, fmt.Errorf("could not read OCI layout at %s: %w", parts[1], err)
		}
	default:
		return nil, fmt.Errorf("unknown resolver argument: %s", targetContent)
	}

	// No error
	return out, nil
}
