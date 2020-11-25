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
	"strings"

	csov1 "github.com/elastic/harp/api/gen/go/cso/v1"
)

const (
	ringMeta     = "meta"
	ringInfra    = "infra"
	ringPlatform = "platform"
	ringProduct  = "product"
	ringApp      = "app"
	ringArtifact = "artifact"
)

// -----------------------------------------------------------------------------

var ringMapNames = strings.Split("invalid;unknown;meta;infra;platform;product;app;artifact", ";")

// ToRingName returns the ring level name
func ToRingName(lvl csov1.RingLevel) string {
	return ringMapNames[lvl]
}

// FromRingName returns the ring level object according to given name
func FromRingName(name string) csov1.RingLevel {
	var i int32

	// Search for value
	for idx, n := range ringMapNames {
		if strings.EqualFold(n, name) {
			i = int32(idx)
		}
	}

	return csov1.RingLevel(i)
}

// -----------------------------------------------------------------------------

var qualityMapNames = strings.Split("invalid;unknown;production;staging;qa;dev", ";")

// ToStageName return the stage name
func ToStageName(lvl csov1.QualityLevel) string {
	return qualityMapNames[lvl]
}

// FromStageName returns the stage level object from given name
func FromStageName(name string) csov1.QualityLevel {
	var i int32

	// Search for value
	for idx, n := range qualityMapNames {
		if strings.EqualFold(n, name) {
			i = int32(idx)
		}
	}

	return csov1.QualityLevel(i)
}
