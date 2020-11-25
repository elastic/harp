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

// Package http contains http server definition.
package http

import (
	"context"
	"net/http"
	"sync"

	"github.com/elastic/harp/cmd/harp-server/internal/config"
)

type application struct {
	cfg    *config.Configuration
	server *http.Server
}

var (
	app  *application
	once sync.Once
)

// -----------------------------------------------------------------------------

// New initialize the application
func New(ctx context.Context, cfg *config.Configuration) (*http.Server, error) {
	var err error

	once.Do(func() {
		// Initialize application
		app = &application{
			cfg: cfg,
		}

		// Initialize core context
		app.server, err = setup(ctx, cfg)
	})

	// Return server
	return app.server, err
}
