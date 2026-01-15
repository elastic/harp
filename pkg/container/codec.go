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

package container

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/awnumar/memguard"
	"google.golang.org/protobuf/proto"

	containerv1 "github.com/elastic/harp/api/gen/go/harp/container/v1"
	"github.com/elastic/harp/pkg/container/identity/key"
	"github.com/elastic/harp/pkg/container/seal"
	v1 "github.com/elastic/harp/pkg/container/seal/v1"
	v2 "github.com/elastic/harp/pkg/container/seal/v2"
	"github.com/elastic/harp/pkg/sdk/types"
)

const (
	containerSealedContentType = "application/vnd.harp.v1.SealedContainer"
	containerMagic             = uint32(0x53CB3701)
	containerVersion           = uint16(0x0002)
)

// Load a reader to extract as a container.
func Load(r io.Reader) (*containerv1.Container, error) {
	// Check parameters
	if types.IsNil(r) {
		return nil, fmt.Errorf("unable to process nil reader")
	}

	// Read magic
	var magic uint32
	if err := binary.Read(r, binary.BigEndian, &magic); err != nil {
		return nil, fmt.Errorf("unable to read magic code: %w", err)
	}

	// Check magic value
	if magic != containerMagic {
		return nil, fmt.Errorf("invalid magic signature")
	}

	// Read container version
	var version uint16
	if err := binary.Read(r, binary.BigEndian, &version); err != nil {
		return nil, fmt.Errorf("unable to read container version: %w", err)
	}

	// Check magic value
	if version != containerVersion {
		return nil, fmt.Errorf("invalid container version %d", version)
	}

	// Drain input reader
	decoded, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("unable to container content")
	}

	// Check content length
	if len(decoded) == 0 {
		return nil, fmt.Errorf("container is empty")
	}

	// Deserialize protobuf payload
	container := &containerv1.Container{}
	if err = proto.Unmarshal(decoded, container); err != nil {
		return nil, fmt.Errorf("unable to decode content as container")
	}

	// Check headers
	if container.Headers == nil {
		container.Headers = &containerv1.Header{}
	}

	// No error
	return container, nil
}

// Dump the marshaled container instance to writer.
func Dump(w io.Writer, c *containerv1.Container) error {
	// Check parameters
	if types.IsNil(w) {
		return fmt.Errorf("unable to process nil writer")
	}
	if c == nil {
		return fmt.Errorf("unable to process nil container")
	}

	// Serialize protobuf payload
	payload, err := proto.Marshal(c)
	if err != nil {
		return fmt.Errorf("unable to encode container content: %w", err)
	}

	// Write packets
	if err = binary.Write(w, binary.BigEndian, containerMagic); err != nil {
		return fmt.Errorf("unable to write container magic: %w", err)
	}
	if err = binary.Write(w, binary.BigEndian, containerVersion); err != nil {
		return fmt.Errorf("unable to write container version: %w", err)
	}
	if _, err = w.Write(payload); err != nil {
		return fmt.Errorf("unable to write container content: %w", err)
	}

	// No error
	return nil
}

// Unseal a sealed container with the given identity
func Unseal(container *containerv1.Container, identity *memguard.LockedBuffer) (*containerv1.Container, error) {
	// Check parameters
	if types.IsNil(container) {
		return nil, fmt.Errorf("unable to process nil container")
	}
	if types.IsNil(container.Headers) {
		return nil, fmt.Errorf("unable to process nil container headers")
	}
	if identity == nil {
		return nil, fmt.Errorf("unable to process without container key")
	}

	// Check headers
	if container.Headers.ContentType != containerSealedContentType {
		return nil, fmt.Errorf("unable to unseal container")
	}

	// Build appropriate unseal strategy processor.
	var ss seal.Strategy
	switch container.Headers.SealVersion {
	case 2:
		ss = v2.New()
	default:
		ss = v1.New()
	}

	// Delegate to strategy
	return ss.Unseal(container, identity)
}

// IsSealed returns true if the given container is sealed.
func IsSealed(container *containerv1.Container) bool {
	// Check parameters
	if types.IsNil(container) {
		return false
	}
	if types.IsNil(container.Headers) {
		return false
	}

	// Check headers
	if container.Headers.ContentType != containerSealedContentType {
		return false
	}

	// Check sealing algorithm
	if container.Headers.SealVersion == 0 {
		return false
	}

	// Default sealed
	return true
}

func Seal(rand io.Reader, container *containerv1.Container, encodedPeersPublicKey ...string) (*containerv1.Container, error) {
	// Check parameters
	if types.IsNil(container) {
		return nil, errors.New("unable to seal a nil container")
	}
	if IsSealed(container) {
		return nil, errors.New("the container is already sealed")
	}

	// Validate peer public keys
	hasV1 := false
	hasV2 := false
	for i, pub := range encodedPeersPublicKey {
		switch {
		case strings.HasPrefix(pub, "v1.sk."):
			hasV1 = true
		case strings.HasPrefix(pub, "v1.ipk."):
			hasV1 = true

			// Convert to sealing public key
			identityPublicKey, err := key.FromString(pub)
			if err != nil {
				return nil, fmt.Errorf("unable to convert v1 identity public key '%s': %w", pub, err)
			}

			// Replace identity key by sealing key
			encodedPeersPublicKey[i] = identityPublicKey.SealingKey()
		case strings.HasPrefix(pub, "v2.sk."):
			hasV2 = true
		case strings.HasPrefix(pub, "v2.ipk."):
			hasV2 = true

			// Convert to sealing public key
			identityPublicKey, err := key.FromString(pub)
			if err != nil {
				return nil, fmt.Errorf("unable to convert v2 identity public key '%s': %w", pub, err)
			}

			// Replace identity key by sealing key
			encodedPeersPublicKey[i] = identityPublicKey.SealingKey()
		default:
			return nil, fmt.Errorf("invalid key '%s'", pub)
		}
	}
	if hasV1 && hasV2 {
		return nil, errors.New("peer public keys are using mixed versions - use v1 or v2 keys")
	}

	// Create sealing strategy instance
	var ss seal.Strategy
	switch {
	case hasV1:
		ss = v1.New()
	case hasV2:
		ss = v2.New()
	default:
		return nil, errors.New("unsupported sealing algorithm")
	}

	// Delegate to strategy
	return ss.Seal(rand, container, encodedPeersPublicKey...)
}
