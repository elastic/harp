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

	"github.com/elastic/harp/pkg/container/identity/key"
	"github.com/stretchr/testify/assert"
)

var (
	v1SecurityIdentity = []byte(`{"@apiVersion":"harp.elastic.co/v1","@kind":"ContainerIdentity","@timestamp":"2021-12-01T22:15:11.144249Z","@description":"security","public":"v1.ipk.7u8B1VFrHyMeWyt8Jzj1Nj2BgVB7z-umD8R-OOnJahE","private":{"content":"ZXlKaGJHY2lPaUpRUWtWVE1pMUlVelV4TWl0Qk1qVTJTMWNpTENKbGJtTWlPaUpCTWpVMlIwTk5JaXdpY0RKaklqbzFNREF3TURFc0luQXljeUk2SW1obU1UVnpUVmRwTmtaMmNUSlVZM0p5TVhReVZtY2lmUS5ZOGtfVXR2dWtmcVRxTE5fQ2l5ajdTejU2dThOYV9uMG1FTG5jMHFCQ1d0MkVqX2VHRk80RmcuN0p2ekhGYkZrRXdXWGxOeC5ycVJLSno1ZWFGajRqSl9wOVAzUVBuUUs3dXhkWUhBOUNIZFUxTEswWkQ3Q2dickJzUDFRRFRTRU1lX3lqbTZVQ1dpNzFUVmxfX3JISVdSR3VDVVpWSE1KMXNtRnR5c2UzdHBURkdnZFRCaVQxTmw4dWlNZ2JiUEN1cHJ4Uy0wUjRGU2dobXFLU0s3TGhRcUxFWFVaNFF0SVliMDd6Y19vMnRZNlVnU3NMaFBlSUFPM1M2WlBwQXFYU3lfSjR3NzEzdFhEU1ZTX2ZuOFJ5MlF2NTJmOHg0cXBiN0Q3NGlTTndOb052Zy5rcHVzTTVoZ21RT3JhS1luNGxTVjZ3"},"signature":"Kq1OJlAOexIvt9TXETYeYGotqqCz8PiqFEYuSbHmJPVBqtYpI2Q_zNE0fO5hs-JdTqG3p6oLiITHK9cYyx2hBw"}`)
	publicOnly         = []byte(`{"@apiVersion":"harp.elastic.co/v1","@kind":"ContainerIdentity","@timestamp":"2021-12-01T20:56:30.832199Z","@description":"security","public":"v1.ipk.PRdbQ8qbrDsfTLA-aeQIdUF0VwnauvWqQF-CXNFp9oM"}`)
	v2SecurityIdentity = []byte(`{"@apiVersion":"harp.elastic.co/v1","@kind":"ContainerIdentity","@timestamp":"2021-12-01T22:15:07.586373Z","@description":"security","public":"v2.ipk.AkLr_HHMO5Loy2bK42mvCADrJ7s2PSYCRTnqDWJV8PCK2EXmu-GTV8HmNJwmA8IJ8Q","private":{"content":"ZXlKaGJHY2lPaUpRUWtWVE1pMUlVelV4TWl0Qk1qVTJTMWNpTENKbGJtTWlPaUpCTWpVMlIwTk5JaXdpY0RKaklqbzFNREF3TURFc0luQXljeUk2SWpCMFozUk5OMm80TmxacFNrWXpjemhYWDBsb05VRWlmUS5QXzVVMTdSR3JSRVg4UHoyNGpRQkQzdGROWGU3ci1UMVh2bnBnT21aMkwweGZXQXNfT2dWcVEuYUpuekZCQTNBWllXSld2TC53SVAtZlRERjN5R0NaRGtldThOM3A1NUZPRF9ZX3QzSV9ubHN2MDVqcWNLdlJLczFfWjVfM2Zhc2Z0cU0tMlRoN25VdDZIaXZWLVB1ckVIQ2hBRENHaF9SZTBySVVwZkV4OHBCcUk4V3BIYTdSYktUTUN3RmNpSDMzeTQxZ1duT1lpN1R1TmJBamhNMjZMdDZZMFN0ekcyRi1FUm9jSWotWklwMDJwcGZjdUpKOU91S1BDOThKTl9ZV3EzcVA2TW55Ym1WTnFFZ1hwdWFVZm9GcTN3ZWlSX2paVkNsRzU5cTBGdWplVHN0UnRzU2xuZFlndTVBTl9LanFWRmluNDBXNGcxZWRMdWZDM1U0UGZhZVMzUlQxSS0wRkVnN0ZGMnE0QVdINy01aF9IQWg4WFR4eXBCTjR3THE1TTd6ZExRLlkzTTlBeDc1bGNYbmNNaGNxV3dOMXc"},"signature":"dpbnMGAPvFbHSjEXs1GMyO8Kmw9cZqTOKI5wAA1ApcO1RXtFGS_GyC1zAtuFDhhVmTWdFzS4HdVg0LEhxBivbqsr_cft_9CR-7uVUPpkb2Hz2d4BkL3yzDo9bkLfllaM"}`)
)

func TestCodec_New(t *testing.T) {
	t.Run("invalid description", func(t *testing.T) {
		id, pub, err := New(rand.Reader, "é", key.Ed25519)
		assert.Error(t, err)
		assert.Nil(t, pub)
		assert.Nil(t, id)
	})

	t.Run("ed25519 - invalid random source", func(t *testing.T) {
		id, pub, err := New(bytes.NewBuffer(nil), "test", key.Ed25519)
		assert.Error(t, err)
		assert.Nil(t, pub)
		assert.Nil(t, id)
	})

	t.Run("p384 - invalid random source", func(t *testing.T) {
		id, pub, err := New(bytes.NewBuffer(nil), "test", key.P384)
		assert.Error(t, err)
		assert.Nil(t, pub)
		assert.Nil(t, id)
	})

	t.Run("legacy - invalid random source", func(t *testing.T) {
		id, pub, err := New(bytes.NewBuffer(nil), "test", key.Legacy)
		assert.Error(t, err)
		assert.Nil(t, pub)
		assert.Nil(t, id)
	})

	t.Run("valid - ed25519", func(t *testing.T) {
		id, pub, err := New(bytes.NewBuffer([]byte("deterministic-random-source-for-test-0001")), "security", key.Ed25519)
		assert.NoError(t, err)
		assert.NotNil(t, pub)
		assert.NotNil(t, id)
		assert.Equal(t, "harp.elastic.co/v1", id.APIVersion)
		assert.Equal(t, "security", id.Description)
		assert.Equal(t, "ContainerIdentity", id.Kind)
		assert.Equal(t, "v1.ipk.2BdsL_FTiaLRwyYwlA2urcZ8TLDdisbzBSEp-LUuHos", id.Public)
		assert.Nil(t, id.Private)
		assert.False(t, id.HasPrivateKey())
	})

	t.Run("valid - p-384", func(t *testing.T) {
		id, pub, err := New(bytes.NewBuffer([]byte("deterministic-random-source-for-test-0001-1ioQiLEbVCm1Y7NfWCf6oNWoV6p5E4spJgRXKQHdV44XcNvqywMnIYYcL8qZ4Wk")), "security", key.P384)
		assert.NoError(t, err)
		assert.NotNil(t, pub)
		assert.NotNil(t, id)
		assert.Equal(t, "harp.elastic.co/v1", id.APIVersion)
		assert.Equal(t, "security", id.Description)
		assert.Equal(t, "ContainerIdentity", id.Kind)
		assert.Equal(t, "v2.ipk.A0X20rlE8Pqp-YoMG8SNOop918AyfoSF_R9Z7MF5vP5nUoc_ZSRWauQR6cL4DqgrRA", id.Public)
		assert.Nil(t, id.Private)
		assert.False(t, id.HasPrivateKey())
	})

	t.Run("valid - legacy", func(t *testing.T) {
		id, pub, err := New(bytes.NewBuffer([]byte("deterministic-random-source-for-test-0001")), "security", key.Legacy)
		assert.NoError(t, err)
		assert.NotNil(t, pub)
		assert.NotNil(t, id)
		assert.Equal(t, "harp.elastic.co/v1", id.APIVersion)
		assert.Equal(t, "security", id.Description)
		assert.Equal(t, "ContainerIdentity", id.Kind)
		assert.Equal(t, "ZxTKWxgrG341_FxatkkfAxedMtfz1zJzAm6FUmitxHM", id.Public)
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

	t.Run("valid - v1", func(t *testing.T) {
		id, err := FromReader(bytes.NewReader(v1SecurityIdentity))
		assert.NoError(t, err)
		assert.NotNil(t, id)
	})

	t.Run("valid - v2", func(t *testing.T) {
		id, err := FromReader(bytes.NewReader(v2SecurityIdentity))
		assert.NoError(t, err)
		assert.NotNil(t, id)
	})
}
