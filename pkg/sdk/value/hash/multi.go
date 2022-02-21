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
	"encoding/hex"
	"fmt"
	"hash"
	"io"
)

func NewMultiHash(r io.Reader, algorithms ...string) (map[string]string, error) {
	hashers := map[string]hash.Hash{}

	// Instanciate hashers
	for _, algo := range algorithms {
		// Create an hasher instance.
		h, err := NewHasher(algo)
		if err != nil {
			return nil, fmt.Errorf("unable to initialize '%s' algorithm: %w", algo, err)
		}

		// Assign to hashers.
		hashers[algo] = h
	}

	// Copy to all hashers
	_, err := io.Copy(io.MultiWriter(hashToMultiWriter(hashers)), r)
	if err != nil {
		return nil, err
	}

	// Finalize
	var res = make(map[string]string)
	for algo, v := range hashers {
		res[algo] = hex.EncodeToString(v.Sum(nil))
	}

	// No error
	return res, nil
}

// -----------------------------------------------------------------------------

func hashToMultiWriter(hashers map[string]hash.Hash) io.Writer {
	var w = make([]io.Writer, 0, len(hashers))
	for _, v := range hashers {
		w = append(w, v)
	}
	return io.MultiWriter(w...)
}
