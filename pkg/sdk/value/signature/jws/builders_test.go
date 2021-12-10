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

package jws

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/elastic/harp/pkg/sdk/value"
	"github.com/stretchr/testify/assert"
)

func TestTransformer(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		want    value.Transformer
		wantErr bool
	}{
		{
			name:    "nil",
			wantErr: true,
		},
		{
			name: "invalid base64",
			args: args{
				key: "jws:123456789%",
			},
			wantErr: true,
		},
		// ---------------------------------------------------------------------
		{
			name: "valid",
			args: args{
				key: fmt.Sprintf("jws:%s", base64.RawURLEncoding.EncodeToString(rsa4096PrivateJWK)),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Transformer(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Transformer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.NotNil(t, got)
			}
		})
	}
}
