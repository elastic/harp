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
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/go-jose/go-jose/v3"
	"github.com/stretchr/testify/assert"
)

func mustDecodeJWK(input []byte) *jose.JSONWebKey {
	var jwk jose.JSONWebKey
	if err := json.Unmarshal(input, &jwk); err != nil {
		panic(err)
	}

	return &jwk
}

func Test_pasetoTransformer_To(t *testing.T) {
	type fields struct {
		key interface{}
	}
	type args struct {
		ctx   context.Context
		input []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:    "nil",
			wantErr: true,
		},
		{
			name: "nil key",
			fields: fields{
				key: nil,
			},
			wantErr: true,
		},
		{
			name: "invalid key",
			fields: fields{
				key: nil,
			},
			args: args{
				ctx:   context.Background(),
				input: []byte("test"),
			},
			wantErr: true,
		},
		{
			name: "public key",
			fields: fields{
				key: mustDecodeJWK(ed25519PrivateJWK).Public().Key,
			},
			args: args{
				ctx:   context.Background(),
				input: []byte("test"),
			},
			wantErr: true,
		},
		// ---------------------------------------------------------------------
		{
			name: "valid - v4",
			fields: fields{
				key: mustDecodeJWK(ed25519PrivateJWK).Key,
			},
			args: args{
				ctx:   context.Background(),
				input: []byte("test"),
			},
			wantErr: false,
			want:    []byte("v4.public.dGVzdAFYwDsZk_qaPetgB1JoOKeti0f87J2ZUWZlmE1d4TQgrbxYqhNYO7pf8H_5RtpILRUi6E1WJXchECtI1-9-Nwk"),
		},
		{
			name: "valid - v3",
			fields: fields{
				key: mustDecodeJWK(p384PrivateJWK).Key,
			},
			args: args{
				ctx:   context.Background(),
				input: []byte("test"),
			},
			wantErr: false,
			want:    []byte("v3.public.dGVzdMaYCiv2p7mNiMmPpHSyZYTuzsGuudFfsjrZsN2j7FErediyHHqnTJdc4DrpDNpupfSc3Q0GreKbX4JNr_FrhV4UFaLEFw_Z3ZPcs_4I-pn3o9DwvlU9fmWqMd9m5QxFZw"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &pasetoTransformer{
				key: tt.fields.key,
			}
			got, err := d.To(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("pasetoTransformer.To() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("pasetoTransformer.To() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func Test_pasetoTransformer_Roundtrip(t *testing.T) {
	testcases := []struct {
		name               string
		privateKey         *jose.JSONWebKey
		signatureAlgorithm jose.SignatureAlgorithm
	}{
		{
			name:               "ed25519",
			privateKey:         mustDecodeJWK(ed25519PrivateJWK),
			signatureAlgorithm: jose.EdDSA,
		},
		{
			name:               "p384",
			privateKey:         mustDecodeJWK(p384PrivateJWK),
			signatureAlgorithm: jose.ES384,
		},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			signer := &pasetoTransformer{
				key: tt.privateKey.Key,
			}

			verifier := &pasetoTransformer{
				key: tt.privateKey.Public().Key,
			}

			// Prepare context
			ctx := context.Background()
			input := []byte("test")

			signed, err := signer.To(ctx, input)
			assert.NoError(t, err)

			payload, err := verifier.From(ctx, signed)
			assert.NoError(t, err)

			assert.Equal(t, input, payload)
		})
	}
}
