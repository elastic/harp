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

package vault

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/google/wire"
	"github.com/gosimple/slug"
	"go.uber.org/zap"

	"github.com/elastic/harp/cmd/harp-server/internal/config"
	"github.com/elastic/harp/cmd/harp-server/internal/dispatchers/vault/routes"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/sdk/tlsconfig"
	"github.com/elastic/harp/pkg/sdk/value"
	"github.com/elastic/harp/pkg/sdk/value/encryption"
	"github.com/elastic/harp/pkg/server/manager"
	"github.com/elastic/harp/pkg/server/storage/backends/container"
	vpath "github.com/elastic/harp/pkg/vault/path"
)

func backendManager(ctx context.Context, cfg *config.Configuration) (manager.Backend, error) {
	// Initialize default manager
	bm := manager.Default()

	// Backends
	for _, b := range cfg.Backends {
		// Register namespace engine
		if err := bm.Register(ctx, vpath.SanitizePath(b.NS), b.URL); err != nil {
			return nil, err
		}
	}

	// No error
	return bm, nil
}

type transformerMap map[string]value.Transformer

func transformers(cfg *config.Configuration) (transformerMap, error) {
	res := transformerMap{}

	if len(cfg.Transformers) == 0 {
		return res, nil
	}

	for _, tr := range cfg.Transformers {
		// Try to initialize the transformer from key
		t, err := encryption.FromKey(tr.Key)
		if err != nil {
			return res, fmt.Errorf("unable to initialize '%s' transformer: %w", tr.Name, err)
		}

		// Add to transfromer map
		res[slug.Make(tr.Name)] = t
	}

	// No error
	return res, nil
}

func httpServer(ctx context.Context, cfg *config.Configuration, bm manager.Backend, tm transformerMap) (*http.Server, error) {
	r := chi.NewRouter()

	// middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	// timeout before request cancelation
	r.Use(middleware.Timeout(60 * time.Second))

	routes.RootHandler(r)
	routes.KVHandler(r, bm)

	// Map transit handlers
	for name, t := range tm {
		routes.TransitHandler(r, name, t)
	}

	// Apply container keyring
	container.SetKeyring(cfg.Keyring)

	// Assign router to server
	server := &http.Server{
		Handler: r,
	}

	// Enable TLS if requested
	if cfg.Vault.UseTLS {
		// Client authentication enabled but not required
		clientAuth := tls.VerifyClientCertIfGiven
		if cfg.Vault.TLS.ClientAuthenticationRequired {
			clientAuth = tls.RequireAndVerifyClientCert
		}

		// Generate TLS configuration
		tlsConfig, err := tlsconfig.Server(&tlsconfig.Options{
			KeyFile:    cfg.Vault.TLS.PrivateKeyPath,
			CertFile:   cfg.Vault.TLS.CertificatePath,
			CAFile:     cfg.Vault.TLS.CACertificatePath,
			ClientAuth: clientAuth,
		})
		if err != nil {
			log.For(ctx).Error("Unable to build TLS configuration from settings", zap.Error(err))
			return nil, err
		}

		// Create the TLS credentials
		server.TLSConfig = tlsConfig
	} else {
		log.For(ctx).Info("No transport encryption enabled for fake Vault server")
	}

	// Return result
	return server, nil
}

// -----------------------------------------------------------------------------

func setup(ctx context.Context, cfg *config.Configuration) (*http.Server, error) {
	wire.Build(
		backendManager,
		transformers,
		httpServer,
	)
	return &http.Server{}, nil
}
