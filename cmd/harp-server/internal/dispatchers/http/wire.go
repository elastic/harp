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

//+build wireinject

package http

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/google/wire"
	"go.uber.org/zap"

	"github.com/elastic/harp/cmd/harp-server/internal/config"
	"github.com/elastic/harp/cmd/harp-server/internal/dispatchers/http/routes"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/sdk/tlsconfig"
	"github.com/elastic/harp/pkg/server/manager"
	"github.com/elastic/harp/pkg/server/storage/backends/container"
)

func backendManager(ctx context.Context, cfg *config.Configuration) (manager.Backend, error) {
	// Initialize default manager
	bm := manager.Default()

	// Backends
	for _, b := range cfg.Backends {
		// Register namespace engine
		if err := bm.Register(ctx, b.NS, b.URL); err != nil {
			return nil, err
		}
	}

	// No error
	return bm, nil
}

func httpServer(ctx context.Context, cfg *config.Configuration, bm manager.Backend) (*http.Server, error) {
	r := chi.NewRouter()

	// middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	// timeout before request cancelation
	r.Use(middleware.Timeout(60 * time.Second))

	// Apply container keyring
	container.SetKeyring(cfg.Keyring)

	// API endpoint
	backendRouter, err := routes.Backends(ctx, cfg, bm)
	if err != nil {
		return nil, err
	}

	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/", http.StripPrefix("/api/v1", backendRouter))
	})

	// Assign router to server
	server := &http.Server{
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		Handler:           r,
	}

	// Enable TLS if requested
	if cfg.HTTP.UseTLS {
		// Client authentication enabled but not required
		clientAuth := tls.VerifyClientCertIfGiven
		if cfg.HTTP.TLS.ClientAuthenticationRequired {
			clientAuth = tls.RequireAndVerifyClientCert
		}

		// Generate TLS configuration
		tlsConfig, err := tlsconfig.Server(&tlsconfig.Options{
			KeyFile:    cfg.HTTP.TLS.PrivateKeyPath,
			CertFile:   cfg.HTTP.TLS.CertificatePath,
			CAFile:     cfg.HTTP.TLS.CACertificatePath,
			ClientAuth: clientAuth,
		})
		if err != nil {
			log.For(ctx).Error("Unable to build TLS configuration from settings", zap.Error(err))
			return nil, err
		}

		// Create the TLS credentials
		server.TLSConfig = tlsConfig
	} else {
		log.For(ctx).Info("No transport encryption enabled for HTTP server")
	}

	// Return result
	return server, nil
}

// -----------------------------------------------------------------------------

func setup(ctx context.Context, cfg *config.Configuration) (*http.Server, error) {
	wire.Build(
		backendManager,
		httpServer,
	)
	return &http.Server{}, nil
}
