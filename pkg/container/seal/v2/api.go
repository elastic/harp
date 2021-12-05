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

package v2

import (
	"crypto/elliptic"

	"github.com/elastic/harp/pkg/container/seal"
)

const (
	SealVersion = 2
)

const (
	containerSealedContentType = "application/vnd.harp.v1.SealedContainer"
	seedSize                   = 32
	publicKeySize              = 49
	privateKeySize             = 48
	encryptionKeySize          = 32
	nonceSize                  = 16
	macSize                    = 48
	signatureSize              = 96
	messageLimit               = 64 * 1024 * 1024
)

var (
	encryptionCurve = elliptic.P384()
	signatureCurve  = elliptic.P384()
)

// -----------------------------------------------------------------------------

func New() seal.Strategy {
	return &adapter{}
}

type adapter struct{}
