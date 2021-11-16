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

package container

import (
	"bytes"
	"encoding/hex"
	"io"
	"io/ioutil"
	"testing"

	"github.com/awnumar/memguard"
	"github.com/davecgh/go-spew/spew"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	fuzz "github.com/google/gofuzz"
	"golang.org/x/crypto/nacl/box"

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

func mustHexDecode(in string) []byte {
	out, err := hex.DecodeString(in)
	if err != nil {
		panic(err)
	}
	return out
}

func TestDump(t *testing.T) {
	type args struct {
		c *containerv1.Container
	}
	tests := []struct {
		name    string
		args    args
		wantW   []byte
		wantErr bool
	}{
		{
			name:    "nil",
			wantW:   nil,
			wantErr: true,
		},
		{
			name: "empty",
			args: args{
				c: &containerv1.Container{},
			},
			wantW:   mustHexDecode("53cb37010002"),
			wantErr: false,
		},
		{
			name: "not empty",
			args: args{
				c: &containerv1.Container{
					Headers: &containerv1.Header{
						ContentEncoding: "gzip",
						ContentType:     "harp.bundle.v1.Bundle",
					},
					Raw: []byte{0x00, 0x00},
				},
			},
			wantW:   mustHexDecode("53cb370100020a1d0a04677a69701215686172702e62756e646c652e76312e42756e646c6512020000"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			if err := Dump(w, tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("Dump() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotW := w.Bytes()
			if diff := cmp.Diff(gotW, tt.wantW, ignoreOpts...); diff != "" {
				t.Errorf("Dump()\n-got/+want\ndiff %s", diff)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    *containerv1.Container
		wantErr bool
	}{
		{
			name:    "nil",
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid magic",
			args: args{
				r: bytes.NewReader(mustHexDecode("FFFFFFFF0001")),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid version",
			args: args{
				r: bytes.NewReader(mustHexDecode("53cb37010002")),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "empty container",
			args: args{
				r: bytes.NewReader(mustHexDecode("53cb37010001")),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid container",
			args: args{
				r: bytes.NewReader(mustHexDecode("53cb370100010a")),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "valid",
			args: args{
				r: bytes.NewReader(mustHexDecode("53cb370100020a1d0a04677a69701215686172702e62756e646c652e76312e42756e646c6512020000")),
			},
			want: &containerv1.Container{
				Headers: &containerv1.Header{
					ContentEncoding: "gzip",
					ContentType:     "harp.bundle.v1.Bundle",
				},
				Raw: []byte{0x00, 0x00},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Load(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want, ignoreOpts...); diff != "" {
				t.Errorf("Load()\n-got/+want\ndiff %s", diff)
			}
		})
	}
}

func TestSeal(t *testing.T) {
	publicKey1, _, _ := box.GenerateKey(bytes.NewReader([]byte("deterministic-generation-for-tests-0002")))
	publicKey2, _, _ := box.GenerateKey(bytes.NewReader([]byte("deterministic-generation-for-tests-0003")))

	type args struct {
		container      *containerv1.Container
		peersPublicKey []*[32]byte
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
			name: "empty container with public keys",
			args: args{
				container: &containerv1.Container{
					Headers: &containerv1.Header{},
				},
				peersPublicKey: []*[32]byte{
					publicKey1,
					publicKey2,
				},
			},
			wantErr: false,
		},
		{
			name: "valid container with public keys",
			args: args{
				container: &containerv1.Container{
					Headers: &containerv1.Header{},
					Raw:     []byte{0x01, 0x02, 0x03, 0x04},
				},
				peersPublicKey: []*[32]byte{
					publicKey1,
					publicKey2,
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
	publicKey1, privateKey1, err := box.GenerateKey(bytes.NewReader([]byte("deterministic-generation-for-tests-0002")))
	if err != nil {
		t.Fatalf("%v", err)
	}

	input := &containerv1.Container{
		Headers: &containerv1.Header{
			ContentEncoding: "gzip",
			ContentType:     "application/vnd.harp.v1.Bundle",
		},
		Raw: []byte{0x00, 0x00},
	}

	sealed, err := Seal(input, nil, publicKey1)
	if err != nil {
		t.Fatalf("unable to seal container: %v", err)
	}

	spew.Dump(sealed)

	unsealed, err := Unseal(sealed, memguard.NewBufferFromBytes(privateKey1[:]))
	if err != nil {
		t.Fatalf("unable to unseal container: %v", err)
	}

	if diff := cmp.Diff(unsealed, input, ignoreOpts...); diff != "" {
		t.Errorf("Seal/Unseal()\n-got/+want\ndiff %s", diff)
	}
}

// -----------------------------------------------------------------------------

func Test_Load_Fuzz(t *testing.T) {
	// Making sure the function never panics
	for i := 0; i < 50000; i++ {
		f := fuzz.New()

		// Prepare arguments
		var (
			raw []byte
		)

		f.Fuzz(&raw)

		// Execute
		Load(bytes.NewReader(raw))
	}
}

func Test_Dump_Fuzz(t *testing.T) {
	// Making sure the function never panics
	for i := 0; i < 50000; i++ {
		f := fuzz.New()

		// Prepare arguments
		var (
			input containerv1.Container
		)

		f.Fuzz(&input)

		// Execute
		Dump(ioutil.Discard, &input)
	}
}

func Test_Seal_Fuzz(t *testing.T) {
	// Making sure the function never panics
	for i := 0; i < 500; i++ {
		f := fuzz.New()

		// Prepare arguments
		var (
			publicKey [32]byte
		)
		input := containerv1.Container{
			Headers: &containerv1.Header{},
			Raw:     []byte{0x00, 0x00},
		}

		f.Fuzz(&input.Headers)
		f.Fuzz(&input.Raw)
		f.Fuzz(&publicKey)

		// Execute
		Seal(&input, &publicKey)
	}
}

func Test_UnSeal_Fuzz(t *testing.T) {
	// Memguard buffer is excluded from fuzz for random race condition error
	// investigation will be done in a separated thread.
	identity := memguard.NewBufferRandom(32)

	// Making sure the function never panics
	for i := 0; i < 500; i++ {
		f := fuzz.New()

		// Prepare arguments
		input := containerv1.Container{
			Headers: &containerv1.Header{},
			Raw:     []byte{0x00, 0x00},
		}

		f.Fuzz(&input.Headers)
		f.Fuzz(&input.Raw)

		// Execute
		Unseal(&input, identity)
	}
}
