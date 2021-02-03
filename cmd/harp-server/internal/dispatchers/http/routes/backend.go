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
	"errors"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/sdk/value/encryption"
	"github.com/elastic/harp/pkg/server/storage"
)

// Backend returns a backend http request handler.
func backend(namespace string, engine storage.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			ctx    = r.Context()
			id     = r.URL.Path
			keyRaw = r.URL.Query().Get("key")
		)

		// Remove namespace prefix
		identifier := strings.TrimPrefix(id, fmt.Sprintf("/%s", namespace))

		// Retrieve secret from engine
		secret, err := engine.Get(ctx, identifier)
		if errors.Is(err, storage.ErrSecretNotFound) {
			http.Error(w, "secret not found", http.StatusNotFound)
			return
		}
		if err != nil {
			log.For(ctx).Error("unable to retrieve secret from engine", zap.Error(err), zap.String("url", r.URL.String()))
			http.Error(w, "unable to retrieve secret", http.StatusBadRequest)
			return
		}

		// key is defined
		if keyRaw != "" {
			// Retrieve transformer from key
			transformer, err := encryption.FromKey(keyRaw)
			if err != nil {
				log.For(ctx).Error("unable to initialize secret transformer", zap.String("url", r.URL.String()))
				http.Error(w, "unable to initialize secret transformer", http.StatusInternalServerError)
				return
			}

			// Apply transformation to secret value
			secret, err = transformer.To(ctx, secret)
			if err != nil {
				log.For(ctx).Error("unable to protect secret", zap.String("url", r.URL.String()))
				http.Error(w, "unable to protect secret", http.StatusBadRequest)
				return
			}
		}

		// Send result
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", secret)
	}
}
