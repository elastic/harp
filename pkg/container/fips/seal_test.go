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
package fips

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/awnumar/memguard"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	containerv1 "github.com/elastic/harp/api/gen/go/harp/container/v1"
)

var (
	opt = cmp.FilterPath(
		func(p cmp.Path) bool {
			// Remove ignoring of the fields below once go-cmp is able to ignore generated fields.
			// See https://github.com/google/go-cmp/issues/153
			ignoreXXXCache :=
				p.String() == "XXX_sizecache" ||
					p.String() == "Headers.XXX_sizecache"
			return ignoreXXXCache
		}, cmp.Ignore())

	ignoreOpts = []cmp.Option{
		cmpopts.IgnoreUnexported(containerv1.Container{}),
		cmpopts.IgnoreUnexported(containerv1.Header{}),
		opt,
	}
)

// -----------------------------------------------------------------------------

func TestSeal(t *testing.T) {
	privKey, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	pubKey := privKey.PublicKey

	type args struct {
		container      *containerv1.Container
		peersPublicKey []interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    *containerv1.Container
		wantErr bool
	}{
		{
			name:    "nil",
			wantErr: true,
		},
		{
			name: "empty container",
			args: args{
				container: &containerv1.Container{},
			},
			wantErr: true,
		},
		{
			name: "empty container headers",
			args: args{
				container: &containerv1.Container{
					Headers: &containerv1.Header{},
				},
			},
			wantErr: true,
		},
		{
			name: "nil public keys",
			args: args{
				container: &containerv1.Container{
					Headers: &containerv1.Header{},
				},
				peersPublicKey: []interface{}{
					nil,
				},
			},
			wantErr: true,
		},
		// ---------------------------------------------------------------------
		{
			name: "valid",
			args: args{
				container: &containerv1.Container{
					Headers: &containerv1.Header{},
					Raw:     []byte{0x01, 0x02, 0x03, 0x04},
				},
				peersPublicKey: []interface{}{
					&pubKey,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Seal(tt.args.container, tt.args.peersPublicKey...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Seal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

// -----------------------------------------------------------------------------

func Test_Seal_Unseal(t *testing.T) {
	privKey, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	pubKey := privKey.PublicKey

	input := &containerv1.Container{
		Headers: &containerv1.Header{
			ContentEncoding: "gzip",
			ContentType:     "application/vnd.harp.v1.Bundle",
		},
		Raw: []byte{0x00, 0x00},
	}

	sealed, err := Seal(input, &pubKey)
	if err != nil {
		t.Fatalf("unable to seal container: %v", err)
	}

	unsealed, err := Unseal(sealed, memguard.NewBufferFromBytes(privKey.D.Bytes()))
	if err != nil {
		t.Fatalf("unable to unseal container: %v", err)
	}

	if diff := cmp.Diff(unsealed, input, ignoreOpts...); diff != "" {
		t.Errorf("Seal/Unseal()\n-got/+want\ndiff %s", diff)
	}
}
