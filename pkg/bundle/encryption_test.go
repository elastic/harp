// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.
package bundle

import (
	"context"
	"fmt"
	"testing"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle/secret"
	"github.com/elastic/harp/pkg/sdk/value"
)

// -----------------------------------------------------------------------------
var _ value.Transformer = (*mockedTransformer)(nil)

type mockedTransformer struct {
	err error
}

func (m *mockedTransformer) To(ctx context.Context, input []byte) ([]byte, error) {
	return input, m.err
}
func (m *mockedTransformer) From(ctx context.Context, input []byte) ([]byte, error) {
	return input, m.err
}

// -----------------------------------------------------------------------------
func TestPartialLock(t *testing.T) {
	type args struct {
		ctx            context.Context
		b              *bundlev1.Bundle
		transformerMap map[string]value.Transformer
		skipUnresolved bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Nil bundle",
			wantErr: true,
		},
		{
			name: "Nil transformer map",
			args: args{
				b:              &bundlev1.Bundle{},
				transformerMap: nil,
			},
			wantErr: true,
		},
		{
			name: "no annotation found",
			args: args{
				b: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{
						{
							Name: "test/app/encrypted",
							Secrets: &bundlev1.SecretChain{
								Data: []*bundlev1.KV{
									{
										Key:   "test",
										Value: secret.MustPack("value"),
									},
								},
							},
						},
					},
				},
				transformerMap: map[string]value.Transformer{
					"test": &mockedTransformer{},
				},
			},
			wantErr: false,
		},
		{
			name: "no key alias found",
			args: args{
				b: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{
						{
							Name: "test/app/encrypted",
							Annotations: map[string]string{
								packageEncryptionAnnotation: "test",
							},
							Secrets: &bundlev1.SecretChain{
								Data: []*bundlev1.KV{
									{
										Key:   "test",
										Value: secret.MustPack("value"),
									},
								},
							},
						},
					},
				},
				transformerMap: map[string]value.Transformer{
					"invalid": &mockedTransformer{},
				},
			},
			wantErr: true,
		},
		{
			name: "alias point to nil transformer",
			args: args{
				b: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{
						{
							Name: "test/app/encrypted",
							Annotations: map[string]string{
								packageEncryptionAnnotation: "test",
							},
							Secrets: &bundlev1.SecretChain{
								Data: []*bundlev1.KV{
									{
										Key:   "test",
										Value: secret.MustPack("value"),
									},
								},
							},
						},
					},
				},
				transformerMap: map[string]value.Transformer{
					"test": nil,
				},
			},
			wantErr: true,
		},
		{
			name: "bundle secret unpack error",
			args: args{
				b: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{
						{
							Name: "test/app/encrypted",
							Annotations: map[string]string{
								packageEncryptionAnnotation: "test",
							},
							Secrets: &bundlev1.SecretChain{
								Data: []*bundlev1.KV{
									{
										Key:   "test",
										Value: []byte("value"),
									},
								},
							},
						},
					},
				},
				transformerMap: map[string]value.Transformer{
					"test": &mockedTransformer{},
				},
			},
			wantErr: true,
		},
		{
			name: "transformer error",
			args: args{
				b: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{
						{
							Name: "test/app/encrypted",
							Annotations: map[string]string{
								packageEncryptionAnnotation: "test",
							},
							Secrets: &bundlev1.SecretChain{
								Data: []*bundlev1.KV{
									{
										Key:   "test",
										Value: secret.MustPack("value"),
									},
								},
							},
						},
					},
				},
				transformerMap: map[string]value.Transformer{
					"test": &mockedTransformer{
						err: fmt.Errorf("test"),
					},
				},
			},
			wantErr: true,
		},
		// ---------------------------------------------------------------------
		{
			name: "valid",
			args: args{
				b: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{
						{
							Name: "test/app/encrypted",
							Annotations: map[string]string{
								packageEncryptionAnnotation: "test",
							},
							Secrets: &bundlev1.SecretChain{
								Data: []*bundlev1.KV{
									{
										Key:   "test",
										Value: secret.MustPack("value"),
									},
								},
							},
						},
					},
				},
				transformerMap: map[string]value.Transformer{
					"test": &mockedTransformer{},
				},
			},
			wantErr: false,
		},
		{
			name: "no key alias found with skip",
			args: args{
				b: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{
						{
							Name: "test/app/encrypted",
							Annotations: map[string]string{
								packageEncryptionAnnotation: "test",
							},
							Secrets: &bundlev1.SecretChain{
								Data: []*bundlev1.KV{
									{
										Key:   "test",
										Value: secret.MustPack("value"),
									},
								},
							},
						},
					},
				},
				transformerMap: map[string]value.Transformer{
					"invalid": &mockedTransformer{},
				},
				skipUnresolved: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := PartialLock(tt.args.ctx, tt.args.b, tt.args.transformerMap, tt.args.skipUnresolved); (err != nil) != tt.wantErr {
				t.Errorf("PartialLock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
