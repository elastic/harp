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

package seal

import (
	"io"

	"github.com/awnumar/memguard"

	containerv1 "github.com/elastic/harp/api/gen/go/harp/container/v1"
)

// Strategy describes the sealing/unsealing contract.
type Strategy interface {
	// GenerateKey create a key pair used as container identifier.
	GenerateKey(...GenerateOption) (publicKey, privateKey string, err error)
	// Seal the given container using the implemented algorithm.
	Seal(io.Reader, *containerv1.Container, ...string) (*containerv1.Container, error)
	// Unseal the given container using the given identity.
	Unseal(*containerv1.Container, *memguard.LockedBuffer) (*containerv1.Container, error)
}

// GenerateOptions represents container key generation options.
type GenerateOptions struct {
	DCKDMasterKey *memguard.LockedBuffer
	DCKDTarget    string
	RandomSource  io.Reader
}

// GenerateOption represents functional pattern builder for optional parameters.
type GenerateOption func(o *GenerateOptions)

// WithDeterministicKey enables deterministic container key generation.
func WithDeterministicKey(masterKey *memguard.LockedBuffer, target string) GenerateOption {
	return func(o *GenerateOptions) {
		o.DCKDMasterKey = masterKey
		o.DCKDTarget = target
	}
}

// WithRandom provides the random source for key generation.
func WithRandom(random io.Reader) GenerateOption {
	return func(o *GenerateOptions) {
		o.RandomSource = random
	}
}
