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

package age

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
)

func Test_Transformer_Age_InvalidKey(t *testing.T) {
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
	identity := `age-identity:AGE-SECRET-KEY-1W8E69DQEVASNK68FX7C6QLD99KTG96RHWW0EZ3RD0L29AHV4S84QHUAP4C`
	//recipient := `age-recipients:age1ce20pmz8z0ue97v7rz838v6pcpvzqan30lr40tjlzy40ez8eldrqf2zuxe`

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
			name: "valid",
			input: []byte(`-----BEGIN AGE ENCRYPTED FILE-----
YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IFgyNTUxOSBFeERxdmZ1SzAzZmVsSWtz
Si9XMVNaV2tLcEdCZUdOQy9RL2dmTnZsbGk4Ci9mcE1DVTZWQTdaenRhSVdQMjZP
ZGpCbmZiUkgzY0NXQk8rT25yQjA1b2sKLS0tIFJNT1R1WlZOTEpnZUNVRTFLZ3Nq
bkZqYUc4NHl2VGJuQU1SeEd5bkFoWmcKV6dwry0emJOt19Jh5IPGKiOzYmJ6AVA2
7aMKPOZ+rMNcgne6
-----END AGE ENCRYPTED FILE-----`),
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
			underTest, err := Transformer(identity)
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
				t.Errorf("%q. Age.From():\n-got/+want\ndiff %s", testCase.name, diff)
			}
		})
	}
}

func Test_Transformer_To(t *testing.T) {
	identity := `age-identity:AGE-SECRET-KEY-1W8E69DQEVASNK68FX7C6QLD99KTG96RHWW0EZ3RD0L29AHV4S84QHUAP4C`
	recipient := `age-recipients:age1ce20pmz8z0ue97v7rz838v6pcpvzqan30lr40tjlzy40ez8eldrqf2zuxe`

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
			encrypt, err := Transformer(recipient)
			if err != nil {
				t.Fatalf("unable to initialize transformer: %v", err)
			}

			// Do the call
			got, err := encrypt.To(ctx, testCase.input)

			// Assert results expectations
			if (err != nil) != testCase.wantErr {
				t.Errorf("error during the call, error = %v, wantErr %v", err, testCase.wantErr)
				return
			}
			if testCase.wantErr {
				return
			}

			// Initialize transformer
			decrypt, err := Transformer(identity)
			if err != nil {
				t.Fatalf("unable to initialize transformer: %v", err)
			}

			out, err := decrypt.From(ctx, got)
			if err != nil {
				t.Errorf("error during the Age.From() call, error = %v", err)
			}
			if diff := cmp.Diff(out, testCase.input); diff != "" {
				t.Errorf("%q. Age.To():\n-got/+want\ndiff %s", testCase.name, diff)
			}
		})
	}
}
