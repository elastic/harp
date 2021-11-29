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
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"

	containerv1 "github.com/elastic/harp/api/gen/go/harp/container/v1"
	"github.com/elastic/harp/pkg/sdk/types"
)

const (
	containerMagic             = uint32(0x53CB3701)
	containerVersion           = uint16(0x0002)
	containerSealedContentType = "application/vnd.harp.v1.SealedContainer"
	publicKeySize              = 32
	privateKeySize             = 32
	encryptionKeySize          = 32
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
// nolint:interfacer // Tighly coupled to type
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
