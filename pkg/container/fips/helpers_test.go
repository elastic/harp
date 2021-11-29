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

package fips

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"testing"

	containerv1 "github.com/elastic/harp/api/gen/go/harp/container/v1"
	"github.com/stretchr/testify/assert"
)

func Test_deriveSharedKeyFromRecipient(t *testing.T) {
	key1, err := ecdsa.GenerateKey(elliptic.P384(), bytes.NewReader([]byte("00001-deterministic-buffer-for-tests-26FBE7DED9E992BC36C06C988C1AC8A1E672B4B5959EF60672A983EFA7C8EE0F")))
	assert.NoError(t, err)
	assert.NotNil(t, key1)

	key2, err := ecdsa.GenerateKey(elliptic.P384(), bytes.NewReader([]byte("00002-deterministic-buffer-for-tests-37ACB0DD3A3CE5A0960CCE0F6A0D7E663DFFD221FBE8EEB03B20D3AD91BCDD55")))
	assert.NoError(t, err)
	assert.NotNil(t, key2)

	dk1, err := deriveSharedKeyFromRecipient(&key1.PublicKey, key2)
	assert.NoError(t, err)
	assert.Equal(t, &[32]byte{
		0x1a, 0x71, 0x90, 0x5d, 0x18, 0x25, 0xba, 0x52,
		0x6c, 0xf7, 0xf8, 0x33, 0x65, 0x37, 0x80, 0xb2,
		0xe8, 0xa0, 0x9, 0xf6, 0x5c, 0x5c, 0x3b, 0xfb,
		0x39, 0xb1, 0x42, 0x81, 0xd7, 0x27, 0x28, 0x43,
	}, dk1)

	dk2, err := deriveSharedKeyFromRecipient(&key2.PublicKey, key1)
	assert.NoError(t, err)
	assert.Equal(t, dk1, dk2)
}

func Test_keyIdentifierFromDerivedKey(t *testing.T) {
	dk := &[32]byte{
		0x9f, 0x6c, 0xb8, 0x33, 0xf4, 0x7a, 0x4, 0xb2,
		0xba, 0x65, 0x30, 0xf2, 0xc, 0x7c, 0xb1, 0x30,
		0x22, 0xa0, 0x6a, 0x15, 0x57, 0x73, 0xc1, 0xa9,
		0xc7, 0x21, 0x48, 0xdd, 0x3c, 0xc8, 0x36, 0xc7,
	}

	id, err := keyIdentifierFromDerivedKey(dk)
	assert.NoError(t, err)
	assert.Equal(t, []byte{
		0xe0, 0xe9, 0xd5, 0xc5, 0x7a, 0x9e, 0x1c, 0x3,
		0x9d, 0x4b, 0xc0, 0x21, 0x6e, 0x72, 0x1a, 0xda,
		0xac, 0xd0, 0x57, 0xb8, 0x21, 0x16, 0x48, 0xac,
		0x46, 0x7c, 0x64, 0xf9, 0x4d, 0xe5, 0x86, 0x23,
	}, id)
}

func Test_packRecipient(t *testing.T) {
	payloadKey := &[32]byte{}

	key1, err := ecdsa.GenerateKey(elliptic.P384(), bytes.NewReader([]byte("00001-deterministic-buffer-for-tests-26FBE7DED9E992BC36C06C988C1AC8A1E672B4B5959EF60672A983EFA7C8EE0F")))
	assert.NoError(t, err)
	assert.NotNil(t, key1)

	key2, err := ecdsa.GenerateKey(elliptic.P384(), bytes.NewReader([]byte("00002-deterministic-buffer-for-tests-37ACB0DD3A3CE5A0960CCE0F6A0D7E663DFFD221FBE8EEB03B20D3AD91BCDD55")))
	assert.NoError(t, err)
	assert.NotNil(t, key2)

	recipient, err := packRecipient(payloadKey, key1, &key2.PublicKey)
	assert.NoError(t, err)
	assert.NotNil(t, recipient)
	assert.Equal(t, []byte{0xeb, 0xe4, 0xc2, 0xff, 0xf7, 0xb5, 0x59, 0x92, 0xf1, 0x6a, 0x2e, 0xe, 0xfd, 0x6c, 0x3b, 0x83, 0x8e, 0xbc, 0x71, 0x83, 0x5f, 0xda, 0x2f, 0xcf, 0x2d, 0x72, 0x9d, 0x59, 0xc0, 0x4, 0xf7, 0x67}, recipient.Identifier)
	assert.Equal(t, 60, len(recipient.Key))
}

func Test_tryRecipientKeys(t *testing.T) {
	payloadKey := &[32]byte{}

	key1, err := ecdsa.GenerateKey(elliptic.P384(), bytes.NewReader([]byte("00001-deterministic-buffer-for-tests-26FBE7DED9E992BC36C06C988C1AC8A1E672B4B5959EF60672A983EFA7C8EE0F")))
	assert.NoError(t, err)
	assert.NotNil(t, key1)

	key2, err := ecdsa.GenerateKey(elliptic.P384(), bytes.NewReader([]byte("00002-deterministic-buffer-for-tests-37ACB0DD3A3CE5A0960CCE0F6A0D7E663DFFD221FBE8EEB03B20D3AD91BCDD55")))
	assert.NoError(t, err)
	assert.NotNil(t, key2)

	recipient, err := packRecipient(payloadKey, key1, &key2.PublicKey)
	assert.NoError(t, err)
	assert.NotNil(t, recipient)
	assert.Equal(t, []byte{0xeb, 0xe4, 0xc2, 0xff, 0xf7, 0xb5, 0x59, 0x92, 0xf1, 0x6a, 0x2e, 0xe, 0xfd, 0x6c, 0x3b, 0x83, 0x8e, 0xbc, 0x71, 0x83, 0x5f, 0xda, 0x2f, 0xcf, 0x2d, 0x72, 0x9d, 0x59, 0xc0, 0x4, 0xf7, 0x67}, recipient.Identifier)
	assert.Equal(t, 60, len(recipient.Key))

	// -------------------------------------------------------------------------
	dk, err := deriveSharedKeyFromRecipient(&key1.PublicKey, key2)
	assert.NoError(t, err)
	assert.Equal(t, &[32]byte{0x1a, 0x71, 0x90, 0x5d, 0x18, 0x25, 0xba, 0x52, 0x6c, 0xf7, 0xf8, 0x33, 0x65, 0x37, 0x80, 0xb2, 0xe8, 0xa0, 0x9, 0xf6, 0x5c, 0x5c, 0x3b, 0xfb, 0x39, 0xb1, 0x42, 0x81, 0xd7, 0x27, 0x28, 0x43}, dk)

	expectedID := []byte{
		0xeb, 0xe4, 0xc2, 0xff, 0xf7, 0xb5, 0x59, 0x92,
		0xf1, 0x6a, 0x2e, 0xe, 0xfd, 0x6c, 0x3b, 0x83,
		0x8e, 0xbc, 0x71, 0x83, 0x5f, 0xda, 0x2f, 0xcf,
		0x2d, 0x72, 0x9d, 0x59, 0xc0, 0x4, 0xf7, 0x67,
	}
	id, err := keyIdentifierFromDerivedKey(dk)
	assert.NoError(t, err)
	assert.Equal(t, expectedID, id)
	assert.Equal(t, expectedID, recipient.Identifier)

	decodedPayloadKey, err := tryRecipientKeys(dk, []*containerv1.Recipient{
		recipient,
	})
	assert.NoError(t, err)
	assert.Equal(t, payloadKey[:], decodedPayloadKey)
}
