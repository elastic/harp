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

import (
	"testing"
)

func TestKeypair(t *testing.T) {
	type testCase struct {
		name    string
		args    string
		want    interface{}
		wantErr bool
	}
	tests := []testCase{
		{
			name:    "nil",
			args:    "",
			wantErr: true,
		},
		{
			name:    "invalid",
			args:    "azer",
			wantErr: true,
		},
	}
	expectedKeyTypes := []string{"rsa", "rsa:normal", "rsa:2048", "rsa:strong", "rsa:4096", "ec", "ec:normal", "ec:p256", "ec:high", "ec:p384", "ec:strong", "ec:p521", "ssh", "ed25519", "naclbox"}
	for _, kt := range expectedKeyTypes {
		tests = append(tests, testCase{
			name:    kt,
			args:    kt,
			wantErr: false,
		})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Keypair(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Keypair() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got == nil {
				t.Errorf("Keypair() = %v", got)
			}
		})
	}
}
