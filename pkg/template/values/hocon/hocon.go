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

package hocon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-akka/configuration"
	"github.com/go-akka/configuration/hocon"
	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/sdk/log"
)

// Parser is a HOCON parser
type Parser struct{}

// Unmarshal unmarshals HOCON files
func (i *Parser) Unmarshal(p []byte, v interface{}) error {
	// Parse HOCON configuration
	rootCfg := configuration.ParseString(string(p), hoconIncludeCallback).Root()

	// Visit config tree
	res := visitNode(rootCfg)

	// Encode as json
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(res); err != nil {
		return fmt.Errorf("unable to encode HOCON map to JSON: %w", err)
	}

	// Decode JSON
	if err := json.Unmarshal(buf.Bytes(), v); err != nil {
		return fmt.Errorf("unable to decode json object as struct: %w", err)
	}

	return nil
}

// -----------------------------------------------------------------------------

func visitNode(node *hocon.HoconValue) interface{} {
	if node.IsArray() {
		nodes := node.GetArray()

		res := make([]interface{}, len(nodes))
		for i, n := range nodes {
			res[i] = visitNode(n)
		}

		return res
	}

	if node.IsObject() {
		obj := node.GetObject()

		res := map[string]interface{}{}
		keys := obj.GetKeys()
		for _, k := range keys {
			res[k] = visitNode(obj.GetKey(k))
		}

		return res
	}

	if node.IsString() {
		return node.GetString()
	}

	if node.IsEmpty() {
		return nil
	}

	return nil
}

func hoconIncludeCallback(filename string) *hocon.HoconRoot {
	files, err := filepath.Glob(filename)
	switch {
	case err != nil:
		log.Bg().Error("hocon: unable to load file glob", zap.Error(err), zap.String("filename", filename))
		return nil
	case len(files) == 0:
		log.Bg().Warn("hocon: unable to load file %s", zap.String("filename", filename))
		return hocon.Parse("", nil)
	default:
		root := hocon.Parse("", nil)
		for _, f := range files {
			//nolint:gosec // G304: f is from glob pattern matching, not direct user input
			data, err := os.ReadFile(f)
			if err != nil {
				log.Bg().Error("hocon: unable to load file glob", zap.Error(err))
				return nil
			}

			node := hocon.Parse(string(data), hoconIncludeCallback)
			if node != nil {
				root.Value().GetObject().Merge(node.Value().GetObject())
				// merge substitutions
				subs := make([]*hocon.HoconSubstitution, 0)
				subs = append(subs, root.Substitutions()...)
				subs = append(subs, node.Substitutions()...)
				root = hocon.NewHoconRoot(root.Value(), subs...)
			}
		}
		return root
	}
}
