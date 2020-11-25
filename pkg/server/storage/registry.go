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
	"errors"
	"fmt"
	"net/url"
	"sync"
)

var (
	once    sync.Once
	engines map[string]EngineFactoryFunc

	// ErrEngineAlreadyRegistered is raise dwhen trying to register an existent engine.
	ErrEngineAlreadyRegistered = errors.New("engine: storage engine already registered")
)

func init() {
	once.Do(func() {
		engines = map[string]EngineFactoryFunc{}
	})
}

// Register a new storage engine.
func Register(name string, factory EngineFactoryFunc) error {
	if _, ok := engines[name]; !ok {
		engines[name] = factory

		// No error
		return nil
	}

	return ErrEngineAlreadyRegistered
}

// MustRegister try to register the engine and panic on error.
func MustRegister(name string, factory EngineFactoryFunc) {
	if err := Register(name, factory); err != nil {
		panic(err)
	}
}

// Build an engine instance with given URL.
func Build(uri string) (Engine, error) {
	// Parse URL
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("engine: unable to instantiate a storage engine: %w", err)
	}

	// Retrieve a factory accorting to url scheme
	factory, ok := engines[u.Scheme]
	if !ok {
		return nil, fmt.Errorf("engine: unable to find registered storage engine '%s'", u.Scheme)
	}

	// Delegate to backend factory
	engine, err := factory(u)
	if err != nil {
		return nil, err
	}

	// No error
	return engine, nil
}
