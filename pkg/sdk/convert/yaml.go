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

package convert

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"sigs.k8s.io/yaml"

	"github.com/elastic/harp/pkg/sdk/types"
)

// YAMLtoJSON reads a given reader in order to extract a JSON representation
func YAMLtoJSON(r io.Reader) (io.Reader, error) {
	// Check arguments
	if types.IsNil(r) {
		return nil, fmt.Errorf("reader is nil")
	}

	// Drain the reader
	jsonReader, err := loadFromYAML(r)
	if err != nil {
		return nil, fmt.Errorf("unable to parse YAML input: %w", err)
	}

	// No error
	return jsonReader, nil
}

// PBtoYAML converts a protobuf object to a YAML representation
func PBtoYAML(msg proto.Message) ([]byte, error) {
	// Check arguments
	if types.IsNil(msg) {
		return nil, fmt.Errorf("msg is nil")
	}

	// Encode protbuf message as JSON
	pb, err := protojson.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("unable to encode protbuf message to JSON: %w", err)
	}

	// Decode input as JSON
	var jsonObj interface{}
	if errDecode := yaml.Unmarshal(pb, &jsonObj); errDecode != nil {
		return nil, fmt.Errorf("unable to decode JSON input: %w", errDecode)
	}

	// Marshal as YAML
	out, errEncode := yaml.Marshal(jsonObj)
	if errEncode != nil {
		return nil, fmt.Errorf("unable to produce YAML output: %w", errEncode)
	}

	// No error
	return out, nil
}

// -----------------------------------------------------------------------------

// loadFromYAML reads YAML definition and returns the PB struct.
//
// Protobuf doesn't contain YAML struct tags and json one are not symetric
// to protobuf. We need to export YAML as JSON, and then read JSON to Protobuf
// as done in k8s yaml loader.
func loadFromYAML(r io.Reader) (io.Reader, error) {
	// Check arguments
	if types.IsNil(r) {
		return nil, fmt.Errorf("reader is nil")
	}

	// Drain input reader
	in, err := io.ReadAll(r)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("unable to drain input reader: %w", err)
	}

	// Decode as YAML any object
	var specBody interface{}
	if errYaml := yaml.Unmarshal(in, &specBody); errYaml != nil {
		return nil, fmt.Errorf("unable to decode spec as YAML: %w", err)
	}

	// Convert map[interface{}]interface{} to a JSON serializable struct
	specBody, err = convertMapStringInterface(specBody)
	if err != nil {
		return nil, fmt.Errorf("unable to prepare spec to json transformation: %w", err)
	}

	// Marshal as json
	jsonData, err := json.Marshal(specBody)
	if err != nil {
		return nil, fmt.Errorf("unable ot marshal spec as JSON: %w", err)
	}

	// No error
	return bytes.NewReader(jsonData), nil
}

// Converts map[interface{}]interface{} into map[string]interface{} for json.Marshaler
func convertMapStringInterface(val interface{}) (interface{}, error) {
	switch items := val.(type) {
	case map[interface{}]interface{}:
		result := map[string]interface{}{}
		for k, v := range items {
			key, ok := k.(string)
			if !ok {
				return nil, fmt.Errorf("TypeError: value %s (type `%s') can't be assigned to type 'string'", k, reflect.TypeOf(k))
			}
			value, err := convertMapStringInterface(v)
			if err != nil {
				return nil, fmt.Errorf("unable to convert map[string] to map[interface{}]: %w", err)
			}
			result[key] = value
		}
		return result, nil
	case []interface{}:
		for k, v := range items {
			value, err := convertMapStringInterface(v)
			if err != nil {
				return nil, fmt.Errorf("unable to convert map[string] to map[interface{}]: %w", err)
			}
			items[k] = value
		}
	}
	return val, nil
}
