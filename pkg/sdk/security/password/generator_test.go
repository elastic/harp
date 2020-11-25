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

package password

import (
	"testing"

	fuzz "github.com/google/gofuzz"
)

func TestFromProfile(t *testing.T) {
	type args struct {
		p *Profile
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "nil",
			wantErr: true,
		},
		{
			name: "paranoid",
			args: args{
				p: ProfileParanoid,
			},
			wantErr: false,
		},
		{
			name: "noSymbol",
			args: args{
				p: ProfileNoSymbol,
			},
			wantErr: false,
		},
		{
			name: "strong",
			args: args{
				p: ProfileStrong,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FromProfile(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromProfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestPredefined(t *testing.T) {
	tests := []struct {
		name       string
		callable   func() (string, error)
		wantLength int
		wantErr    bool
	}{
		{
			name:       "paranoid",
			callable:   Paranoid,
			wantLength: ProfileParanoid.Length,
			wantErr:    false,
		},
		{
			name:       "strong",
			callable:   Strong,
			wantLength: ProfileStrong.Length,
			wantErr:    false,
		},
		{
			name:       "noSymbol",
			callable:   NoSymbol,
			wantLength: ProfileNoSymbol.Length,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.callable()
			if (err != nil) != tt.wantErr {
				t.Errorf("Predefined() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotLength := len(got)
			if (tt.wantLength > 0) && tt.wantLength != gotLength {
				t.Errorf("Predefined() expected length = %v, got %v", tt.wantLength, gotLength)
				return
			}
		})
	}
}

// -----------------------------------------------------------------------------

func TestGenerate_Fuzz(t *testing.T) {
	// Making sure that it never panics
	for i := 0; i < 50; i++ {
		f := fuzz.New()

		// Prepare arguments
		var (
			length, numDigits, numSymbol int
			noUpper, allowRepeat         bool
		)

		// Fuzz input
		f.Fuzz(&length)
		f.Fuzz(&numDigits)
		f.Fuzz(&numSymbol)
		f.Fuzz(&noUpper)
		f.Fuzz(&allowRepeat)

		// Execute
		Generate(length, numDigits, numSymbol, noUpper, allowRepeat)
	}
}

func TestFromProfile_Fuzz(t *testing.T) {
	// Making sure that it never panics
	for i := 0; i < 50; i++ {
		f := fuzz.New()

		// Prepare arguments
		var p Profile

		// Fuzz input
		f.Fuzz(&p)

		// Execute
		FromProfile(&p)
	}
}
