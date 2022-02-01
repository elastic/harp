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

package encryption_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/elastic/harp/pkg/sdk/value"
	"github.com/elastic/harp/pkg/sdk/value/encryption"

	// Register encryption transformers
	_ "github.com/elastic/harp/pkg/sdk/value/encryption/aead"
	_ "github.com/elastic/harp/pkg/sdk/value/encryption/age"
	_ "github.com/elastic/harp/pkg/sdk/value/encryption/dae"
	_ "github.com/elastic/harp/pkg/sdk/value/encryption/fernet"
	_ "github.com/elastic/harp/pkg/sdk/value/encryption/jwe"
	_ "github.com/elastic/harp/pkg/sdk/value/encryption/paseto"
	_ "github.com/elastic/harp/pkg/sdk/value/encryption/secretbox"
	"github.com/elastic/harp/pkg/sdk/value/mock"
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
			name: "aes-gcm",
			args: args{
				keyValue: "aes-gcm:zQyPnNa-jlQsLW3Ypd87cX88ROMkdgnqv0a3y8LiISg=",
			},
			wantErr: false,
		},
		{
			name: "dae-aes-gcm",
			args: args{
				keyValue: "dae-aes-gcm:zQyPnNa-jlQsLW3Ypd87cX88ROMkdgnqv0a3y8LiISg=",
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
			name: "dae-chacha",
			args: args{
				keyValue: "dae-chacha:gCUODuqhcktiM1USKOfkwVlKhoUyHxXZm6d64nztCp0=",
			},
			wantErr: false,
		},
		{
			name: "xchacha",
			args: args{
				keyValue: "xchacha:VhfCXaD_QwwwoPCjLJx6vgnaSo0sMPjdCmT0RUUQjBQ=",
			},
			wantErr: false,
		},
		{
			name: "dae-xchacha",
			args: args{
				keyValue: "dae-xchacha:VhfCXaD_QwwwoPCjLJx6vgnaSo0sMPjdCmT0RUUQjBQ=",
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
		{
			name: "aes-siv",
			args: args{
				keyValue: "aes-siv:2XEKpPbE8T0ghLj8Wr9v6stV0YrUCNSoSbtc69Kh-n7-pVaKmWZ8LSvaJOK9BJHqDWE8vyNSzyNpcTYv3-J9lw==",
			},
			wantErr: false,
		},
		{
			name: "dae-aes-siv",
			args: args{
				keyValue: "dae-aes-siv:2XEKpPbE8T0ghLj8Wr9v6stV0YrUCNSoSbtc69Kh-n7-pVaKmWZ8LSvaJOK9BJHqDWE8vyNSzyNpcTYv3-J9lw==",
			},
			wantErr: false,
		},
		{
			name: "aes-pmac-siv",
			args: args{
				keyValue: "aes-pmac-siv:Brfled4G7okhpCb6T2HMWKgDo1vyqrEdWWVIXfcFUysHaOacXkER5z9GHRuz89scK2TSE962nAFUcScAkihP9w==",
			},
			wantErr: false,
		},
		{
			name: "dae-aes-pmac-siv",
			args: args{
				keyValue: "dae-aes-pmac-siv:Brfled4G7okhpCb6T2HMWKgDo1vyqrEdWWVIXfcFUysHaOacXkER5z9GHRuz89scK2TSE962nAFUcScAkihP9w==",
			},
			wantErr: false,
		},
		{
			name: "jwe",
			args: args{
				keyValue: "jwe:a256kw:ZER8WwNyw5Dsd65bctxillSrRMX4ObaZsQjaNW1nBBI=",
			},
			wantErr: false,
		},
		{
			name: "paseto",
			args: args{
				keyValue: "paseto:kP1yHnBcOhjowNFXSCyycSuXdUqTlbuE6ES5tTp-I_o=",
			},
			wantErr: false,
		},
		/*{
			name: "age-recipients",
			args: args{
				keyValue: "age-recipients:age1ce20pmz8z0ue97v7rz838v6pcpvzqan30lr40tjlzy40ez8eldrqf2zuxe",
			},
			wantErr: false,
		},
		{
			name: "age-identity",
			args: args{
				keyValue: "age-identity:AGE-SECRET-KEY-1W8E69DQEVASNK68FX7C6QLD99KTG96RHWW0EZ3RD0L29AHV4S84QHUAP4C",
			},
			wantErr: false,
		},*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := encryption.FromKey(tt.args.keyValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				return
			}

			// Ensure not panic
			assert.NotPanics(t, func() {
				encryption.Must(got, err)
			})

			// Encrypt
			msg := []byte("msg")
			encrypted, err := got.To(context.Background(), msg)
			if err != nil {
				t.Errorf("To() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Decrypt
			decrypted, err := got.From(context.Background(), encrypted)
			if err != nil {
				t.Errorf("From() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check identity
			if !reflect.DeepEqual(msg, decrypted) {
				t.Errorf("expectd: %v, got: %v", msg, decrypted)
				return
			}
		})
	}
}

func TestMust(t *testing.T) {
	assert.Panics(t, func() {
		encryption.Must(mock.Transformer(nil), errors.New("test"))
	})

	assert.Panics(t, func() {
		encryption.Must(nil, nil)
	})
}
