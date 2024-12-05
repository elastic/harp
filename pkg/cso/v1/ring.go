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
	"fmt"
	"strings"
)

// Ring describes secret ring contract.
type Ring interface {
	Level() int
	Name() string
	Prefix() string
	Path(...string) (string, error)
}

var (
	// RingMeta represents R0 secrets
	RingMeta = &ring{
		level:  0,
		name:   "Meta",
		prefix: ringMeta,
		pathBuilderFunc: func(_ Ring, values ...string) (string, error) {
			return csoPath("meta/%s", 1, values...)
		},
	}
	// RingInfra represents R1 secrets
	RingInfra = &ring{
		level:  1,
		name:   "Infrastructure",
		prefix: ringInfra,
		pathBuilderFunc: func(_ Ring, values ...string) (string, error) {
			return csoPath("infra/%s/%s/%s/%s/%s", 5, values...)
		},
	}
	// RingPlatform represents R2 secrets
	RingPlatform = &ring{
		level:  2,
		name:   "Platform",
		prefix: ringPlatform,
		pathBuilderFunc: func(_ Ring, values ...string) (string, error) {
			return csoPath("platform/%s/%s/%s/%s/%s", 5, values...)
		},
	}
	// RingProduct represents R3 secrets
	RingProduct = &ring{
		level:  3,
		name:   "Product",
		prefix: ringProduct,
		pathBuilderFunc: func(_ Ring, values ...string) (string, error) {
			return csoPath("product/%s/%s/%s/%s", 4, values...)
		},
	}
	// RingApplication represents R4 secrets
	RingApplication = &ring{
		level:  4,
		name:   "Application",
		prefix: ringApp,
		pathBuilderFunc: func(_ Ring, values ...string) (string, error) {
			return csoPath("app/%s/%s/%s/%s/%s/%s", 6, values...)
		},
	}
	// RingArtifact represents R5 secrets
	RingArtifact = &ring{
		level:  5,
		name:   "Artifact",
		prefix: ringArtifact,
		pathBuilderFunc: func(_ Ring, values ...string) (string, error) {
			return csoPath("artifact/%s/%s/%s", 3, values...)
		},
	}
)

// -----------------------------------------------------------------------------

type ring struct {
	level           int
	name            string
	prefix          string
	pathBuilderFunc func(Ring, ...string) (string, error)
}

func (r ring) Level() int {
	return r.level
}

func (r ring) Name() string {
	return r.name
}

func (r ring) Prefix() string {
	return r.prefix
}

func (r ring) Path(values ...string) (string, error) {
	return r.pathBuilderFunc(r, values...)
}

// -----------------------------------------------------------------------------

// csoPath build and validate a secret path according to CSO specification
func csoPath(format string, count int, values ...string) (string, error) {
	// Check values count
	if len(values) < count {
		return "", fmt.Errorf("expected (%d) and received (%d) value count doesn't match", count, len(values))
	}

	// Prepare suffix
	suffix := strings.Join(values[count-1:], "/")

	// Prepare values
	var items []interface{}
	for i := 0; i < count-1; i++ {
		items = append(items, values[i])
	}
	items = append(items, suffix)

	// Prepare validation
	csoPath := fmt.Sprintf(format, items...)

	// Validate secret path
	if err := Validate(csoPath); err != nil {
		return "", fmt.Errorf("'%s' is not a compliant CSO path: %w", csoPath, err)
	}

	// No Error
	return csoPath, nil
}
