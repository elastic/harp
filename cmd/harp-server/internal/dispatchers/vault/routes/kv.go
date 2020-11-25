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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dchest/uniuri"
	"github.com/go-chi/chi"
	"github.com/gosimple/slug"
	"go.uber.org/zap"

	"github.com/elastic/harp/build/version"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/server/manager"
	"github.com/elastic/harp/pkg/server/storage"
	vpath "github.com/elastic/harp/pkg/vault/path"
)

// KVHandler initializes Vault KV API handler for given bundle
func KVHandler(bm manager.Backend) http.Handler {
	r := chi.NewRouter()

	// Initialize controler
	ctrl := &vaultKVHandler{
		bm: bm,
	}

	// Map routes
	r.Get("/v1/sys/seal-status", ctrl.sealStatus())
	r.Get("/v1/sys/leader", ctrl.leaderStatus())
	r.Put("/v1/auth/token/renew-self", ctrl.selfRenew())
	r.Get("/v1/secret/config", ctrl.getConfig())
	r.Get("/v1/sys/internal/ui/mounts/*", ctrl.getMount())
	r.Get("/v1/secret/data/*", ctrl.getSecret())

	return r
}

// KV is an alias to map for readability.
type KV map[string]interface{}

type vaultKVHandler struct {
	bm manager.Backend
}

func (h *vaultKVHandler) sealStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		with(w, r, http.StatusOK, &KV{
			"type":         "shamir",
			"initialized":  true,
			"sealed":       false,
			"t":            1,
			"n":            1,
			"progress":     0,
			"version":      version.Version,
			"cluster_name": "harp-container-server",
			"cluster_id":   "763d1163-18f9-46d8-b1ca-2d327c0cc57f",
			"nonce":        "",
		})
	}
}

func (h *vaultKVHandler) leaderStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		with(w, r, http.StatusOK, &KV{
			"ha_enabled": false,
			"is_self":    true,
		})
	}
}

func (h *vaultKVHandler) selfRenew() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		with(w, r, http.StatusOK, &KV{
			"auth": &KV{
				"client_token": fmt.Sprintf("harp.%s", uniuri.NewLen(12)),
				"policies":     []string{"harp", "read-only"},
				"metadata": &KV{
					"user": "harp",
				},
				"lease_duration": 3600,
				"renewable":      true,
			},
		})
	}
}

func (h *vaultKVHandler) getMount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		with(w, r, http.StatusOK, &KV{
			"data": &KV{
				"type": "kv",
				"path": "secret/",
				"options": &KV{
					"version": "2",
				},
			},
		})
	}
}

func (h *vaultKVHandler) getConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		with(w, r, http.StatusOK, &KV{
			"data": &KV{
				"max_versions": "0",
			},
		})
	}
}

func (h *vaultKVHandler) getSecret() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get namespace from headers
		ns := slug.Make(r.Header.Get("X-Vault-Namespace"))
		if ns == "" {
			ns = "root"
		}

		// Extract path
		p := strings.TrimPrefix(r.URL.Path, "/v1/secret/data")

		// Retrieve secret from engine
		secret, err := h.bm.GetSecret(ctx, vpath.SanitizePath(ns), p)
		if err == storage.ErrSecretNotFound {
			http.Error(w, "secret not found", http.StatusNotFound)
			return
		}
		if err != nil {
			log.For(ctx).Error("unable to retrieve secret from engine", zap.Error(err), zap.String("url", r.URL.String()))
			http.Error(w, "unable to retrieve secret", http.StatusBadRequest)
			return
		}

		// Decode secret as JSON
		var data interface{}
		if err := json.Unmarshal(secret, &data); err != nil {
			log.For(ctx).Error("unable to decode secret from engine", zap.Error(err), zap.String("url", r.URL.String()))
			http.Error(w, "unable to decode secret", http.StatusBadRequest)
			return
		}

		// Send response
		with(w, r, http.StatusOK, &KV{
			"data": &KV{
				"data": data,
			},
			"metadata": &KV{
				"version": "1",
			},
		})
	}
}
