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

package secret

import (
	"encoding/asn1"
	"fmt"

	msgpack "github.com/vmihailenco/msgpack/v5"
)

const (
	formatVersion = int(0x00000001)
)

// Pack a secret value.
func Pack(value interface{}) ([]byte, error) {
	// Encode the payload
	payload, err := msgpack.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("unable to pack secret value: %w", err)
	}

	// Pack header
	header, err := asn1.Marshal(formatVersion)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal header of sequence: %w", err)
	}

	// Pack body
	body, err := asn1.Marshal(asn1.RawValue{
		Class:      asn1.ClassUniversal,
		IsCompound: true,
		Tag:        asn1.TagSequence,
		Bytes:      append(header, payload...),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to marshal final sequence: %w", err)
	}

	// No error
	return body, nil
}

// Unpack a secret value.
func Unpack(in []byte, out interface{}) error {
	var raw asn1.RawValue

	_, err := asn1.Unmarshal(in, &raw)
	if err != nil {
		return fmt.Errorf("unable to unpack secret header: %w", err)
	}
	if raw.Class != asn1.ClassUniversal || raw.Tag != asn1.TagSequence || !raw.IsCompound {
		return asn1.StructuralError{Msg: fmt.Sprintf(
			"invalid packed structure object - class [%02x], tag [%02x]",
			raw.Class, raw.Tag)}
	}

	var version int
	rest, err := asn1.Unmarshal(raw.Bytes, &version)
	if err != nil {
		return fmt.Errorf("unable to unpack format version: %w", err)
	}

	// Compare with expected
	if version != formatVersion {
		return fmt.Errorf("unexpected packed version, received %d, expected %d", version, formatVersion)
	}

	// Decode the value
	if err := msgpack.Unmarshal(rest, out); err != nil {
		return fmt.Errorf("unable to unpack secret value: %w", err)
	}

	return nil
}
