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

package cubbyhole

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/golang/snappy"
	"github.com/hashicorp/vault/api"

	"github.com/elastic/harp/pkg/vault/logical"
	vpath "github.com/elastic/harp/pkg/vault/path"
)

type service struct {
	logical   logical.Logical
	mountPath string
}

// New instantiates a Vault cubbyhole backend service.
func New(client *api.Client, mountPath string) (Service, error) {
	// Apply default cubbyhole mountpath if not overrided.
	if mountPath == "" {
		mountPath = "cubbyhole"
	}

	return &service{
		logical:   client.Logical(),
		mountPath: vpath.SanitizePath(mountPath),
	}, nil
}

// -----------------------------------------------------------------------------

const (
	secretSizeLimit = 1024 * 1024 * 1024 // 1Mb
)

// Put a secret in cubbyhole to retrieve a wrapping token.
//
//nolint:interfacer // -- wants to replace time.Duration by fmt.Stringer
func (s *service) Put(_ context.Context, r io.Reader) (string, error) {
	// Encode secret
	payload, err := io.ReadAll(io.LimitReader(r, secretSizeLimit))
	if err != nil {
		return "", fmt.Errorf("unable to drain secret reader: %w", err)
	}

	// Compress and encode
	final := base64.StdEncoding.EncodeToString(snappy.Encode(nil, payload))

	// Add to cubbyhole
	return addToCubbyhole(s.logical, s.mountPath, final)
}

// Get a secret from wrapping token.
func (s *service) Get(_ context.Context, token string, w io.Writer) error {
	// Unwrap token
	encoded, err := unWrap(s.logical, token)
	if err != nil {
		return err
	}

	// Decode
	payload, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return fmt.Errorf("invalid secret payload: %w", err)
	}

	// Decompress
	final, err := snappy.Decode(nil, payload)
	if err != nil {
		return fmt.Errorf("invalid secret payload: %w", err)
	}

	// Return result to writer
	_, err = w.Write(final)
	if err != nil {
		return fmt.Errorf("unable to write result to the writer: %w", err)
	}

	// No error
	return nil
}
