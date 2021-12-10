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
	"crypto/ecdsa"
	"crypto/elliptic"
	"hash"
	"math/big"
)

// SignECDSA signs an arbitrary length hash (which should be the result of
// hashing a larger message) using the private key, priv. It returns the
// signature as a pair of integers.
//
// Note that FIPS 186-3 section 4.6 specifies that the hash should be truncated
// to the byte-length of the subgroup. This function does not perform that
// truncation itself.
func SignECDSA(priv *ecdsa.PrivateKey, digest []byte, alg func() hash.Hash) (r, s *big.Int) {
	c := priv.PublicKey.Curve
	N := c.Params().N

	generateSecret(N, priv.D, alg, digest, func(k *big.Int) bool {
		inv := new(big.Int).ModInverse(k, N)
		r, _ = priv.Curve.ScalarBaseMult(k.Bytes())
		r.Mod(r, N)

		if r.Sign() == 0 {
			return false
		}

		e := hashToInt(digest, c)
		s = new(big.Int).Mul(priv.D, r)
		s.Add(s, e)
		s.Mul(s, inv)
		s.Mod(s, N)

		return s.Sign() != 0
	})

	return
}

// copied from crypto/ecdsa
func hashToInt(digest []byte, c elliptic.Curve) *big.Int {
	orderBits := c.Params().N.BitLen()
	orderBytes := (orderBits + 7) / 8
	if len(digest) > orderBytes {
		digest = digest[:orderBytes]
	}

	ret := new(big.Int).SetBytes(digest)
	excess := len(digest)*8 - orderBits
	if excess > 0 {
		ret.Rsh(ret, uint(excess))
	}
	return ret
}
