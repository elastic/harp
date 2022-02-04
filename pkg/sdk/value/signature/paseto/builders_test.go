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

package paseto

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/elastic/harp/pkg/sdk/value"
)

var p384PrivateJWK = []byte(`{
    "kty": "EC",
    "d": "7YcsmkNxmZdzGyb46ZeDb2I1yr-ja1iw9gspGjq7UDqQ6a61h_ES8c4uU__adkFV",
    "crv": "P-384",
    "x": "dWLSo6PTkL1G68bzTwY3zzrL_QX-pwvP9HUPpQGeSFmj20EWOtfvXXKDrCR0jnJD",
    "y": "lFvTFechH_KmbOEvycryCHy23Cm1qekJYAtn7T_TELpm7zsY290NYlvqDKesGeXx"
}`)

var ed25519PrivateJWK = []byte(`{
    "kty": "OKP",
    "d": "ytOw6kKTTVJUKCnX5HgmhsGguNFQ18ECIS2C-ujJv-s",
    "crv": "Ed25519",
    "x": "K5i0d37-eRk8-EPwo2bpcmM-HGmzLiqRtWnk7oR3FCs"
}`)

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
				key: "paseto:123456789%",
			},
			wantErr: true,
		},
		// ---------------------------------------------------------------------
		{
			name: "valid - v3",
			args: args{
				key: fmt.Sprintf("paseto:%s", base64.RawURLEncoding.EncodeToString(p384PrivateJWK)),
			},
			wantErr: false,
		},
		{
			name: "valid - v4",
			args: args{
				key: fmt.Sprintf("paseto:%s", base64.RawURLEncoding.EncodeToString(ed25519PrivateJWK)),
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
