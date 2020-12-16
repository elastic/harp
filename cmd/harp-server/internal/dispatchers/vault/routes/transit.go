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
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/gosimple/slug"

	"github.com/elastic/harp/pkg/sdk/value"
)

// TransitHandler initializes Vault Transit API handler for given transformer
func TransitHandler(r chi.Router, keyName string, tr value.Transformer) {
	// Initialize controler
	ctrl := &vaultTransitHandler{
		tr: tr,
	}

	// Map routes
	r.Post(fmt.Sprintf("/v1/transit/encrypt/%s", slug.Make(keyName)), ctrl.encryptData())
	r.Put(fmt.Sprintf("/v1/transit/encrypt/%s", slug.Make(keyName)), ctrl.encryptData())
	r.Post(fmt.Sprintf("/v1/transit/decrypt/%s", slug.Make(keyName)), ctrl.decryptData())
	r.Put(fmt.Sprintf("/v1/transit/decrypt/%s", slug.Make(keyName)), ctrl.decryptData())
}

type vaultTransitHandler struct {
	tr value.Transformer
}

func (h *vaultTransitHandler) encryptData() http.HandlerFunc {
	type request struct {
		PlainText string `json:"plaintext,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if err := decodeJSONBody(w, r, &req); err != nil {
			http.Error(w, "request is invalid", http.StatusBadRequest)
			return
		}

		// Check plaintext encoding
		rawPlainText, err := base64.StdEncoding.DecodeString(req.PlainText)
		if err != nil {
			http.Error(w, "plaintext must be a valid base64 encoded value", http.StatusBadRequest)
			return
		}

		// Encrypt plaintext with transformer
		cipherRaw, err := h.tr.To(r.Context(), rawPlainText)
		if err != nil {
			http.Error(w, "unable to encrypt plaintext", http.StatusBadRequest)
			return
		}

		// Return response
		with(w, r, http.StatusOK, &KV{
			"data": &KV{
				"ciphertext": fmt.Sprintf("vault:v1:%s", base64.StdEncoding.EncodeToString(cipherRaw)),
			},
		})
	}
}

func (h *vaultTransitHandler) decryptData() http.HandlerFunc {
	type request struct {
		CipherText string `json:"ciphertext,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if err := decodeJSONBody(w, r, &req); err != nil {
			http.Error(w, "request is invalid", http.StatusBadRequest)
			return
		}

		// Remove Vault prefix
		req.CipherText = strings.TrimPrefix(req.CipherText, "vault:v1:")

		// Check plaintext encoding
		rawCipherText, err := base64.StdEncoding.DecodeString(req.CipherText)
		if err != nil {
			http.Error(w, "ciphertext must be a valid base64 encoded value", http.StatusBadRequest)
			return
		}

		// Encrypt plaintext with transformer
		cipherRaw, err := h.tr.From(r.Context(), rawCipherText)
		if err != nil {
			http.Error(w, "unable to decrypt ciphertext", http.StatusBadRequest)
			return
		}

		// Return response
		with(w, r, http.StatusOK, &KV{
			"data": &KV{
				"plaintext": base64.StdEncoding.EncodeToString(cipherRaw),
			},
		})
	}
}
