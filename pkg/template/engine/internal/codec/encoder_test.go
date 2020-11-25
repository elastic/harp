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
	"testing"

	fuzz "github.com/google/gofuzz"
)

func TestToYAML(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty",
			args: args{
				v: map[string]string{},
			},
			want: "{}",
		},
		{
			name: "object",
			args: args{
				v: map[string]string{
					"key": "value",
				},
			},
			want: "key: value",
		},
		{
			name: "non-serializable",
			args: args{
				v: map[string]interface{}{
					"key": make(chan string, 1),
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToYAML(tt.args.v); got != tt.want {
				t.Errorf("ToYAML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToTOML(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty",
			args: args{
				v: map[string]string{},
			},
			want: "",
		},
		{
			name: "object",
			args: args{
				v: map[string]string{
					"key": "value",
				},
			},
			want: "key = \"value\"\n",
		},
		/*{
			name: "non-serializable",
			args: args{
				v: map[string]interface{}{
					"key": make(chan string, 1),
				},
			},
			want: "",
		},*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToTOML(tt.args.v); got != tt.want {
				t.Errorf("ToTOML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToJSON(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty",
			args: args{
				v: map[string]string{},
			},
			want: "{}",
		},
		{
			name: "object",
			args: args{
				v: map[string]string{
					"key": "value",
				},
			},
			want: `{"key":"value"}`,
		},
		{
			name: "non-serializable",
			args: args{
				v: map[string]interface{}{
					"key": make(chan string, 1),
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToJSON(tt.args.v); got != tt.want {
				t.Errorf("ToJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

// -----------------------------------------------------------------------------

func TestToYAML_Fuzz(t *testing.T) {
	// Making sure that it never panics
	for i := 0; i < 50; i++ {
		f := fuzz.New()

		// Prepare arguments
		var input struct {
			Integer int
			String  string
			Map     map[string]string
		}

		// Fuzz input
		f.Fuzz(&input)

		// Execute
		ToYAML(input)
	}
}

func TestToTOML_Fuzz(t *testing.T) {
	// Making sure that it never panics
	for i := 0; i < 50; i++ {
		f := fuzz.New()

		// Prepare arguments
		var input struct {
			Integer int
			String  string
			Map     map[string]string
		}

		// Fuzz input
		f.Fuzz(&input)

		// Execute
		ToTOML(input)
	}
}

func TestToJSON_Fuzz(t *testing.T) {
	// Making sure that it never panics
	for i := 0; i < 50; i++ {
		f := fuzz.New()

		// Prepare arguments
		var input struct {
			Integer int
			String  string
			Map     map[string]string
		}

		// Fuzz input
		f.Fuzz(&input)

		// Execute
		ToJSON(input)
	}
}
