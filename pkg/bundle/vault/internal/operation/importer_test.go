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

package operation

import (
	"context"
	"testing"

	"github.com/hashicorp/vault/api"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
)

func Test_Importer(t *testing.T) {
	type args struct {
		client            *api.Client
		bundleFile        *bundlev1.Bundle
		prefix            string
		withMetadata      bool
		withVaultMetadata bool
		maxWorkerCount    int64
	}
	tests := []struct {
		name                    string
		args                    args
		wantWithMetadata        bool
		wantWithVaultMetadata   bool
		wantBackendsInitialized bool
	}{
		{
			name: "all false metadata",
			args: args{
				client:            &api.Client{},
				bundleFile:        &bundlev1.Bundle{},
				prefix:            "test",
				withMetadata:      false,
				withVaultMetadata: false,
				maxWorkerCount:    10,
			},
			wantWithMetadata:        false,
			wantWithVaultMetadata:   false,
			wantBackendsInitialized: true,
		},
		{
			name: "withMetadata true, withVaultMetadata false",
			args: args{
				client:            &api.Client{},
				bundleFile:        &bundlev1.Bundle{},
				prefix:            "test",
				withMetadata:      true,
				withVaultMetadata: false,
				maxWorkerCount:    10,
			},
			wantWithMetadata:        true,
			wantWithVaultMetadata:   false,
			wantBackendsInitialized: true,
		},
		{
			name: "withMetadata false, withVaultMetadata true",
			args: args{
				client:            &api.Client{},
				bundleFile:        &bundlev1.Bundle{},
				prefix:            "test",
				withMetadata:      false,
				withVaultMetadata: true,
				maxWorkerCount:    10,
			},
			wantWithMetadata:        true, // Should be true because of withVaultMetadata
			wantWithVaultMetadata:   true,
			wantBackendsInitialized: true,
		},
		{
			name: "both metadata true",
			args: args{
				client:            &api.Client{},
				bundleFile:        &bundlev1.Bundle{},
				prefix:            "test",
				withMetadata:      true,
				withVaultMetadata: true,
				maxWorkerCount:    10,
			},
			wantWithMetadata:        true,
			wantWithVaultMetadata:   true,
			wantBackendsInitialized: true,
		},
		{
			name: "empty prefix",
			args: args{
				client:            &api.Client{},
				bundleFile:        &bundlev1.Bundle{},
				prefix:            "",
				withMetadata:      false,
				withVaultMetadata: false,
				maxWorkerCount:    5,
			},
			wantWithMetadata:        false,
			wantWithVaultMetadata:   false,
			wantBackendsInitialized: true,
		},
		{
			name: "zero max worker count",
			args: args{
				client:            &api.Client{},
				bundleFile:        &bundlev1.Bundle{},
				prefix:            "vault",
				withMetadata:      false,
				withVaultMetadata: false,
				maxWorkerCount:    0,
			},
			wantWithMetadata:        false,
			wantWithVaultMetadata:   false,
			wantBackendsInitialized: true,
		},
		{
			name: "nil bundle",
			args: args{
				client:            &api.Client{},
				bundleFile:        nil,
				prefix:            "test",
				withMetadata:      false,
				withVaultMetadata: false,
				maxWorkerCount:    10,
			},
			wantWithMetadata:        false,
			wantWithVaultMetadata:   false,
			wantBackendsInitialized: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Importer(tt.args.client, tt.args.bundleFile, tt.args.prefix, tt.args.withMetadata, tt.args.withVaultMetadata, tt.args.maxWorkerCount)

			// Type assert to access internal fields
			imp, ok := got.(*importer)
			if !ok {
				t.Errorf("Importer() returned wrong type, expected *importer")
				return
			}

			// Check client
			if imp.client != tt.args.client {
				t.Errorf("Importer().client = %v, want %v", imp.client, tt.args.client)
			}

			// Check bundle
			if imp.bundle != tt.args.bundleFile {
				t.Errorf("Importer().bundle = %v, want %v", imp.bundle, tt.args.bundleFile)
			}

			// Check prefix
			if imp.prefix != tt.args.prefix {
				t.Errorf("Importer().prefix = %v, want %v", imp.prefix, tt.args.prefix)
			}

			// Check withMetadata (should be true if either withMetadata OR withVaultMetadata is true)
			if imp.withMetadata != tt.wantWithMetadata {
				t.Errorf("Importer().withMetadata = %v, want %v", imp.withMetadata, tt.wantWithMetadata)
			}

			// Check withVaultMetadata
			if imp.withVaultMetadata != tt.wantWithVaultMetadata {
				t.Errorf("Importer().withVaultMetadata = %v, want %v", imp.withVaultMetadata, tt.wantWithVaultMetadata)
			}

			// Check maxWorkerCount
			if imp.maxWorkerCount != tt.args.maxWorkerCount {
				t.Errorf("Importer().maxWorkerCount = %v, want %v", imp.maxWorkerCount, tt.args.maxWorkerCount)
			}

			// Check backends is initialized
			if tt.wantBackendsInitialized && imp.backends == nil {
				t.Errorf("Importer().backends should be initialized, got nil")
			}

			// Check backends is empty map
			if tt.wantBackendsInitialized && len(imp.backends) != 0 {
				t.Errorf("Importer().backends should be empty, got %d items", len(imp.backends))
			}
		})
	}
}

func Test_logBundleAnalysis(t *testing.T) {
	tests := []struct {
		name   string
		bundle *bundlev1.Bundle
		prefix string
	}{
		{
			name: "empty bundle",
			bundle: &bundlev1.Bundle{
				Packages: []*bundlev1.Package{},
			},
			prefix: "",
		},
		{
			name: "bundle with empty packages",
			bundle: &bundlev1.Bundle{
				Packages: []*bundlev1.Package{
					{
						Name:    "app/test/empty1",
						Secrets: nil,
					},
					{
						Name: "app/test/empty2",
						Secrets: &bundlev1.SecretChain{
							Data: []*bundlev1.KV{},
						},
					},
				},
			},
			prefix: "",
		},
		{
			name: "bundle with secrets",
			bundle: &bundlev1.Bundle{
				Packages: []*bundlev1.Package{
					{
						Name: "app/production/database",
						Secrets: &bundlev1.SecretChain{
							Data: []*bundlev1.KV{
								{Key: "username", Value: []byte("admin")},
								{Key: "password", Value: []byte("secret")},
							},
						},
					},
					{
						Name: "app/production/api",
						Secrets: &bundlev1.SecretChain{
							Data: []*bundlev1.KV{
								{Key: "api_key", Value: []byte("key123")},
							},
						},
					},
				},
			},
			prefix: "",
		},
		{
			name: "bundle with mixed packages",
			bundle: &bundlev1.Bundle{
				Packages: []*bundlev1.Package{
					{
						Name: "app/staging/database",
						Secrets: &bundlev1.SecretChain{
							Data: []*bundlev1.KV{
								{Key: "username", Value: []byte("user")},
							},
						},
					},
					{
						Name:    "app/staging/empty",
						Secrets: nil,
					},
					{
						Name: "app/staging/cache",
						Secrets: &bundlev1.SecretChain{
							Data: []*bundlev1.KV{
								{Key: "redis_url", Value: []byte("localhost")},
								{Key: "redis_port", Value: []byte("6379")},
								{Key: "redis_password", Value: []byte("pass")},
							},
						},
					},
				},
			},
			prefix: "",
		},
		{
			name: "bundle with prefix",
			bundle: &bundlev1.Bundle{
				Packages: []*bundlev1.Package{
					{
						Name: "database/config",
						Secrets: &bundlev1.SecretChain{
							Data: []*bundlev1.KV{
								{Key: "host", Value: []byte("localhost")},
							},
						},
					},
				},
			},
			prefix: "vault/secrets",
		},
		{
			name: "bundle with multiple backends",
			bundle: &bundlev1.Bundle{
				Packages: []*bundlev1.Package{
					{
						Name: "app/database",
						Secrets: &bundlev1.SecretChain{
							Data: []*bundlev1.KV{
								{Key: "key1", Value: []byte("val1")},
							},
						},
					},
					{
						Name: "infra/kubernetes",
						Secrets: &bundlev1.SecretChain{
							Data: []*bundlev1.KV{
								{Key: "key2", Value: []byte("val2")},
							},
						},
					},
					{
						Name: "platform/monitoring",
						Secrets: &bundlev1.SecretChain{
							Data: []*bundlev1.KV{
								{Key: "key3", Value: []byte("val3")},
							},
						},
					},
				},
			},
			prefix: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create importer with test bundle
			imp := &importer{
				bundle: tt.bundle,
				prefix: tt.prefix,
			}
			ctx := context.Background()
			imp.logBundleAnalysis(ctx)
		})
	}
}
