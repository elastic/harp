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

package engine

import (
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestFuncs(t *testing.T) {
	// TODO write tests for failure cases
	tests := []struct {
		tpl, expect string
		vars        interface{}
	}{{
		tpl:    `{{ toYaml . }}`,
		expect: `foo: bar`,
		vars:   map[string]interface{}{"foo": "bar"},
	}, {
		tpl:    `{{ toToml . }}`,
		expect: "foo = \"bar\"\n",
		vars:   map[string]interface{}{"foo": "bar"},
	}, {
		tpl:    `{{ toJson . }}`,
		expect: `{"foo":"bar"}`,
		vars:   map[string]interface{}{"foo": "bar"},
	}, {
		tpl:    `{{ fromYaml . }}`,
		expect: "map[hello:world]",
		vars:   `hello: world`,
	}, {
		tpl:    `{{ fromYamlArray . }}`,
		expect: "[one 2 map[name:helm]]",
		vars:   "- one\n- 2\n- name: helm\n",
	}, {
		tpl:    `{{ fromYamlArray . }}`,
		expect: "[one 2 map[name:helm]]",
		vars:   `["one", 2, { "name": "helm" }]`,
	}, {
		tpl:    `{{ toToml . }}`,
		expect: "\n[mast]\n  sail = \"white\"\n",
		vars:   map[string]map[string]string{"mast": {"sail": "white"}},
	}, {
		tpl:    `{{ fromYaml . }}`,
		expect: "map[Error:error unmarshaling JSON: while decoding JSON: json: cannot unmarshal array into Go value of type map[string]interface {}]",
		vars:   "- one\n- two\n",
	}, {
		tpl:    `{{ fromJson .}}`,
		expect: `map[hello:world]`,
		vars:   `{"hello":"world"}`,
	}, {
		tpl:    `{{ fromJson . }}`,
		expect: `map[Error:json: cannot unmarshal array into Go value of type map[string]interface {}]`,
		vars:   `["one", "two"]`,
	}, {
		tpl:    `{{ fromJsonArray . }}`,
		expect: `[one 2 map[name:helm]]`,
		vars:   `["one", 2, { "name": "helm" }]`,
	}, {
		tpl:    `{{ fromJsonArray . }}`,
		expect: `[json: cannot unmarshal object into Go value of type []interface {}]`,
		vars:   `{"hello": "world"}`,
	}, {
		tpl:    `{{ merge .dict (fromYaml .yaml) }}`,
		expect: `map[a:map[b:c]]`,
		vars:   map[string]interface{}{"dict": map[string]interface{}{"a": map[string]interface{}{"b": "c"}}, "yaml": `{"a":{"b":"d"}}`},
	}, {
		tpl:    `{{ merge (fromYaml .yaml) .dict }}`,
		expect: `map[a:map[b:d]]`,
		vars:   map[string]interface{}{"dict": map[string]interface{}{"a": map[string]interface{}{"b": "c"}}, "yaml": `{"a":{"b":"d"}}`},
	}, {
		tpl:    `{{ fromYaml . }}`,
		expect: `map[Error:error unmarshaling JSON: while decoding JSON: json: cannot unmarshal array into Go value of type map[string]interface {}]`,
		vars:   `["one", "two"]`,
	}, {
		tpl:    `{{ fromYamlArray . }}`,
		expect: `[error unmarshaling JSON: while decoding JSON: json: cannot unmarshal object into Go value of type []interface {}]`,
		vars:   `hello: world`,
	}, {
		tpl:    `{{ jsonEscape . }}`,
		expect: `backslash: \\, A: \u0026 \u003c`,
		vars:   `backslash: \, A: & <`,
	}, {
		tpl:    `{{ jsonUnescape . }}`,
		expect: `backslash: \, A: & <`,
		vars:   `"backslash: \\, A: \u0026 \u003c"`,
	}, {
		tpl:    `{{ unquote . }}`,
		expect: `{"channel":"buu","name":"john", "msg":"doe"}`,
		vars:   `"{\"channel\":\"buu\",\"name\":\"john\", \"msg\":\"doe\"}"`,
	}}

	for _, tt := range tests {
		var b strings.Builder
		err := template.Must(template.New("test").Funcs(FuncMap(nil)).Parse(tt.tpl)).Execute(&b, tt.vars)
		assert.NoError(t, err)
		assert.Equal(t, tt.expect, b.String(), tt.tpl)
	}
}
