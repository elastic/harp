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

package raw

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/harp/pkg/sdk/value/signature"
	"github.com/stretchr/testify/assert"
	"gopkg.in/square/go-jose.v2"
)

// -----------------------------------------------------------------------------

var p256PrivateJWK = []byte(`{
    "kty": "EC",
    "d": "sXCIy5HxtyG24MTl3hsgLDqi0dd33WAB_Rae1I_o2Is",
    "crv": "P-256",
    "x": "ykS0SN-EaFIVQUBC7norE9yYAN0ZFxSYYP6p0iofMxw",
    "y": "faQhXipqrhZeHIPFzJEYlxVvCdezZnJs2mKxnraO8_M"
}`)

var p384PrivateJWK = []byte(`{
    "kty": "EC",
    "d": "7YcsmkNxmZdzGyb46ZeDb2I1yr-ja1iw9gspGjq7UDqQ6a61h_ES8c4uU__adkFV",
    "crv": "P-384",
    "x": "dWLSo6PTkL1G68bzTwY3zzrL_QX-pwvP9HUPpQGeSFmj20EWOtfvXXKDrCR0jnJD",
    "y": "lFvTFechH_KmbOEvycryCHy23Cm1qekJYAtn7T_TELpm7zsY290NYlvqDKesGeXx"
}`)

var p521PrivateJWK = []byte(`{
    "kty": "EC",
    "d": "AIqIPpDjCGGwdG1usjkOkzovnv0SMiMgfLTn938E_gp4NBEyQVy4myOilDAEKrxPWw8f1u3FLKhGza-yxevMnfnr",
    "crv": "P-521",
    "x": "AVfi6aKylpZU334mETb2lNO5Ckpzp_L06WG4UQpiFxQMdxxKeldRJTxgt3FCYg5rXbUcKB2vm7Yq1Mxl3CHeBGQ8",
    "y": "AQQurRdp6oLjLbOTosM2cnu91dBL2YShDnqXbaUyFlGYoUJB6LPwwph9Uu0qHKCeK6QxZmHWxST2iky7ObEfM8GC"
}`)

var ed25519PrivateJWK = []byte(`{
    "kty": "OKP",
    "d": "ytOw6kKTTVJUKCnX5HgmhsGguNFQ18ECIS2C-ujJv-s",
    "crv": "Ed25519",
    "x": "K5i0d37-eRk8-EPwo2bpcmM-HGmzLiqRtWnk7oR3FCs"
}`)

func mustDecodeJWK(input []byte) *jose.JSONWebKey {
	var jwk jose.JSONWebKey
	if err := json.Unmarshal(input, &jwk); err != nil {
		panic(err)
	}

	return &jwk
}

// -----------------------------------------------------------------------------
func Test_rawTransformer_Roundtrip(t *testing.T) {
	testcases := []struct {
		name       string
		privateKey *jose.JSONWebKey
	}{
		{
			name:       "ed25519",
			privateKey: mustDecodeJWK(ed25519PrivateJWK),
		},
		{
			name:       "p256",
			privateKey: mustDecodeJWK(p256PrivateJWK),
		},
		{
			name:       "p384",
			privateKey: mustDecodeJWK(p384PrivateJWK),
		},
		{
			name:       "p521",
			privateKey: mustDecodeJWK(p521PrivateJWK),
		},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			signer := &rawTransformer{
				key: tt.privateKey.Key,
			}

			verifier := &rawTransformer{
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

func Test_rawTransformer_PreHashed(t *testing.T) {
	testcases := []struct {
		name       string
		privateKey *jose.JSONWebKey
		input      []byte
	}{
		{
			name:       "p256",
			privateKey: mustDecodeJWK(p256PrivateJWK),
			input:      []byte("00000000000000000000000000000000"),
		},
		{
			name:       "p384",
			privateKey: mustDecodeJWK(p384PrivateJWK),
			input:      []byte("000000000000000000000000000000000000000000000000"),
		},
		{
			name:       "p521",
			privateKey: mustDecodeJWK(p521PrivateJWK),
			input:      []byte("0000000000000000000000000000000000000000000000000000000000000000"),
		},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			signer := &rawTransformer{
				key: tt.privateKey.Key,
			}

			verifier := &rawTransformer{
				key: tt.privateKey.Public().Key,
			}

			// Prepare context
			ctx := signature.WithInputPreHashed(context.Background(), true)

			signed, err := signer.To(ctx, tt.input)
			assert.NoError(t, err)

			payload, err := verifier.From(ctx, signed)
			assert.NoError(t, err)

			assert.Equal(t, tt.input, payload)
		})
	}
}

func Test_rawTransformer_Roundtrip_WithDeterministic(t *testing.T) {
	testcases := []struct {
		name       string
		privateKey *jose.JSONWebKey
	}{
		{
			name:       "p256",
			privateKey: mustDecodeJWK(p256PrivateJWK),
		},
		{
			name:       "p384",
			privateKey: mustDecodeJWK(p384PrivateJWK),
		},
		{
			name:       "p521",
			privateKey: mustDecodeJWK(p521PrivateJWK),
		},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			signer := &rawTransformer{
				key: tt.privateKey.Key,
			}

			verifier := &rawTransformer{
				key: tt.privateKey.Public().Key,
			}

			// Prepare context
			ctx := signature.WithDetermisticSignature(context.Background(), true)
			input := []byte("test")

			signed, err := signer.To(ctx, input)
			assert.NoError(t, err)

			payload, err := verifier.From(ctx, signed)
			assert.NoError(t, err)

			assert.Equal(t, input, payload)
		})
	}
}
