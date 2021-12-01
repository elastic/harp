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
package v1

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/awnumar/memguard"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/assert"
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

func TestSeal(t *testing.T) {
	publicKey1, _, _ := box.GenerateKey(bytes.NewReader([]byte("deterministic-generation-for-tests-0002")))
	publicKey2, _, _ := box.GenerateKey(bytes.NewReader([]byte("deterministic-generation-for-tests-0003")))

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
			name: "empty container with public keys",
			args: args{
				container: &containerv1.Container{
					Headers: &containerv1.Header{},
				},
				peersPublicKey: []interface{}{
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
					Raw:     memguard.NewBufferRandom(1024).Bytes(),
				},
				peersPublicKey: []interface{}{
					publicKey1,
					publicKey2,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := New()
			_, err := adapter.Seal(rand.Reader, tt.args.container, tt.args.peersPublicKey...)
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
		Raw: memguard.NewBufferRandom(1024).Bytes(),
	}

	adapter := New()

	sealed, err := adapter.Seal(rand.Reader, input, nil, publicKey1)
	if err != nil {
		t.Fatalf("unable to seal container: %v", err)
	}

	unsealed, err := adapter.Unseal(sealed, memguard.NewBufferFromBytes(privateKey1[:]))
	if err != nil {
		t.Fatalf("unable to unseal container: %v", err)
	}

	if diff := cmp.Diff(unsealed, input, ignoreOpts...); diff != "" {
		t.Errorf("Seal/Unseal()\n-got/+want\ndiff %s", diff)
	}
}

func Test_Seal_Fuzz(t *testing.T) {
	adapter := New()

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
		adapter.Seal(rand.Reader, &input, &publicKey)
	}
}

func Test_UnSeal_Fuzz(t *testing.T) {
	// Memguard buffer is excluded from fuzz for random race condition error
	// investigation will be done in a separated thread.
	identity := memguard.NewBufferRandom(32)

	adapter := New()

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
		adapter.Unseal(&input, identity)
	}
}

// -----------------------------------------------------------------------------
func benchmarkSeal(container *containerv1.Container, peersPublicKeys []interface{}, b *testing.B) {
	adapter := New()
	for n := 0; n < b.N; n++ {
		_, err := adapter.Seal(rand.Reader, container, peersPublicKeys...)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Seal(b *testing.B) {
	publicKey, _, err := box.GenerateKey(bytes.NewReader([]byte("deterministic-generation-for-tests-0002")))
	assert.NoError(b, err)

	input := &containerv1.Container{
		Headers: &containerv1.Header{
			ContentEncoding: "gzip",
			ContentType:     "application/vnd.harp.v1.Bundle",
		},
		Raw: make([]byte, 1024),
	}

	benchmarkSeal(input, []interface{}{publicKey}, b)
}
