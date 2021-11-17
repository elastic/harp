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
package identity

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	securityIdentity = []byte(`{"@apiVersion": "harp.elastic.co/v1", "@kind": "ContainerIdentity", "@timestamp": "2021-11-15T11:58:13.662568Z", "@description": "security", "public": "security1r6t9kagaafun6zvkx4ysm2kh9xswca6x79dlu4lvmg6hynywx7nsvpgple", "private": { "encoding": "jwe", "content": "eyJhbGciOiJQQkVTMi1IUzUxMitBMjU2S1ciLCJjdHkiOiJqd2sranNvbiIsImVuYyI6IkEyNTZHQ00iLCJwMmMiOjUwMDAwMSwicDJzIjoiTlVVNVlWZDZTRVpJVldoMFNFSnNaUSJ9.pqZ9kim7OzW6lVLmPf4wXRYx8IvHmZi7ChzxmWqtGHo2zHeyp3Bhqw.x76wqFYsB-E-E0ov.1Adrme-LS8tC05n1D3FLUSiDGCMcf30lRjWDCB2CSh-3x4K2fZ2gibsvtp7aO4IjxkESnrUV6vCCAtXDa2I4f-aYAYzl1CkgSw-1JulQmVjl4l3NTcI189icJT0HxJ7-F0SGtpmTU1bGoGR9z_ERVErom3I6bSAl2OV4WcDVTfmyXBoJqM-hXYtIeIpLC0B4sxi3CFPhFQlEHF65AYwC2QgZb2qoP-tLnJG1FA.g-hH5zr7ksKhWS2aXAWP0Q"}}`)
	publicOnly       = []byte(`{"@apiVersion": "harp.elastic.co/v1", "@kind": "ContainerIdentity", "@timestamp": "2021-11-15T11:58:13.662568Z", "@description": "security", "public": "security1r6t9kagaafun6zvkx4ysm2kh9xswca6x79dlu4lvmg6hynywx7nsvpgple"}`)
)

func TestCodec_New(t *testing.T) {
	t.Run("invalid description", func(t *testing.T) {
		id, pub, err := New(rand.Reader, "é")
		assert.Error(t, err)
		assert.Nil(t, pub)
		assert.Nil(t, id)
	})

	t.Run("large description", func(t *testing.T) {
		id, pub, err := New(rand.Reader, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
		assert.Error(t, err)
		assert.Nil(t, pub)
		assert.Nil(t, id)
	})

	t.Run("invalid random source", func(t *testing.T) {
		id, pub, err := New(bytes.NewBuffer(nil), "test")
		assert.Error(t, err)
		assert.Nil(t, pub)
		assert.Nil(t, id)
	})

	t.Run("valid", func(t *testing.T) {
		id, pub, err := New(bytes.NewBuffer([]byte("deterministic-random-source-for-test-0001")), "security")
		assert.NoError(t, err)
		assert.NotNil(t, pub)
		assert.NotNil(t, id)
		assert.Equal(t, "harp.elastic.co/v1", id.APIVersion)
		assert.Equal(t, "security", id.Description)
		assert.Equal(t, "ContainerIdentity", id.Kind)
		assert.Equal(t, "security1mqtkctl32wy695wryccfgrdw4hr8cn9smk9vduc9yy5l3dfwr69swl0vee", id.Public)
		assert.Nil(t, id.Private)
		assert.False(t, id.HasPrivateKey())
	})
}

func TestCodec_FromReader(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		id, err := FromReader(nil)
		assert.Error(t, err)
		assert.Nil(t, id)
	})

	t.Run("empty", func(t *testing.T) {
		id, err := FromReader(bytes.NewReader([]byte("{}")))
		assert.Error(t, err)
		assert.Nil(t, id)
	})

	t.Run("invalid json", func(t *testing.T) {
		id, err := FromReader(bytes.NewReader([]byte("{")))
		assert.Error(t, err)
		assert.Nil(t, id)
	})

	t.Run("public key only", func(t *testing.T) {
		id, err := FromReader(bytes.NewReader(publicOnly))
		assert.Error(t, err)
		assert.Nil(t, id)
	})

	t.Run("valid", func(t *testing.T) {
		id, err := FromReader(bytes.NewReader(securityIdentity))
		assert.NoError(t, err)
		assert.NotNil(t, id)
	})
}

func TestPublicKeysFromIdentities(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		identities := []string{}
		publicKeys, err := SealingKeys(identities...)
		assert.Error(t, err)
		assert.Nil(t, publicKeys)
	})

	t.Run("invalid bech32 encoding", func(t *testing.T) {
		identities := []string{
			"recovery1hytnx4qpta252s5s7wypzq7rp3puks38vd8p4x9nhysvfylyjl/q^^wvr6",
		}
		publicKeys, err := SealingKeys(identities...)
		assert.Error(t, err)
		assert.Nil(t, publicKeys)
	})

	t.Run("valid", func(t *testing.T) {
		identities := []string{
			"recovery1hytnx4qpta252s5s7wypzq7rp3puks38vd8p4x9nhysvfylyjlwscswvr6",
			"security1eervzac2v26wxehktnf6lq0vpegrqcpg4uv4uxdr5vpzzmtdpflqttlghh",
		}
		publicKeys, err := SealingKeys(identities...)
		assert.NoError(t, err)
		assert.NotEmpty(t, publicKeys)
		assert.Equal(t, [32]byte{0xc1, 0x48, 0x9f, 0x5a, 0x41, 0xc7, 0x32, 0xa1, 0x3, 0xc1, 0x65, 0x9e, 0xeb, 0xc1, 0x95, 0x47, 0xd6, 0xea, 0x53, 0x7f, 0xf5, 0x48, 0x2d, 0x61, 0xa0, 0x60, 0x81, 0xe2, 0xe9, 0x37, 0x8, 0x2}, *publicKeys[0])
		assert.Equal(t, [32]byte{0x74, 0x70, 0x5d, 0xdc, 0x92, 0xa0, 0x95, 0x8b, 0xa6, 0x45, 0xfd, 0x52, 0xe0, 0x10, 0x69, 0x71, 0x9f, 0x92, 0x5d, 0xdf, 0x7d, 0x86, 0x6b, 0xf7, 0x20, 0x80, 0xfa, 0xd4, 0x5c, 0x59, 0x70, 0x70}, *publicKeys[1])
	})

	t.Run("valid - dedup", func(t *testing.T) {
		identities := []string{
			"recovery1hytnx4qpta252s5s7wypzq7rp3puks38vd8p4x9nhysvfylyjlwscswvr6",
			"recovery1hytnx4qpta252s5s7wypzq7rp3puks38vd8p4x9nhysvfylyjlwscswvr6",
			"security1eervzac2v26wxehktnf6lq0vpegrqcpg4uv4uxdr5vpzzmtdpflqttlghh",
		}
		publicKeys, err := SealingKeys(identities...)
		assert.NoError(t, err)
		assert.NotEmpty(t, publicKeys)
		assert.Equal(t, [32]byte{0xc1, 0x48, 0x9f, 0x5a, 0x41, 0xc7, 0x32, 0xa1, 0x3, 0xc1, 0x65, 0x9e, 0xeb, 0xc1, 0x95, 0x47, 0xd6, 0xea, 0x53, 0x7f, 0xf5, 0x48, 0x2d, 0x61, 0xa0, 0x60, 0x81, 0xe2, 0xe9, 0x37, 0x8, 0x2}, *publicKeys[0])
		assert.Equal(t, [32]byte{0x74, 0x70, 0x5d, 0xdc, 0x92, 0xa0, 0x95, 0x8b, 0xa6, 0x45, 0xfd, 0x52, 0xe0, 0x10, 0x69, 0x71, 0x9f, 0x92, 0x5d, 0xdf, 0x7d, 0x86, 0x6b, 0xf7, 0x20, 0x80, 0xfa, 0xd4, 0x5c, 0x59, 0x70, 0x70}, *publicKeys[1])
	})
}
