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

package routes

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/gosimple/slug"
	"go.uber.org/zap"

	"github.com/elastic/harp/cmd/harp-server/internal/config"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/server/manager"
)

// Backends returns an HTTP router for backends
func Backends(ctx context.Context, cfg *config.Configuration, bm manager.Backend) (http.Handler, error) {
	r := chi.NewRouter()

	// Backends
	for _, b := range cfg.Backends {
		// Retrieve backend engine
		engine, err := bm.GetNameSpace(ctx, b.NS)
		if err != nil {
			return nil, err
		}

		// Wrap engine with handler
		ns := clean(b.NS)
		r.Route(fmt.Sprintf("/%s", ns), func(r chi.Router) {
			r.Get("/*", backend(ns, engine))
		})

		log.For(ctx).Info("Bakend registered", zap.String("path", b.NS))
	}

	// Return no error
	return r, nil
}

func clean(ns string) string {
	// Remove any starting "/"
	ns = strings.TrimPrefix(ns, "/")
	// Slugify
	return slug.Make(ns)
}
