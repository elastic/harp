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

//go:build fips

package crypto

import (
	"slices"
	"testing"

	"github.com/go-jose/go-jose/v4"
)

func TestJOSESignatureAlgorithms_FIPS(t *testing.T) {
	if slices.Contains(DefaultJWTSignatureAlgorithms(), jose.EdDSA) {
		t.Error("EdDSA must not be in the default signature algorithms in FIPS builds")
	}
	if slices.Contains(SupportedJWTSignatureAlgorithms(), jose.EdDSA) {
		t.Error("EdDSA must not be in the supported signature algorithms in FIPS builds")
	}

	// Explicit options must not re-enable it, and must not be modified.
	opt := WithSignatureAlgorithms(jose.EdDSA, jose.RS256)
	for i := 0; i < 2; i++ {
		got := newJOSEOptions([]JOSEOption{opt}).signatureAlgorithms
		if !slices.Equal(got, []jose.SignatureAlgorithm{jose.RS256}) {
			t.Errorf("call %d: signature algorithms = %v, want [RS256]", i, got)
		}
	}

	if _, err := ParseJWTAlgorithms("EdDSA"); err == nil {
		t.Error("ParseJWTAlgorithms(EdDSA) should fail in FIPS builds")
	}
}
