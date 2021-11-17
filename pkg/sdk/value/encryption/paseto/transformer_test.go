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
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
)

func Test_Transformer_InvalidKey(t *testing.T) {
	keys := []string{
		"",
		"foo",
		"123456",
	}
	for _, k := range keys {
		key := k
		t.Run(fmt.Sprintf("key `%s`", key), func(t *testing.T) {
			underTest, err := Transformer(key)
			if err == nil {
				t.Fatalf("Transformer should raise an error with key `%s`", key)
			}
			if underTest != nil {
				t.Fatalf("Transformer instance should be nil")
			}
		})
	}
}

func Test_Transformer_From(t *testing.T) {
	// Prepare testcases
	testCases := []struct {
		name    string
		input   []byte
		wantErr bool
		want    []byte
	}{
		{
			name:    "Invalid encrypted payload",
			input:   []byte("bad-encryption-payload"),
			wantErr: true,
		},
		{
			name:    "Valid payload",
			input:   []byte("v4.local.tMe_MuiltiVR4NlnbtCiXP7w3v2rkE3iLpOJG4Gyfxc3UTHIbHzKIrRu0e8Mb_Q93kXTm99GU5AjquJalAG8qTp7fxs"),
			wantErr: false,
			want:    []byte("test"),
		},
	}

	// For each testcase
	for _, tc := range testCases {
		testCase := tc
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Initialize mock
			ctx := context.Background()

			// Initialize transformer
			underTest, err := Transformer("kP1yHnBcOhjowNFXSCyycSuXdUqTlbuE6ES5tTp-I_o=")
			if err != nil {
				t.Fatalf("unable to initialize transformer: %v", err)
			}

			// Do the call
			got, err := underTest.From(ctx, testCase.input)

			// Assert results expectations
			if (err != nil) != testCase.wantErr {
				t.Errorf("error during the call, error = %v, wantErr %v", err, testCase.wantErr)
				return
			}
			if testCase.wantErr {
				return
			}
			if diff := cmp.Diff(got, testCase.want); diff != "" {
				t.Errorf("%q. Paseto.From():\n-got/+want\ndiff %s", testCase.name, diff)
			}
		})
	}
}

func Test_Transformer_To(t *testing.T) {
	// Prepare testcases
	testCases := []struct {
		name    string
		input   []byte
		wantErr bool
		want    []byte
	}{
		{
			name:    "Valid payload",
			input:   []byte("test"),
			wantErr: false,
		},
	}

	// For each testcase
	for _, tc := range testCases {
		testCase := tc
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Initialize mock
			ctx := context.Background()

			// Initialize transformer
			underTest, err := Transformer("kP1yHnBcOhjowNFXSCyycSuXdUqTlbuE6ES5tTp-I_o=")
			if err != nil {
				t.Fatalf("unable to initialize transformer: %v", err)
			}

			// Do the call
			got, err := underTest.To(ctx, testCase.input)

			// Assert results expectations
			if (err != nil) != testCase.wantErr {
				t.Errorf("error during the call, error = %v, wantErr %v", err, testCase.wantErr)
				return
			}
			if testCase.wantErr {
				return
			}
			out, err := underTest.From(ctx, got)
			if err != nil {
				t.Errorf("error during the Paseto.From() call, error = %v", err)
			}
			if diff := cmp.Diff(out, testCase.input); diff != "" {
				t.Errorf("%q. Paseto.To():\n-got/+want\ndiff %s", testCase.name, diff)
			}
		})
	}
}
