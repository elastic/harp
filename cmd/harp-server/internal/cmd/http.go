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

package cmd

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/oklog/run"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/elastic/harp/build/version"
	"github.com/elastic/harp/cmd/harp-server/internal/config"
	"github.com/elastic/harp/cmd/harp-server/internal/dispatchers/http"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/sdk/platform"
)

var httpNamespaces []string

// -----------------------------------------------------------------------------

var httpCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "http",
		Short: "Starts an HTTP container server",
		Run:   runHTTPServer,
	}

	// Parameters
	cmd.Flags().StringSliceVarP(&httpNamespaces, "namespace", "n", nil, "namespace mapping (ns:url)")
	log.CheckErr("unable to mark 'namespace' flag as required.", cmd.MarkFlagRequired("namespace"))

	return cmd
}

func runHTTPServer(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	// Initialize config
	initConfig()

	// Starting banner
	log.For(ctx).Info("Starting harp HTTP bundle server ...")

	// Start goroutine group
	err := platform.Serve(ctx, &platform.Server{
		Debug:           conf.Debug.Enable,
		Name:            "harp-server-http",
		Version:         version.Version,
		Revision:        version.Revision,
		Instrumentation: conf.Instrumentation,
		Network:         conf.HTTP.Network,
		Address:         conf.HTTP.Listen,
		Builder: func(ln net.Listener, group *run.Group) {
			// Override config
			if err := overrideBackendConfig(conf, httpNamespaces); err != nil {
				log.For(ctx).Fatal("Unable to parse namespace mapping", zap.Error(err))
			}

			server, err := http.New(ctx, conf)
			if err != nil {
				log.For(ctx).Fatal("Unable to start HTTP server", zap.Error(err))
			}

			group.Add(
				func() error {
					if conf.HTTP.UseTLS {
						log.For(ctx).Info("Starting HTTPS server", zap.Stringer("address", ln.Addr()))
						return server.ServeTLS(ln, conf.HTTP.TLS.CertificatePath, conf.HTTP.TLS.PrivateKeyPath)
					}

					log.For(ctx).Info("Starting HTTP server", zap.Stringer("address", ln.Addr()))
					return server.Serve(ln)
				},
				func(e error) {
					log.For(ctx).Info("Shutting HTTP server down")

					shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
					defer cancel()
					if err := server.Shutdown(shutdownCtx); err != nil {
						log.For(ctx).Fatal("Server Shutdown Failed", zap.Error(err))
					}
				},
			)
		},
	})
	log.CheckErrCtx(ctx, "Unable to run application", err)
}

// -----------------------------------------------------------------------------

func overrideBackendConfig(cfg *config.Configuration, backends []string) error {
	// Parse backend mapping declaration
	for _, decl := range backends {
		parts := strings.SplitN(decl, ":", 2)

		// Mapping must have 2 parts
		if len(parts) != 2 {
			return fmt.Errorf("unable to parse backend declaration, invalid part count")
		}

		ns := parts[0]
		engineURL := parts[1]

		log.Bg().Debug("Backend override", zap.String("ns", ns), zap.String("url", engineURL))

		// Initialize backend list
		if cfg.Backends == nil {
			cfg.Backends = []config.Backend{}
		}

		// Add to backend
		cfg.Backends = append(cfg.Backends, config.Backend{
			NS:  ns,
			URL: engineURL,
		})
	}

	return nil
}
