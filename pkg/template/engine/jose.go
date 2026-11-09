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

// JWEDecrypter returns a JWE decryption function restricted to the given
// algorithm allow-lists.
func JWEDecrypter(opts []crypto.JOSEOption) func(key, token string) (interface{}, error) {
	return func(key, token string) (interface{}, error) {
		return crypto.DecryptJWE(key, token, opts...)
	}
}

// JWTParser returns a JWT parsing function restricted to the given algorithm
// allow-lists.
func JWTParser(opts []crypto.JOSEOption) func(token string) (interface{}, error) {
	return func(token string) (interface{}, error) {
		return crypto.ParseJWT(token, opts...)
	}
}

// JWTVerifier returns a JWT verification function restricted to the given
// algorithm allow-lists.
func JWTVerifier(opts []crypto.JOSEOption) func(token string, key interface{}) (interface{}, error) {
	return func(token string, key interface{}) (interface{}, error) {
		return crypto.VerifyJWT(token, key, opts...)
	}
}
