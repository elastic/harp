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
	"fmt"
	"io"
	"io/ioutil"
	"reflect"

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
	in, err := ioutil.ReadAll(r)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("unable to drain input reader: %w", err)
	}

	// Decode as YAML any object
	var specBody interface{}
	if errYaml := yaml.Unmarshal(in, &specBody); errYaml != nil {
		return nil, errYaml
	}

	// Convert map[interface{}]interface{} to a JSON serializable struct
	specBody, err = convertMapStringInterface(specBody)
	if err != nil {
		return nil, err
	}

	// Marshal as json
	jsonData, err := json.Marshal(specBody)
	if err != nil {
		return nil, err
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
				return nil, err
			}
			result[key] = value
		}
		return result, nil
	case []interface{}:
		for k, v := range items {
			value, err := convertMapStringInterface(v)
			if err != nil {
				return nil, err
			}
			items[k] = value
		}
	}
	return val, nil
}
