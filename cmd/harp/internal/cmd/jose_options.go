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

package cmd

import (
	"fmt"
	"strings"

	jose "github.com/go-jose/go-jose/v4"

	"github.com/elastic/harp/pkg/sdk/security/crypto"
)

// Default JOSE algorithm allow-lists used by the template functions
// (decryptJwe, parseJwt, verifyJwt). They can be overridden using CLI flags.
var (
	defaultJWEKeyAlgorithms = []string{
		string(jose.PBES2_HS256_A128KW),
	}
	defaultJWEContentEncryptions = []string{
		string(jose.A128GCM),
	}
	defaultJWTSignatureAlgorithms = []string{
		string(jose.EdDSA),
		string(jose.HS256), string(jose.HS384), string(jose.HS512),
		string(jose.RS256), string(jose.RS384), string(jose.RS512),
		string(jose.PS256), string(jose.PS384), string(jose.PS512),
		string(jose.ES256), string(jose.ES384), string(jose.ES512),
	}

	supportedJWEKeyAlgorithms = []string{
		string(jose.ED25519), string(jose.RSA1_5), string(jose.RSA_OAEP), string(jose.RSA_OAEP_256),
		string(jose.A128KW), string(jose.A192KW), string(jose.A256KW), string(jose.DIRECT),
		string(jose.ECDH_ES), string(jose.ECDH_ES_A128KW), string(jose.ECDH_ES_A192KW), string(jose.ECDH_ES_A256KW),
		string(jose.A128GCMKW), string(jose.A192GCMKW), string(jose.A256GCMKW),
		string(jose.PBES2_HS256_A128KW), string(jose.PBES2_HS384_A192KW), string(jose.PBES2_HS512_A256KW),
	}
	supportedJWEContentEncryptions = []string{
		string(jose.A128CBC_HS256), string(jose.A192CBC_HS384), string(jose.A256CBC_HS512),
		string(jose.A128GCM), string(jose.A192GCM), string(jose.A256GCM),
	}
)

// joseOptions validates the given allow-lists and returns the matching crypto
// options.
func joseOptions(keyAlgorithms, contentEncryptions, signatureAlgorithms []string) ([]crypto.JOSEOption, error) {
	keyAlgs, err := validateAlgorithms("jwe-key-algorithms", keyAlgorithms, supportedJWEKeyAlgorithms)
	if err != nil {
		return nil, err
	}
	contentEncs, err := validateAlgorithms("jwe-content-encryptions", contentEncryptions, supportedJWEContentEncryptions)
	if err != nil {
		return nil, err
	}
	sigAlgs, err := validateAlgorithms("jwt-signature-algorithms", signatureAlgorithms, defaultJWTSignatureAlgorithms)
	if err != nil {
		return nil, err
	}

	joseKeyAlgs := make([]jose.KeyAlgorithm, 0, len(keyAlgs))
	for _, v := range keyAlgs {
		joseKeyAlgs = append(joseKeyAlgs, jose.KeyAlgorithm(v))
	}
	joseContentEncs := make([]jose.ContentEncryption, 0, len(contentEncs))
	for _, v := range contentEncs {
		joseContentEncs = append(joseContentEncs, jose.ContentEncryption(v))
	}
	joseSigAlgs := make([]jose.SignatureAlgorithm, 0, len(sigAlgs))
	for _, v := range sigAlgs {
		joseSigAlgs = append(joseSigAlgs, jose.SignatureAlgorithm(v))
	}

	return []crypto.JOSEOption{
		crypto.WithKeyAlgorithms(joseKeyAlgs...),
		crypto.WithContentEncryptions(joseContentEncs...),
		crypto.WithSignatureAlgorithms(joseSigAlgs...),
	}, nil
}

// validateAlgorithms ensures the values are non-empty and supported. "none" is
// never accepted.
func validateAlgorithms(flag string, values, supported []string) ([]string, error) {
	if len(values) == 0 {
		return nil, fmt.Errorf("--%s must contain at least one algorithm", flag)
	}

	allowed := map[string]struct{}{}
	for _, s := range supported {
		allowed[s] = struct{}{}
	}

	out := make([]string, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if _, ok := allowed[v]; !ok {
			return nil, fmt.Errorf("--%s: unsupported algorithm %q (supported: %s)", flag, v, strings.Join(supported, ", "))
		}
		out = append(out, v)
	}

	return out, nil
}
