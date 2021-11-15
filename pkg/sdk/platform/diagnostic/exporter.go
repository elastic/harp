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

package diagnostic

import (
	"context"
	"net/http"
	"net/http/pprof"

	"github.com/google/gops/agent"
	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/sdk/log"
)

// Register adds diagnostic tools to main process
func Register(ctx context.Context, conf *Config, r *http.ServeMux) (func(), error) {
	if conf.GOPS.Enabled {
		// Start diagnostic handler
		if conf.GOPS.RemoteURL != "" {
			log.For(ctx).Info("Starting gops agent", zap.String("url", conf.GOPS.RemoteURL))
			if err := agent.Listen(agent.Options{Addr: conf.GOPS.RemoteURL}); err != nil {
				log.For(ctx).Error("Error on starting gops agent", zap.Error(err))
			}
		} else {
			log.For(ctx).Info("Starting gops agent locally")
			if err := agent.Listen(agent.Options{}); err != nil {
				log.For(ctx).Error("Error on starting gops agent locally", zap.Error(err))
			}
		}
	}

	if conf.PProf.Enabled {
		r.HandleFunc("/debug/pprof", pprof.Index)
		r.HandleFunc("/debug/cmdline", pprof.Cmdline)
		r.HandleFunc("/debug/profile", pprof.Profile)
		r.HandleFunc("/debug/symbol", pprof.Symbol)
		r.HandleFunc("/debug/trace", pprof.Trace)
		r.Handle("/debug/goroutine", pprof.Handler("goroutine"))
		r.Handle("/debug/heap", pprof.Handler("heap"))
		r.Handle("/debug/threadcreate", pprof.Handler("threadcreate"))
		r.Handle("/debug/block", pprof.Handler("block"))
	}

	// No error
	return func() {
		agent.Close()
	}, nil
}
