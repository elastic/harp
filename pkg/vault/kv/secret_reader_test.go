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

package kv

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/vault/api"
)

func TestSecretReader_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/sys/internal/ui/mounts/application/secret/not/found":
			w.WriteHeader(200)
			fmt.Fprintf(w, `{"data":{"type":"kv", "path":"application/", "options":{"version": "2"}}}`)
		case "/v1/application/data/secret/not/found":
			w.WriteHeader(404)
			fmt.Fprintf(w, `{}`)
		default:
			w.WriteHeader(400)
		}
	}))
	defer server.Close()

	// Initialize Vault client
	vaultClient, err := api.NewClient(&api.Config{
		Address:    server.URL,
		Timeout:    time.Second * 1,
		MaxRetries: 1,
		HttpClient: &http.Client{Transport: cleanhttp.DefaultTransport(), Timeout: time.Second * 2},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Build reader
	underTest := SecretGetter(vaultClient)

	_, err = underTest("application/secret/not/found")
	if err != nil && !errors.Is(err, ErrPathNotFound) {
		t.Errorf("SecretReader() error = %v, expected %v", err, ErrPathNotFound)
	}
}

func TestSecretReader_Found(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/sys/internal/ui/mounts/application/secret/found":
			w.WriteHeader(200)
			fmt.Fprintf(w, `{"data":{"type":"kv", "path":"application/", "options":{"version": "2"}}}`)
		case "/v1/application/data/secret/found":
			w.WriteHeader(200)
			fmt.Fprintf(w, `{"data":{"data":{"key":"value"},"metadata":{}}}`)
		default:
			w.WriteHeader(400)
		}
	}))
	defer server.Close()

	// Initialize Vault client
	vaultClient, err := api.NewClient(&api.Config{
		Address:    server.URL,
		Timeout:    time.Second * 1,
		MaxRetries: 1,
		HttpClient: &http.Client{Transport: cleanhttp.DefaultTransport(), Timeout: time.Second * 2},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Build reader
	underTest := SecretGetter(vaultClient)

	res, err := underTest("application/secret/found")
	if err != nil {
		t.Errorf("SecretReader() error = %v, expected nil", err)
	}
	expectedRes := map[string]interface{}{
		"key": "value",
	}
	if !reflect.DeepEqual(res, expectedRes) {
		t.Errorf("SecretReader() got %v, expected %v", res, expectedRes)
	}
}
