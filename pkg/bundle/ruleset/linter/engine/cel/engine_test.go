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

package cel

import (
	"context"
	"testing"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle/ruleset/linter/engine"
	"github.com/elastic/harp/pkg/bundle/secret"
)

func TestNew(t *testing.T) {
	type args struct {
		expressions []string
	}
	tests := []struct {
		name    string
		args    args
		want    engine.PackageLinter
		wantErr bool
	}{
		{
			name: "empty",
			args: args{
				expressions: []string{},
			},
			wantErr: false,
		},
		{
			name: "funcs",
			args: args{
				expressions: []string{
					`p.match_path("app/production/test")`,
					`p.has_secret("test") && p.secret("test").is_base64()`,
					`p.has_all_secrets(["test","test2"])`,
					`p.is_cso_compliant()`,
				},
			},
			wantErr: false,
		},
		{
			name: "not a boolean result",
			args: args{
				expressions: []string{
					`""`,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.args.expressions)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func mustPack(in interface{}) []byte {
	out, err := secret.Pack(in)
	if err != nil {
		panic(err)
	}
	return out
}

func Test_ruleEngine_EvaluatePackage(t *testing.T) {
	type fields struct {
		expressions []string
	}
	type args struct {
		ctx context.Context
		p   *bundlev1.Package
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid: match_path strict",
			fields: fields{
				expressions: []string{
					`p.match_path("app/production/test")`,
				},
			},
			args: args{
				p: &bundlev1.Package{
					Name: "app/production/test",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid: match_path strict",
			fields: fields{
				expressions: []string{
					`p.match_path("app/staging/test")`,
				},
			},
			args: args{
				p: &bundlev1.Package{
					Name: "app/production/test",
				},
			},
			wantErr: true,
		},
		{
			name: "valid: match_path regex",
			fields: fields{
				expressions: []string{
					`p.match_path("app/{production,staging}/test")`,
				},
			},
			args: args{
				p: &bundlev1.Package{
					Name: "app/production/test",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid: match_path regex",
			fields: fields{
				expressions: []string{
					`p.match_path("app/(production|staging)/test")`,
				},
			},
			args: args{
				p: &bundlev1.Package{
					Name: "app/qa/test",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid: has_secret - no secret data",
			fields: fields{
				expressions: []string{
					`p.has_secret("test")`,
				},
			},
			args: args{
				p: &bundlev1.Package{
					Name: "app/qa/test",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid: has_secret - secret not found",
			fields: fields{
				expressions: []string{
					`p.has_secret("test")`,
				},
			},
			args: args{
				p: &bundlev1.Package{
					Name: "app/qa/test",
					Secrets: &bundlev1.SecretChain{
						Data: []*bundlev1.KV{
							{
								Key: "test2",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid: has_secret",
			fields: fields{
				expressions: []string{
					`p.has_secret("test")`,
				},
			},
			args: args{
				p: &bundlev1.Package{
					Name: "app/qa/test",
					Secrets: &bundlev1.SecretChain{
						Data: []*bundlev1.KV{
							{
								Key: "test",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid: has_all_secret - secret not found",
			fields: fields{
				expressions: []string{
					`p.has_all_secrets(["test"])`,
				},
			},
			args: args{
				p: &bundlev1.Package{
					Name: "app/qa/test",
					Secrets: &bundlev1.SecretChain{
						Data: []*bundlev1.KV{
							{
								Key: "test2",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid: has_all_secret",
			fields: fields{
				expressions: []string{
					`p.has_all_secrets(["test"])`,
				},
			},
			args: args{
				p: &bundlev1.Package{
					Name: "app/qa/test",
					Secrets: &bundlev1.SecretChain{
						Data: []*bundlev1.KV{
							{
								Key: "test",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid: is_cso_compliant",
			fields: fields{
				expressions: []string{
					`p.is_cso_compliant()`,
				},
			},
			args: args{
				p: &bundlev1.Package{
					Name: "app/qa/security",
				},
			},
			wantErr: true,
		},
		{
			name: "valid: is_cso_compliant",
			fields: fields{
				expressions: []string{
					`p.is_cso_compliant()`,
				},
			},
			args: args{
				p: &bundlev1.Package{
					Name: "app/qa/security/harp/v1.0.0/server/database/credentials",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid: is_base64",
			fields: fields{
				expressions: []string{
					`p.has_secret("test") && p.secret("test").is_base64()`,
				},
			},
			args: args{
				p: &bundlev1.Package{
					Name: "app/qa/security",
					Secrets: &bundlev1.SecretChain{
						Data: []*bundlev1.KV{
							{
								Key: "test",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid: is_base64",
			fields: fields{
				expressions: []string{
					`p.has_secret("test") && p.secret("test").is_base64()`,
				},
			},
			args: args{
				p: &bundlev1.Package{
					Name: "app/qa/security/harp/v1.0.0/server/database/credentials",
					Secrets: &bundlev1.SecretChain{
						Data: []*bundlev1.KV{
							{
								Key:   "test",
								Value: mustPack(""),
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, _ := New(tt.fields.expressions)
			if err := re.EvaluatePackage(tt.args.ctx, tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("ruleEngine.EvaluatePackage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
