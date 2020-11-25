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

package security

import (
	"testing"
)

func TestSecureCompare(t *testing.T) {
	type args struct {
		given  []byte
		actual []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "not equal, same size",
			args: args{
				given:  []byte{0x01},
				actual: []byte{0x02},
			},
			want: false,
		},
		{
			name: "not equal, different size",
			args: args{
				given:  []byte{0x01, 0x02},
				actual: []byte{0x02},
			},
			want: false,
		},
		{
			name: "equal, different size",
			args: args{
				given:  []byte{0x00},
				actual: []byte{},
			},
			want: false,
		},
		{
			name: "equal, same size",
			args: args{
				given:  []byte{0x01},
				actual: []byte{0x01},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SecureCompare(tt.args.given, tt.args.actual); got != tt.want {
				t.Errorf("SecureCompare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSecureCompareString(t *testing.T) {
	type args struct {
		given  string
		actual string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "not equal, same size",
			args: args{
				given:  "a",
				actual: "b",
			},
			want: false,
		},
		{
			name: "not equal, different size",
			args: args{
				given:  "ab",
				actual: "a",
			},
			want: false,
		},
		{
			name: "equal, different size",
			args: args{
				given:  "\x00",
				actual: "",
			},
			want: false,
		},
		{
			name: "equal, same size",
			args: args{
				given:  "a",
				actual: "a",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SecureCompareString(tt.args.given, tt.args.actual); got != tt.want {
				t.Errorf("SecureCompareString() = %v, want %v", got, tt.want)
			}
		})
	}
}
