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
	"fmt"
	"net/http"

	"github.com/dchest/uniuri"
	"github.com/go-chi/chi"

	"github.com/elastic/harp/build/version"
)

// RootHandler initializes Vault KV API handler for given bundle
func RootHandler(r chi.Router) {
	// Initialize controler
	ctrl := &vaultRootHandler{}

	// Map routes
	r.Get("/v1/sys/seal-status", ctrl.sealStatus())
	r.Get("/v1/sys/leader", ctrl.leaderStatus())
	r.Put("/v1/auth/token/renew-self", ctrl.selfRenew())
}

type vaultRootHandler struct{}

func (h *vaultRootHandler) sealStatus() http.HandlerFunc {
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

func (h *vaultRootHandler) leaderStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		with(w, r, http.StatusOK, &KV{
			"ha_enabled": false,
			"is_self":    true,
		})
	}
}

func (h *vaultRootHandler) selfRenew() http.HandlerFunc {
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
