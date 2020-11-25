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

	"github.com/oklog/run"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/elastic/harp/build/version"
	"github.com/elastic/harp/cmd/harp-server/internal/dispatchers/grpc"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/sdk/platform"
)

var grpcNamespaces []string

// -----------------------------------------------------------------------------

var grpcCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grpc",
		Short: "Starts a gRPC container server",
		Run:   runGRPCServer,
	}

	// Parameters
	cmd.Flags().StringSliceVarP(&grpcNamespaces, "namespace", "n", nil, "namespace mapping (ns:url)")
	log.CheckErr("unable to mark 'namespace' flag as required.", cmd.MarkFlagRequired("namespace"))

	return cmd
}

func runGRPCServer(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	// Initialize config
	initConfig()

	// Starting banner
	log.For(ctx).Info("Starting harp gRPC bundle server ...")

	// Start goroutine group
	errServe := platform.Serve(ctx, &platform.Server{
		Debug:           conf.Debug.Enable,
		Name:            "harp-server-grpc",
		Version:         version.Version,
		Revision:        version.Revision,
		Instrumentation: conf.Instrumentation,
		Network:         conf.GRPC.Network,
		Address:         conf.GRPC.Listen,
		Builder: func(ln net.Listener, group *run.Group) {
			// Override config
			if err := overrideBackendConfig(conf, grpcNamespaces); err != nil {
				log.For(ctx).Fatal("Unable to parse backend mapping", zap.Error(err))
			}

			server, err := grpc.New(ctx, conf)
			if err != nil {
				log.For(ctx).Fatal("Unable to start gRPC server", zap.Error(err))
			}

			group.Add(
				func() error {
					log.For(ctx).Info("Starting gRPC server", zap.Stringer("address", ln.Addr()))
					return server.Serve(ln)
				},
				func(e error) {
					log.For(ctx).Info("Shutting gRPC server down")
					server.GracefulStop()
				},
			)
		},
	})
	log.CheckErrCtx(ctx, "Unable to run application", errServe)
}
