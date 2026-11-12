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

package crypto

import (
	"fmt"
	"strings"

	jose "github.com/go-jose/go-jose/v4"
)

// JOSEOption configures the algorithm allow-lists used when parsing JWE and
// JWT tokens.
type JOSEOption func(*joseOptions)

type joseOptions struct {
	keyAlgorithms       []jose.KeyAlgorithm
	contentEncryptions  []jose.ContentEncryption
	signatureAlgorithms []jose.SignatureAlgorithm
}

// DefaultJWEKeyAlgorithms returns the JWE key management algorithms accepted
// on decryption when WithKeyAlgorithms is not provided.
func DefaultJWEKeyAlgorithms() []jose.KeyAlgorithm {
	return []jose.KeyAlgorithm{jose.PBES2_HS256_A128KW}
}

// DefaultJWEContentEncryptions returns the JWE content encryption algorithms
// accepted on decryption when WithContentEncryptions is not provided.
func DefaultJWEContentEncryptions() []jose.ContentEncryption {
	return []jose.ContentEncryption{jose.A128GCM}
}

// DefaultJWTSignatureAlgorithms returns the signature algorithms accepted when
// parsing or verifying JWT tokens when WithSignatureAlgorithms is not provided.
// The "none" algorithm is never part of the list.
func DefaultJWTSignatureAlgorithms() []jose.SignatureAlgorithm {
	return []jose.SignatureAlgorithm{
		jose.EdDSA,
		jose.HS256, jose.HS384, jose.HS512,
		jose.RS256, jose.RS384, jose.RS512,
		jose.PS256, jose.PS384, jose.PS512,
		jose.ES256, jose.ES384, jose.ES512,
	}
}

// newJOSEOptions returns options initialized with safe defaults, then applies
// the given options on top of them.
func newJOSEOptions(opts []JOSEOption) *joseOptions {
	o := &joseOptions{
		keyAlgorithms:       DefaultJWEKeyAlgorithms(),
		contentEncryptions:  DefaultJWEContentEncryptions(),
		signatureAlgorithms: DefaultJWTSignatureAlgorithms(),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(o)
		}
	}

	return o
}

// WithKeyAlgorithms sets the JWE key management algorithms accepted on
// decryption.
func WithKeyAlgorithms(algs ...jose.KeyAlgorithm) JOSEOption {
	return func(o *joseOptions) {
		if len(algs) > 0 {
			o.keyAlgorithms = algs
		}
	}
}

// WithContentEncryptions sets the JWE content encryption algorithms accepted on
// decryption.
func WithContentEncryptions(encs ...jose.ContentEncryption) JOSEOption {
	return func(o *joseOptions) {
		if len(encs) > 0 {
			o.contentEncryptions = encs
		}
	}
}

// WithSignatureAlgorithms sets the signature algorithms accepted when parsing
// or verifying JWT tokens.
func WithSignatureAlgorithms(algs ...jose.SignatureAlgorithm) JOSEOption {
	return func(o *joseOptions) {
		if len(algs) > 0 {
			o.signatureAlgorithms = algs
		}
	}
}

// SupportedJWEKeyAlgorithms returns the JWE key management algorithms that can
// be selected in an allow-list.
//
// RSA1_5 (RSAES-PKCS1-v1_5) is intentionally excluded, see RFC 8725 section 3.2.
func SupportedJWEKeyAlgorithms() []jose.KeyAlgorithm {
	return []jose.KeyAlgorithm{
		jose.RSA_OAEP, jose.RSA_OAEP_256,
		jose.A128KW, jose.A192KW, jose.A256KW, jose.DIRECT,
		jose.ECDH_ES, jose.ECDH_ES_A128KW, jose.ECDH_ES_A192KW, jose.ECDH_ES_A256KW,
		jose.A128GCMKW, jose.A192GCMKW, jose.A256GCMKW,
		jose.PBES2_HS256_A128KW, jose.PBES2_HS384_A192KW, jose.PBES2_HS512_A256KW,
	}
}

// SupportedJWEContentEncryptions returns the JWE content encryption algorithms
// that can be selected in an allow-list.
func SupportedJWEContentEncryptions() []jose.ContentEncryption {
	return []jose.ContentEncryption{
		jose.A128CBC_HS256, jose.A192CBC_HS384, jose.A256CBC_HS512,
		jose.A128GCM, jose.A192GCM, jose.A256GCM,
	}
}

// SupportedJWTSignatureAlgorithms returns the signature algorithms that can be
// selected in an allow-list. The "none" algorithm is never supported.
func SupportedJWTSignatureAlgorithms() []jose.SignatureAlgorithm {
	return DefaultJWTSignatureAlgorithms()
}

// ParseJOSEOptions validates the given algorithm names against the supported
// lists and returns the matching options. An empty list leaves the
// corresponding default in place.
func ParseJOSEOptions(keyAlgorithms, contentEncryptions, signatureAlgorithms []string) ([]JOSEOption, error) {
	keyAlgs, err := parseAlgorithms("JWE key algorithms", keyAlgorithms, SupportedJWEKeyAlgorithms())
	if err != nil {
		return nil, err
	}
	contentEncs, err := parseAlgorithms("JWE content encryptions", contentEncryptions, SupportedJWEContentEncryptions())
	if err != nil {
		return nil, err
	}
	sigAlgs, err := parseAlgorithms("JWT signature algorithms", signatureAlgorithms, SupportedJWTSignatureAlgorithms())
	if err != nil {
		return nil, err
	}

	return []JOSEOption{
		WithKeyAlgorithms(keyAlgs...),
		WithContentEncryptions(contentEncs...),
		WithSignatureAlgorithms(sigAlgs...),
	}, nil
}

// parseAlgorithms ensures the values are supported and converts them to the
// typed algorithm. Empty input returns an empty result.
func parseAlgorithms[T ~string](name string, values []string, supported []T) ([]T, error) {
	out := make([]T, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)

		found := false
		for _, s := range supported {
			if string(s) == v {
				out = append(out, s)
				found = true

				break
			}
		}
		if !found {
			names := make([]string, 0, len(supported))
			for _, s := range supported {
				names = append(names, string(s))
			}

			return nil, fmt.Errorf("%s: unsupported algorithm %q (supported: %s)", name, v, strings.Join(names, ", "))
		}
	}

	return out, nil
}
