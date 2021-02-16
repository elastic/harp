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

package encryption

import (
	"context"
	"reflect"
	"testing"

	"github.com/elastic/harp/pkg/sdk/value"
)

func TestFromKey(t *testing.T) {
	type args struct {
		keyValue string
	}
	tests := []struct {
		name    string
		args    args
		want    value.Transformer
		wantErr bool
	}{
		{
			name: "blank",
			args: args{
				keyValue: "",
			},
			wantErr: true,
		},
		{
			name: "invalid aes-gcm",
			args: args{
				keyValue: "aes-gcm:zQyPnNa-jlQsLW3Ypd87cX88ROMkdgnqv0a3y8",
			},
			wantErr: true,
		},
		{
			name: "invalid secretbox",
			args: args{
				keyValue: "secretbox:gCUODuqhcktiM1USKOfkwVlKhoUyHxXZm6d6",
			},
			wantErr: true,
		},
		{
			name: "invalid fernet",
			args: args{
				keyValue: "fernet:ZER8WwNyw5Dsd65bctxillSrRMX4ObaZsQjaNW1",
			},
			wantErr: true,
		},
		{
			name: "default",
			args: args{
				keyValue: "ZER8WwNyw5Dsd65bctxillSrRMX4ObaZsQjaNW1nBBI=",
			},
			wantErr: false,
		},
		{
			name: "aes-gcm",
			args: args{
				keyValue: "aes-gcm:zQyPnNa-jlQsLW3Ypd87cX88ROMkdgnqv0a3y8LiISg=",
			},
			wantErr: false,
		},
		{
			name: "secretbox",
			args: args{
				keyValue: "secretbox:gCUODuqhcktiM1USKOfkwVlKhoUyHxXZm6d64nztCp0=",
			},
			wantErr: false,
		},
		{
			name: "chacha",
			args: args{
				keyValue: "chacha:gCUODuqhcktiM1USKOfkwVlKhoUyHxXZm6d64nztCp0=",
			},
			wantErr: false,
		},
		{
			name: "fernet",
			args: args{
				keyValue: "fernet:ZER8WwNyw5Dsd65bctxillSrRMX4ObaZsQjaNW1nBBI=",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromKey(tt.args.keyValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				return
			}

			// Encrypt
			msg := []byte("msg")
			encrypted, err := got.To(context.Background(), msg)
			if err != nil {
				t.Errorf("FromKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Decrypt
			decrypted, err := got.From(context.Background(), encrypted)
			if err != nil {
				t.Errorf("FromKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check identity
			if !reflect.DeepEqual(msg, decrypted) {
				t.Errorf("FromKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
