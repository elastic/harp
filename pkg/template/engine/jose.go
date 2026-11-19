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

package engine

import (
	"github.com/elastic/harp/pkg/sdk/security/crypto"
)

// decryptJWE decrypts a JWE token. The expected key management algorithm and
// content encryption can be given as trailing arguments, otherwise the crypto
// package defaults apply.
//
//	{{ decryptJwe $key $token }}
//	{{ decryptJwe $key $token "A256KW" "A256GCM" }}
func decryptJWE(key, token string, algs ...string) (interface{}, error) {
	opts, err := crypto.ParseJWEAlgorithms(algs...)
	if err != nil {
		return nil, err
	}

	return crypto.DecryptJWE(key, token, opts...)
}

// parseJWT parses a JWT without signature verification. The accepted signature
// algorithms can be given as trailing arguments, otherwise the crypto package
// defaults apply.
//
//	{{ parseJwt $token }}
//	{{ parseJwt $token "ES256" }}
func parseJWT(token string, algs ...string) (interface{}, error) {
	opts, err := crypto.ParseJWTAlgorithms(algs...)
	if err != nil {
		return nil, err
	}

	return crypto.ParseJWT(token, opts...)
}

// verifyJWT parses a JWT and verifies its signature. The accepted signature
// algorithms can be given as trailing arguments, otherwise the crypto package
// defaults apply.
//
//	{{ verifyJwt $token $key }}
//	{{ verifyJwt $token $key "ES256" }}
func verifyJWT(token string, key interface{}, algs ...string) (interface{}, error) {
	opts, err := crypto.ParseJWTAlgorithms(algs...)
	if err != nil {
		return nil, err
	}

	return crypto.VerifyJWT(token, key, opts...)
}
