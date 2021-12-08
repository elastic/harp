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

package rfc6979

// From https://github.com/codahale/rfc6979

import (
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"testing"
)

// https://tools.ietf.org/html/rfc6979#appendix-A.1
func TestGenerateSecret(t *testing.T) {
	q, _ := new(big.Int).SetString("4000000000000000000020108A2E0CC0D99F8A5EF", 16)
	x, _ := new(big.Int).SetString("09A4D6792295A7F730FC3F2B49CBC0F62E862272F", 16)
	hash, _ := hex.DecodeString("AF2BDBE1AA9B6EC1E2ADE1D694F41FC71A831D0268E9891562113D8A62ADD1BF")

	expected, _ := new(big.Int).SetString("23AF4074C90A02B3FE61D286D5C87F425E6BDD81B", 16)
	var actual *big.Int
	generateSecret(q, x, sha256.New, hash, func(k *big.Int) bool {
		actual = k
		return true
	})

	if actual.Cmp(expected) != 0 {
		t.Errorf("Expected %x, got %x", expected, actual)
	}
}
