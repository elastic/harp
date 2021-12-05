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

package v1

import (
	"crypto/ed25519"

	"github.com/elastic/harp/pkg/container/seal"
)

const (
	SealVersion = 1
)

const (
	containerSealedContentType = "application/vnd.harp.v1.SealedContainer"
	publicKeySize              = 32
	privateKeySize             = 32
	encryptionKeySize          = 32
	keyIdentifierSize          = 32
	nonceSize                  = 24
	signatureSize              = ed25519.SignatureSize
	messageLimit               = 64 * 1024 * 1024

	staticSignatureNonce      = "harp_container_psigk_box"
	signatureDomainSeparation = "harp encrypted signature"
)

// -----------------------------------------------------------------------------

func New() seal.Strategy {
	return &adapter{}
}

type adapter struct{}
