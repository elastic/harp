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
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/vault/api"
)

func TestBuilder_V1(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		switch r.URL.Path {
		case "/v1/sys/internal/ui/mounts/application":
			w.WriteHeader(200)
			fmt.Fprintf(w, `{"data":{"type":"kv", "path":"application/", "options":{"version": "1"}}}`)
		case "/v1/application/secret/foo":
			switch r.Method {
			case http.MethodGet:
				if q.Get("list") == "true" {
					w.WriteHeader(200)
					fmt.Fprintf(w, `{"data":{"keys":[]}}`)
				} else {
					w.WriteHeader(200)
					fmt.Fprintf(w, `{"data":{"key":"value"}}`)
				}
			case http.MethodPut:
				w.WriteHeader(200)
				fmt.Fprintf(w, `{"data":{}}`)
			case "LIST":
				w.WriteHeader(200)
				fmt.Fprintf(w, `{"data":{"keys":[]}}`)
			default:
				w.WriteHeader(400)
			}

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
	underTest, err := New(vaultClient, "application")
	if err != nil {
		t.Errorf("BuilderV1() error = %v", err)
		return
	}

	// Read the value
	_, _, err = underTest.Read(context.Background(), "application/secret/foo")
	if err != nil {
		t.Errorf("BuilderV1() - Read error = %v", err)
		return
	}

	// List secrets
	_, err = underTest.List(context.Background(), "application/secret/foo")
	if err != nil {
		t.Errorf("BuilderV1() - List error = %v", err)
		return
	}

	// Write secrets
	if err := underTest.Write(context.Background(), "application/secret/foo", map[string]interface{}{"key": "value"}); err != nil {
		t.Errorf("BuilderV1() - List error = %v", err)
		return
	}
}

func TestBuilder_V2(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/sys/internal/ui/mounts/application":
			w.WriteHeader(200)
			fmt.Fprintf(w, `{"data":{"type":"kv", "path":"application/", "options":{"version": "2"}}}`)
		case "/v1/application/data/secret/foo":
			switch r.Method {
			case http.MethodGet:
				w.WriteHeader(200)
				fmt.Fprintf(w, `{"data":{"data":{"key":"value"},"metadata":{"created_time": "2018-03-22T02:24:06.945319214Z","deletion_time": "","destroyed": false,"version": 2}}}`)
			case http.MethodPut:
				w.WriteHeader(200)
				fmt.Fprintf(w, `{"data":{"data":{}}}`)
			default:
				w.WriteHeader(400)
			}
		case "/v1/application/metadata/secret/foo":
			switch r.Method {
			case http.MethodGet:
				w.WriteHeader(200)
				fmt.Fprintf(w, `{"data":{"keys":[]}}`)
			default:
				w.WriteHeader(400)
			}
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
	underTest, err := New(vaultClient, "application")
	if err != nil {
		t.Errorf("BuilderV2() error = %v", err)
		return
	}

	// Read the value
	_, _, err = underTest.Read(context.Background(), "application/secret/foo")
	if err != nil {
		t.Errorf("BuilderV2() - Read error = %v", err)
		return
	}

	// List secrets
	_, err = underTest.List(context.Background(), "application/secret/foo")
	if err != nil {
		t.Errorf("BuilderV2() - List error = %v", err)
		return
	}

	// Write secrets
	if err := underTest.Write(context.Background(), "application/secret/foo", map[string]interface{}{"key": "value"}); err != nil {
		t.Errorf("BuilderV2() - List error = %v", err)
		return
	}
}
