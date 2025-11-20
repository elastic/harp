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

package operation

import (
	"errors"
	"testing"
)

func Test_ClassifyVaultError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "nil error",
			err:  nil,
			want: "unknown",
		},
		{
			name: "connection error",
			err:  errors.New("connection refused"),
			want: "connection",
		},
		{
			name: "connection error with context",
			err:  errors.New("vault: connection timeout after 30s"),
			want: "connection",
		},
		{
			name: "permission error",
			err:  errors.New("permission denied"),
			want: "permission",
		},
		{
			name: "permission error with vault prefix",
			err:  errors.New("vault: permission denied for path secret/data/foo"),
			want: "permission",
		},
		{
			name: "access denied error",
			err:  errors.New("access denied"),
			want: "permission",
		},
		{
			name: "denied error",
			err:  errors.New("request denied by policy"),
			want: "permission",
		},
		{
			name: "timeout error",
			err:  errors.New("operation timeout"),
			want: "timeout",
		},
		{
			name: "timeout error with context",
			err:  errors.New("vault: timeout waiting for response"),
			want: "timeout",
		},
		{
			name: "unknown error",
			err:  errors.New("some random error"),
			want: "unknown",
		},
		{
			name: "not found error",
			err:  errors.New("path not found"),
			want: "unknown",
		},
		{
			name: "empty error message",
			err:  errors.New(""),
			want: "unknown",
		},
		{
			name: "mixed case connection error",
			err:  errors.New("Connection Failed"),
			want: "connection",
		},
		{
			name: "mixed case permission error",
			err:  errors.New("Permission Denied"),
			want: "permission",
		},
		{
			name: "mixed case timeout error",
			err:  errors.New("Timeout Exceeded"),
			want: "timeout",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyVaultError(tt.err)
			if got != tt.want {
				t.Errorf("ClassifyVaultError() = %v, want %v", got, tt.want)
			}
		})
	}
}
