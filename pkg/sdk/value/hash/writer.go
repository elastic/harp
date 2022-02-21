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

package hash

import (
	// ensure crypto algorithms are initialized
	"crypto"
	//nolint:gosec // For legacy compatibility
	_ "crypto/md5"
	//nolint:gosec // For legacy compatibility
	_ "crypto/sha1"
	_ "crypto/sha256"
	_ "crypto/sha512"
	"fmt"
	"hash"
	"sort"
	"strings"

	// ensure crypto algorithms are initialized
	_ "golang.org/x/crypto/blake2b"
	//nolint:staticcheck // For legacy compatibility
	_ "golang.org/x/crypto/md4"
	//nolint:staticcheck // For legacy compatibility
	_ "golang.org/x/crypto/ripemd160"
	_ "golang.org/x/crypto/sha3"
)

var (
	name2Hash = map[string]crypto.Hash{}
)

// -----------------------------------------------------------------------------

func init() {
	name2Hash["md4"] = crypto.MD4
	name2Hash["md5"] = crypto.MD5
	name2Hash["sha1"] = crypto.SHA1
	name2Hash["sha224"] = crypto.SHA224
	name2Hash["sha256"] = crypto.SHA256
	name2Hash["sha512"] = crypto.SHA512
	name2Hash["sha512/224"] = crypto.SHA512_224
	name2Hash["sha512/256"] = crypto.SHA512_256
	name2Hash["blake2b-256"] = crypto.BLAKE2b_256
	name2Hash["blake2b-384"] = crypto.BLAKE2b_384
	name2Hash["blake2b-512"] = crypto.BLAKE2b_512
	name2Hash["blake2s-256"] = crypto.BLAKE2s_256
	name2Hash["ripemd160"] = crypto.RIPEMD160
	name2Hash["sha3-224"] = crypto.SHA3_224
	name2Hash["sha3-256"] = crypto.SHA3_256
	name2Hash["sha3-384"] = crypto.SHA3_384
	name2Hash["sha3-512"] = crypto.SHA3_512
}

// NewHasher returns a hasher instance.
func NewHasher(algorithm string) (hash.Hash, error) {
	// Normalize input
	algorithm = strings.TrimSpace(strings.ToLower(algorithm))

	// Resolve algorithm
	hf, ok := name2Hash[algorithm]
	if !ok {
		return nil, fmt.Errorf("unsupported hash algorithm '%s'", algorithm)
	}
	if !hf.Available() {
		return nil, fmt.Errorf("hash algorithm '%s' is not available", algorithm)
	}

	// Build hash instance.
	h := hf.New()

	// No error
	return h, nil
}

// SupportedAlgorithms returns the available hash algorithms.
func SupportedAlgorithms() []string {
	res := []string{}
	for n, c := range name2Hash {
		if c.Available() {
			res = append(res, n)
		}
	}

	// Sort all algorithms
	sort.Strings(res)

	return res
}
