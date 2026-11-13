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

//go:build !fips

package crypto

import (
	"slices"
	"testing"

	"github.com/go-jose/go-jose/v4"
)

func TestJOSESignatureAlgorithms_NonFIPS(t *testing.T) {
	if !slices.Contains(DefaultJWTSignatureAlgorithms(), jose.EdDSA) {
		t.Error("EdDSA should be in the default signature algorithms outside FIPS builds")
	}
	if _, err := ParseJWTAlgorithms("EdDSA"); err != nil {
		t.Errorf("ParseJWTAlgorithms(EdDSA) should succeed outside FIPS builds: %v", err)
	}
}
