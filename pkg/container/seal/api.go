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

// Streategy describes the sealing/unsealing contract.
type Strategy interface {
	// PublicKeys return the appropriate key format used by the sealing strategy.
	PublicKeys(keys ...string) ([]interface{}, error)
	// Seal the given container using the implemented algorithm.
	Seal(io.Reader, *containerv1.Container, ...interface{}) (*containerv1.Container, error)
	// Unseal the given container using the given identity.
	Unseal(*containerv1.Container, *memguard.LockedBuffer) (*containerv1.Container, error)
}
