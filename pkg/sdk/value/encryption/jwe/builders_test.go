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

package jwe

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Transformer_JWE_InvalidKey(t *testing.T) {
	keys := []string{
		"",
		"foo",
		"123456",
	}
	for _, k := range keys {
		key := k
		t.Run(fmt.Sprintf("key `%s`", key), func(t *testing.T) {
			underTest, err := Transformer(key)
			if err == nil {
				t.Fatalf("Transformer should raise an error with key `%s`", key)
			}
			if underTest != nil {
				t.Fatalf("Transformer instance should be nil")
			}
		})
	}
}

func Test_Transformer(t *testing.T) {
	keys := []string{
		"a128kw:abSOB6OHnFK1CHIm60OXsA==",
		"a192kw:b4JtOwQLOks1-RWxXUh5eG54nbdBihLT",
		"a256kw:TkxS6qSV6eDBjn29JmU2ieMPnuCZNn3JelI1CDNqAQ8=",
		"pbes2-hs256-a128kw:stalemate-parkway-hardened-jeep-shrink-dimmer-platter-pretense",
		"pbes2-hs384-a192kw:stalemate-parkway-hardened-jeep-shrink-dimmer-platter-pretense",
		"pbes2-hs512-a256kw:stalemate-parkway-hardened-jeep-shrink-dimmer-platter-pretense",
	}
	for _, k := range keys {
		key := k
		t.Run(fmt.Sprintf("key `%s`", key), func(t *testing.T) {
			underTest, err := Transformer(key)
			assert.NoError(t, err)
			assert.NotNil(t, underTest)

			// Try to encrypt
			ctx := context.Background()
			encrypted, err := underTest.To(ctx, []byte("cleartext"))
			assert.NoError(t, err)
			assert.NotEmpty(t, encrypted)

			// Try to decrypt
			out, err := underTest.From(ctx, encrypted)
			assert.NoError(t, err)
			assert.Equal(t, []byte("cleartext"), out)
		})
	}
}
