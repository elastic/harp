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

package identity

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_Identity_From(t *testing.T) {
	testCases := []struct {
		name    string
		input   []byte
		wantErr bool
		want    []byte
	}{
		{
			name:    "Nil input",
			input:   nil,
			wantErr: false,
			want:    nil,
		},
		{
			name:    "Empty input",
			input:   []byte{},
			wantErr: false,
			want:    []byte{},
		},
		{
			name:    "Something",
			input:   []byte("foo"),
			wantErr: false,
			want:    []byte("foo"),
		},
	}
	for _, tC := range testCases {
		testCase := tC
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			underTest := Transformer()

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
			if diff := cmp.Diff(got, testCase.input); diff != "" {
				t.Errorf("%q. Identity.From():\n-got/+want\ndiff %s", testCase.name, diff)
			}
		})
	}
}

func Test_Identity_To(t *testing.T) {
	testCases := []struct {
		name    string
		input   []byte
		wantErr bool
		want    []byte
	}{
		{
			name:    "Nil input",
			input:   nil,
			wantErr: false,
			want:    nil,
		},
		{
			name:    "Empty input",
			input:   []byte{},
			wantErr: false,
			want:    []byte{},
		},
		{
			name:    "Something",
			input:   []byte("foo"),
			wantErr: false,
			want:    []byte("foo"),
		},
	}
	for _, tC := range testCases {
		testCase := tC
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			underTest := Transformer()

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
			if diff := cmp.Diff(got, testCase.input); diff != "" {
				t.Errorf("%q. Identity.To():\n-got/+want\ndiff %s", testCase.name, diff)
			}
		})
	}
}
