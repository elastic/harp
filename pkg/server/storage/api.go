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

package storage

import (
	"context"
	"errors"
	"net/url"
)

// ErrSecretNotFound is raised when trying to access non-existing secret.
var ErrSecretNotFound = errors.New("engine: secret not found")

// EngineFactoryFunc is the storage engine factory contract.
type EngineFactoryFunc func(*url.URL) (Engine, error)

//go:generate mockgen -destination test/mock/engine.gen.go -package mock github.com/elastic/harp/pkg/server/storage Engine

// Engine represents storage engine contract
type Engine interface {
	Get(ctx context.Context, id string) ([]byte, error)
}
