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

package diceware

import (
	"fmt"
	"strings"

	"github.com/sethvargo/go-diceware/diceware"
)

const (
	// MinWordCount defines the lowest bound for allowed word count.
	MinWordCount = 4
	// MaxWordCount defines the highest bound for allowed word count.
	MaxWordCount = 24
	// BasicWordCount defines basic passphrase word count (4 words).
	BasicWordCount = 4
	// StrongWordCount defines strong passphrase word count (8 words).
	StrongWordCount = 8
	// ParanoidWordCount defines paranoid passphrase word count (12 words).
	ParanoidWordCount = 12
	// MasterWordCount defines master passphrase word count (24 words).
	MasterWordCount = 24
)

// Diceware generates a passphrase using english words
func Diceware(count int) (string, error) {
	// Check parameters
	if count < MinWordCount {
		count = MinWordCount
	}
	if count > MaxWordCount {
		count = MaxWordCount
	}

	// Generate word list
	list, err := diceware.Generate(count)
	if err != nil {
		return "", fmt.Errorf("unable to generate daceware passphrase: %w", err)
	}

	// Assemble result
	return strings.Join(list, "-"), nil
}

// Basic generates 4 words diceware passphrase
func Basic() (string, error) {
	return Diceware(BasicWordCount)
}

// Strong generates 8 words diceware passphrase
func Strong() (string, error) {
	return Diceware(StrongWordCount)
}

// Paranoid generates 12 words diceware passphrase
func Paranoid() (string, error) {
	return Diceware(ParanoidWordCount)
}

// Master generates 24 words diceware passphrase
func Master() (string, error) {
	return Diceware(MasterWordCount)
}
