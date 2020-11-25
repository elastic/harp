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

package codec

import (
	"reflect"
	"testing"

	fuzz "github.com/google/gofuzz"
)

func TestFromYAML(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			name: "empty",
			args: args{
				str: "",
			},
			want: map[string]interface{}{},
		},
		{
			name: "with error",
			args: args{
				str: ";",
			},
			want: map[string]interface{}{"Error": "error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type map[string]interface {}"},
		},
		{
			name: "valid",
			args: args{
				str: "key: value",
			},
			want: map[string]interface{}{"key": "value"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromYAML(tt.args.str); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromYAML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromYAMLArray(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{
			name: "empty",
			args: args{
				str: "",
			},
			want: []interface{}{},
		},
		{
			name: "with error",
			args: args{
				str: ";",
			},
			want: []interface{}{"error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type []interface {}"},
		},
		{
			name: "valid",
			args: args{
				str: "['1','2']",
			},
			want: []interface{}{"1", "2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromYAMLArray(tt.args.str); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromYAMLArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromJSON(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			name: "empty",
			args: args{
				str: "",
			},
			want: map[string]interface{}{"Error": "unexpected end of JSON input"},
		},
		{
			name: "with error",
			args: args{
				str: ";",
			},
			want: map[string]interface{}{"Error": "invalid character ';' looking for beginning of value"},
		},
		{
			name: "valid",
			args: args{
				str: `{"key": "value"}`,
			},
			want: map[string]interface{}{"key": "value"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromJSON(tt.args.str); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromJSONArray(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{
			name: "empty",
			args: args{
				str: "",
			},
			want: []interface{}{"unexpected end of JSON input"},
		},
		{
			name: "with error",
			args: args{
				str: ";",
			},
			want: []interface{}{"invalid character ';' looking for beginning of value"},
		},
		{
			name: "valid",
			args: args{
				str: `["1","2"]`,
			},
			want: []interface{}{"1", "2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromJSONArray(tt.args.str); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromJSONArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

// -----------------------------------------------------------------------------

func TestFromYAML_Fuzz(t *testing.T) {
	// Making sure that it never panics
	for i := 0; i < 50; i++ {
		f := fuzz.New()

		// Prepare arguments
		var input string

		// Fuzz input
		f.Fuzz(&input)

		// Execute
		FromYAML(input)
	}
}

func TestFromYAMLArray_Fuzz(t *testing.T) {
	// Making sure that it never panics
	for i := 0; i < 50; i++ {
		f := fuzz.New()

		// Prepare arguments
		var input string

		// Fuzz input
		f.Fuzz(&input)

		// Execute
		FromYAMLArray(input)
	}
}

func TestFromJSON_Fuzz(t *testing.T) {
	// Making sure that it never panics
	for i := 0; i < 50; i++ {
		f := fuzz.New()

		// Prepare arguments
		var input string

		// Fuzz input
		f.Fuzz(&input)

		// Execute
		FromJSON(input)
	}
}

func TestFromJSONArray_Fuzz(t *testing.T) {
	// Making sure that it never panics
	for i := 0; i < 50; i++ {
		f := fuzz.New()

		// Prepare arguments
		var input string

		// Fuzz input
		f.Fuzz(&input)

		// Execute
		FromJSONArray(input)
	}
}
