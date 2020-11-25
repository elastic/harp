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

package crypto

import "testing"

func TestKey(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		want    string
		wantErr bool
	}{
		{
			name:    "nil",
			args:    "",
			wantErr: true,
		},
		{
			name:    "invalid",
			args:    "foo",
			wantErr: true,
		},
		{
			name:    "aes-128",
			args:    "aes:128",
			wantErr: false,
		},
		{
			name:    "aes-256",
			args:    "aes:256",
			wantErr: false,
		},
		{
			name:    "secretbox",
			args:    "secretbox",
			wantErr: false,
		},
		{
			name:    "fernet",
			args:    "fernet",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Key(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Key() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
