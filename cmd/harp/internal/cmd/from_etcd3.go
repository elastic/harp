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
	"time"

	"github.com/spf13/cobra"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/kv/etcd3"
	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/sdk/tlsconfig"
	"github.com/elastic/harp/pkg/tasks/from"
)

// -----------------------------------------------------------------------------

type fromEtcd3Params struct {
	outputPath           string
	basePaths            []string
	lastPathItemAsSecret bool

	endpoints   []string
	dialTimeout time.Duration
	username    string
	password    string

	useTLS             bool
	caFile             string
	certFile           string
	keyFile            string
	passphrase         string
	insecureSkipVerify bool
}

var fromEtcd3Cmd = func() *cobra.Command {
	var params fromEtcd3Params

	cmd := &cobra.Command{
		Use:   "etcd3",
		Short: "Extract KV pairs from CoreOS Etcdv3 KV Store",
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-kv-from-etcdv3", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			runFromEtcd3(ctx, &params)
		},
	}

	// Add parameters
	cmd.Flags().StringVar(&params.outputPath, "out", "-", "Container output path ('-' for stdout)")
	cmd.Flags().StringSliceVar(&params.basePaths, "paths", []string{}, "Exported base paths")
	cmd.Flags().BoolVarP(&params.lastPathItemAsSecret, "last-path-item-as-secret-key", "k", false, "Use the last path element as secret key")

	cmd.Flags().StringArrayVar(&params.endpoints, "endpoints", []string{"http://localhost:2379"}, "Etcd cluster endpoints")
	cmd.Flags().DurationVar(&params.dialTimeout, "dial-timeout", 15*time.Second, "Etcd cluster dial timeout")
	cmd.Flags().StringVar(&params.username, "username", "", "Etcd cluster connection username")
	cmd.Flags().StringVar(&params.password, "password", "", "Etcd cluster connection password")

	cmd.Flags().BoolVar(&params.useTLS, "tls", false, "Enable TLS")
	cmd.Flags().StringVar(&params.caFile, "ca-file", "", "TLS CA Certificate file path")
	cmd.Flags().StringVar(&params.certFile, "cert-file", "", "TLS Client certificate file path")
	cmd.Flags().StringVar(&params.keyFile, "key-file", "", "TLS Client private key file path")
	cmd.Flags().StringVar(&params.passphrase, "key-passphrase", "", "TLS Client private key passphrase")
	cmd.Flags().BoolVar(&params.insecureSkipVerify, "insecure-skip-verify", false, "Disable TLS certificate verification")

	return cmd
}

func runFromEtcd3(ctx context.Context, params *fromEtcd3Params) {
	// Create config
	config := clientv3.Config{
		Context:     ctx,
		Endpoints:   params.endpoints,
		DialTimeout: params.dialTimeout,
		Username:    params.username,
		Password:    params.password,
	}

	if params.useTLS {
		tlsConfig, err := tlsconfig.Client(&tlsconfig.Options{
			InsecureSkipVerify: params.insecureSkipVerify,
			CAFile:             params.caFile,
			CertFile:           params.certFile,
			KeyFile:            params.keyFile,
			Passphrase:         params.passphrase,
		})
		if err != nil {
			log.For(ctx).Fatal("unable to initialize TLS settings", zap.Error(err))
			return
		}

		// Assign TLS settings
		config.TLS = tlsConfig
	}

	// Creates a new client
	client, err := clientv3.New(config)
	if err != nil {
		log.For(ctx).Fatal("unable to connect to etcdv3 cluster", zap.Error(err))
		return
	}

	// Delegate to task
	t := &from.ExtractKVTask{
		Store:                   etcd3.Store(client),
		ContainerWriter:         cmdutil.FileWriter(params.outputPath),
		BasePaths:               params.basePaths,
		LastPathItemAsSecretKey: params.lastPathItemAsSecret,
	}

	// Run the task
	if err := t.Run(ctx); err != nil {
		log.For(ctx).Fatal("unable to execute kv extract task", zap.Error(err))
		return
	}
}
