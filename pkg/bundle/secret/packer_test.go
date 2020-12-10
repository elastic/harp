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

package secret

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	fuzz "github.com/google/gofuzz"
)

func Test_Pack_Pack_Unpack(t *testing.T) {
	testCases := []struct {
		desc    string
		in      interface{}
		wantErr bool
	}{
		{
			desc:    "empty struct",
			in:      map[interface{}]interface{}{},
			wantErr: true,
		},
		{
			desc:    "string",
			in:      "foo",
			wantErr: false,
		},
		{
			desc:    "bytes",
			in:      []byte("foo"),
			wantErr: false,
		},
		{
			desc:    "uint8 array",
			in:      []uint8("foo"),
			wantErr: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got, err := Pack(tC.in)
			// Assert results expectations
			if (err != nil) != tC.wantErr {
				t.Errorf("error during the call, error = %v, wantErr %v", err, tC.wantErr)
				return
			}

			var out interface{}
			err = Unpack(got, &out)
			// Assert results expectations
			if (err != nil) != tC.wantErr {
				t.Errorf("error during the call, error = %v, wantErr %v", err, tC.wantErr)
				return
			}

			if tC.wantErr {
				return
			}

			if diff := cmp.Diff(out, tC.in); diff != "" {
				t.Errorf("%q. Secret.PackUnpack():\n-got/+want\ndiff %s", tC.desc, diff)
			}
		})
	}
}

func Test_UnPack_Fuzz(t *testing.T) {
	// Making sure the descrption never panics
	for i := 0; i < 100000; i++ {
		f := fuzz.New()

		var (
			in  []byte
			out struct{}
		)

		// Fuzz input
		f.Fuzz(&in)

		// Execute
		Unpack(in, &out)
	}
}
