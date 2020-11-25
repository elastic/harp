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

package x25519

import (
	"encoding/base64"
	"testing"
)

func mustDecode(in string) []byte {
	out, err := base64.RawURLEncoding.DecodeString(in)
	if err != nil {
		panic(err)
	}

	return out
}

func TestIsValidPublicKey(t *testing.T) {
	type args struct {
		key []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nil",
			args: args{
				key: nil,
			},
			want: false,
		},
		{
			name: "blank",
			args: args{
				key: []byte{},
			},
			want: false,
		},
		{
			name: "invalid length",
			args: args{
				key: mustDecode("xS2R26nLLI2vdaPtY1g"),
			},
			want: false,
		},
		{
			name: "valid",
			args: args{
				key: mustDecode("xS2R26nLLI2vdaPtY1gUZSvgazVEHJJfZ7QKzjBGESk"),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidPublicKey(tt.args.key); got != tt.want {
				t.Errorf("IsValidPublicKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
