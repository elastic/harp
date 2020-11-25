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
	"net"
	"time"

	"github.com/oklog/run"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/elastic/harp/build/version"
	"github.com/elastic/harp/cmd/harp-server/internal/dispatchers/vault"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/sdk/platform"
)

var vaultNamespaces []string

// -----------------------------------------------------------------------------

var vaultCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vault",
		Short: "Starts a Vault container server",
		Run:   runVaultServer,
	}

	// Parameters
	cmd.Flags().StringSliceVarP(&vaultNamespaces, "namespace", "n", nil, "namespace mapping (ns:url)")
	log.CheckErr("unable to mark 'namespace' flag as required.", cmd.MarkFlagRequired("namespace"))

	return cmd
}

func runVaultServer(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	// Initialize config
	initConfig()

	// Starting banner
	log.For(ctx).Info("Starting harp Vault bundle server ...")

	// Start goroutine group
	errServe := platform.Serve(ctx, &platform.Server{
		Debug:           conf.Debug.Enable,
		Name:            "harp-server-vault",
		Version:         version.Version,
		Revision:        version.Revision,
		Instrumentation: conf.Instrumentation,
		Network:         conf.Vault.Network,
		Address:         conf.Vault.Listen,
		Builder: func(ln net.Listener, group *run.Group) {
			// Override config
			if err := overrideBackendConfig(conf, vaultNamespaces); err != nil {
				log.For(ctx).Fatal("Unable to parse backend mapping", zap.Error(err))
			}

			server, err := vault.New(ctx, conf)
			if err != nil {
				log.For(ctx).Fatal("Unable to start Vault API server", zap.Error(err))
			}

			group.Add(
				func() error {
					if conf.Vault.UseTLS {
						log.For(ctx).Info("Starting Vault API HTTPS server", zap.Stringer("address", ln.Addr()))
						return server.ServeTLS(ln, conf.Vault.TLS.CertificatePath, conf.Vault.TLS.PrivateKeyPath)
					}

					log.For(ctx).Info("Starting Vault API HTTP server", zap.Stringer("address", ln.Addr()))
					return server.Serve(ln)
				},
				func(e error) {
					log.For(ctx).Info("Shutting Vault API server down")

					shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
					defer cancel()
					if err := server.Shutdown(shutdownCtx); err != nil {
						log.For(ctx).Fatal("Server Shutdown Failed", zap.Error(err))
					}
				},
			)
		},
	})
	log.CheckErrCtx(ctx, "Unable to run application", errServe)
}
