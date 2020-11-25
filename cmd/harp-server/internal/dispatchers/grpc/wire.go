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

package grpc

import (
	"context"
	"crypto/tls"

	"github.com/google/wire"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/cmd/harp-server/internal/config"
	"github.com/elastic/harp/cmd/harp-server/internal/dispatchers/grpc/server"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/sdk/tlsconfig"
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

func grpcServer(ctx context.Context, cfg *config.Configuration, bm manager.Backend) (*grpc.Server, error) {

	// Apply container keyring
	container.SetKeyring(cfg.Keyring)

	// gRPC middlewares
	sopts := []grpc.ServerOption{}

	// Enable TLS if requested
	if cfg.GRPC.UseTLS {
		// Client authentication enabled but not required
		clientAuth := tls.VerifyClientCertIfGiven
		if cfg.GRPC.TLS.ClientAuthenticationRequired {
			clientAuth = tls.RequireAndVerifyClientCert
		}

		// Generate TLS configuration
		tlsConfig, err := tlsconfig.Server(&tlsconfig.Options{
			KeyFile:    cfg.GRPC.TLS.PrivateKeyPath,
			CertFile:   cfg.GRPC.TLS.CertificatePath,
			CAFile:     cfg.GRPC.TLS.CACertificatePath,
			ClientAuth: clientAuth,
		})
		if err != nil {
			log.For(ctx).Error("Unable to build TLS configuration from settings", zap.Error(err))
			return nil, err
		}

		// Create the TLS credentials
		sopts = append(sopts, grpc.Creds(credentials.NewTLS(tlsConfig)))
	} else {
		log.For(ctx).Info("No transport encryption enabled for gRPC server")
	}

	// Initialize the server
	grpcServer := grpc.NewServer(sopts...)

	// Health
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	// Register services
	bundlev1.RegisterBundleAPIServer(grpcServer, server.Bundle(bm))
	healthServer.SetServingStatus("harp.bundle.v1", healthpb.HealthCheckResponse_SERVING)

	// Reflection
	reflection.Register(grpcServer)

	// Return result
	return grpcServer, nil
}

// -----------------------------------------------------------------------------

func setup(ctx context.Context, cfg *config.Configuration) (*grpc.Server, error) {
	wire.Build(
		backendManager,
		grpcServer,
	)
	return &grpc.Server{}, nil
}
