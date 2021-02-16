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

package patch

import (
	"os"
	"reflect"
	"testing"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	fuzz "github.com/google/gofuzz"
)

var (
	opt = cmp.FilterPath(
		func(p cmp.Path) bool {
			// Remove ignoring of the fields below once go-cmp is able to ignore generated fields.
			// See https://github.com/google/go-cmp/issues/153
			ignoreXXXCache :=
				p.String() == "XXX_sizecache" ||
					p.String() == "Packages.XXX_sizecache" ||
					p.String() == "Packages.Secrets.XXX_sizecache" ||
					p.String() == "Packages.Secrets.Data.XXX_sizecache"
			return ignoreXXXCache
		}, cmp.Ignore())

	ignoreOpts = []cmp.Option{
		cmpopts.IgnoreUnexported(bundlev1.Bundle{}),
		cmpopts.IgnoreUnexported(bundlev1.Package{}),
		cmpopts.IgnoreUnexported(bundlev1.SecretChain{}),
		cmpopts.IgnoreUnexported(bundlev1.KV{}),
		opt,
	}
)

func TestValidate(t *testing.T) {
	type args struct {
		spec *bundlev1.Patch
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "nil",
			wantErr: true,
		},
		{
			name: "invalid apiVersion",
			args: args{
				spec: &bundlev1.Patch{
					ApiVersion: "foo",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid kind",
			args: args{
				spec: &bundlev1.Patch{
					ApiVersion: "harp.elastic.co/v1",
					Kind:       "foo",
				},
			},
			wantErr: true,
		},
		{
			name: "nil meta",
			args: args{
				spec: &bundlev1.Patch{
					ApiVersion: "harp.elastic.co/v1",
					Kind:       "BundlePatch",
				},
			},
			wantErr: true,
		},
		{
			name: "meta name not defined",
			args: args{
				spec: &bundlev1.Patch{
					ApiVersion: "harp.elastic.co/v1",
					Kind:       "BundlePatch",
					Meta:       &bundlev1.PatchMeta{},
				},
			},
			wantErr: true,
		},
		{
			name: "nil spec",
			args: args{
				spec: &bundlev1.Patch{
					ApiVersion: "harp.elastic.co/v1",
					Kind:       "BundlePatch",
					Meta:       &bundlev1.PatchMeta{},
				},
			},
			wantErr: true,
		},
		{
			name: "no action patch",
			args: args{
				spec: &bundlev1.Patch{
					ApiVersion: "harp.elastic.co/v1",
					Kind:       "BundlePatch",
					Meta:       &bundlev1.PatchMeta{},
					Spec:       &bundlev1.PatchSpec{},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Validate(tt.args.spec); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestChecksum(t *testing.T) {
	type args struct {
		spec *bundlev1.Patch
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "nil",
			wantErr: true,
		},
		{
			name: "valid",
			args: args{
				spec: &bundlev1.Patch{
					ApiVersion: "harp.elastic.co/v1",
					Kind:       "BundlePatch",
					Meta:       &bundlev1.PatchMeta{},
					Spec:       &bundlev1.PatchSpec{},
				},
			},
			wantErr: false,
			want:    "BkFRGRHhouZLyiZe0CUyyZSlt_guk7tJonToaV4zOC4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Checksum(tt.args.spec)
			if (err != nil) != tt.wantErr {
				t.Errorf("Checksum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Checksum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApply_Fuzz(t *testing.T) {
	// Making sure the descrption never panics
	for i := 0; i < 500; i++ {
		f := fuzz.New()

		// Prepare arguments
		values := map[string]interface{}{}
		spec := &bundlev1.Patch{
			ApiVersion: "harp.elastic.co/v1",
			Kind:       "BundlePatch",
			Meta: &bundlev1.PatchMeta{
				Name: "test-patch",
			},
			Spec: &bundlev1.PatchSpec{
				Rules: []*bundlev1.PatchRule{
					{
						Package:  &bundlev1.PatchPackage{},
						Selector: &bundlev1.PatchSelector{},
					},
				},
			},
		}
		file := bundlev1.Bundle{
			Packages: []*bundlev1.Package{
				{
					Name: "foo",
					Secrets: &bundlev1.SecretChain{
						Data: []*bundlev1.KV{
							{
								Key:   "k1",
								Value: []byte("v1"),
							},
						},
					},
				},
			},
		}

		f.Fuzz(&spec)
		f.Fuzz(&file)

		// Execute
		Apply(spec, &file, values)
	}
}

func mustLoadPatch(filePath string) *bundlev1.Patch {
	f, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	p, err := YAML(f)
	if err != nil {
		panic(err)
	}

	return p
}

func TestApply(t *testing.T) {
	type args struct {
		spec   *bundlev1.Patch
		b      *bundlev1.Bundle
		values map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    *bundlev1.Bundle
		wantErr bool
	}{
		{
			name:    "nil",
			wantErr: true,
		},
		{
			name: "empty bundle",
			args: args{
				spec:   mustLoadPatch("../../../test/fixtures/patch/valid/path-cleaner.yaml"),
				b:      &bundlev1.Bundle{},
				values: map[string]interface{}{},
			},
			wantErr: false,
			want: &bundlev1.Bundle{
				Packages: []*bundlev1.Package{},
			},
		},
		{
			name: "modifiable bundle",
			args: args{
				spec: mustLoadPatch("../../../test/fixtures/patch/valid/path-cleaner.yaml"),
				b: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{
						{
							Name: "secrets/application/component-1.yaml",
						},
						{
							Name: "secrets/application/component-2.yaml",
						},
					},
				},
				values: map[string]interface{}{},
			},
			wantErr: false,
			want: &bundlev1.Bundle{
				Packages: []*bundlev1.Package{
					{
						Annotations: map[string]string{
							"patched":             "true",
							"secret-path-cleaner": "true",
						},
						Name: "application/component-1",
					},
					{
						Annotations: map[string]string{
							"patched":             "true",
							"secret-path-cleaner": "true",
						},
						Name: "application/component-2",
					},
				},
			},
		},
		{
			name: "duplicate package paths",
			args: args{
				spec: mustLoadPatch("../../../test/fixtures/patch/valid/path-cleaner.yaml"),
				b: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{
						{
							Name: "secrets/application/component-1.yaml",
						},
						{
							Name: "secrets/application/component-1.yaml",
						},
					},
				},
				values: map[string]interface{}{},
			},
			wantErr: false,
			want: &bundlev1.Bundle{
				Packages: []*bundlev1.Package{
					{
						Annotations: map[string]string{
							"patched":             "true",
							"secret-path-cleaner": "true",
						},
						Name: "application/component-1",
					},
					{
						Annotations: map[string]string{
							"patched":             "true",
							"secret-path-cleaner": "true",
						},
						Name: "application/component-1",
					},
				},
			},
		},
		{
			name: "remove package",
			args: args{
				spec: mustLoadPatch("../../../test/fixtures/patch/valid/remove-package.yaml"),
				b: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{
						{
							Name: "application/to-be-removed",
						},
						{
							Name: "secrets/application/component-2.yaml",
						},
					},
				},
				values: map[string]interface{}{},
			},
			wantErr: false,
			want: &bundlev1.Bundle{
				Packages: []*bundlev1.Package{
					{
						Name: "secrets/application/component-2.yaml",
					},
				},
			},
		},
		{
			name: "add package",
			args: args{
				spec: mustLoadPatch("../../../test/fixtures/patch/valid/add-package.yaml"),
				b: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{
						{
							Name: "secrets/application/component-2.yaml",
						},
					},
				},
				values: map[string]interface{}{},
			},
			wantErr: false,
			want: &bundlev1.Bundle{
				Packages: []*bundlev1.Package{
					{
						Name: "secrets/application/component-2.yaml",
					},
					{
						Name: "application/created-package",
						Annotations: map[string]string{
							"package-creator":                       "true",
							"patched":                               "true",
							"secret-service.elstc.co/encryptionKey": "UcbPlrEJ9jZEQX06n8oMln_mCl3EU2zl2ZVc-obb7Dw=",
						},
						Secrets: &bundlev1.SecretChain{
							Annotations: map[string]string{
								"secret-service.elstc.co/encryptionKey": "DrZ-0yEA18iS7A4xaR_pd-relh9KMtTw2q11nBEJykg=",
							},
							Data: []*bundlev1.KV{
								{
									Key:   "key",
									Type:  "string",
									Value: []byte("0\n\x02\x01\x01\x13\x05value"),
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Apply(tt.args.spec, tt.args.b, tt.args.values)
			if (err != nil) != tt.wantErr {
				t.Errorf("Apply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want, ignoreOpts...); diff != "" {
				t.Errorf("%q. Patch.Apply():\n-got/+want\ndiff %s", tt.name, diff)
			}
		})
	}
}
