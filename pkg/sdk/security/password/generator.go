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
	"fmt"
	"math"

	"github.com/sethvargo/go-password/password"
)

const (
	// MaxPasswordLen defines the upper bound for password generation length.
	MaxPasswordLen = 1024
)

// Generate a custom password
func Generate(length, numDigits, numSymbol int, noUpper, allowRepeat bool) (string, error) {
	// Check parameters
	if length < 0 || length > MaxPasswordLen {
		length = MaxPasswordLen
	}
	if numDigits < 0 || numDigits > MaxPasswordLen {
		numDigits = int(math.Floor(0.2 * float64(length))) // 20% of length
	}
	if numSymbol < 0 || numSymbol > MaxPasswordLen {
		numSymbol = int(math.Floor(0.1 * float64(length))) // 10% of length
	}

	p, err := password.Generate(length, numDigits, numSymbol, noUpper, allowRepeat)
	if err != nil {
		return "", fmt.Errorf("unable to generate a password: %w", err)
	}

	return p, nil
}

// FromProfile uses given profile to generate a password which profile constraints.
func FromProfile(p *Profile) (string, error) {
	// Check parameters
	if p == nil {
		return "", fmt.Errorf("unable to generate paswword without a nil profile")
	}

	// Delegate to generator
	return Generate(p.Length, p.NumDigits, p.NumSymbol, p.NoUpper, p.AllowRepeat)
}

// Paranoid generates a 64 character length password with 10 digits count,
// 10 symbol count, with all cases, and character repeat.
func Paranoid() (string, error) {
	return FromProfile(ProfileParanoid)
}

// NoSymbol generates a 32 character length password with 10 digits count,
// no symbol, with all cases, and character repeat.
func NoSymbol() (string, error) {
	return FromProfile(ProfileNoSymbol)
}

// Strong generates a 32 character length password with 10 digits count,
// 10 symbol count, with all cases, and character repeat.
func Strong() (string, error) {
	return FromProfile(ProfileStrong)
}
