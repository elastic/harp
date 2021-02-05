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

package compare

import (
	"testing"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle/secret"
	"github.com/google/go-cmp/cmp"
	fuzz "github.com/google/gofuzz"
)

func MustPack(value interface{}) []byte {
	out, err := secret.Pack(value)
	if err != nil {
		panic(err)
	}

	return out
}

func TestDiff(t *testing.T) {
	type args struct {
		src *bundlev1.Bundle
		dst *bundlev1.Bundle
	}
	tests := []struct {
		name    string
		args    args
		want    []DiffItem
		wantErr bool
	}{
		{
			name:    "src nil",
			wantErr: true,
		},
		{
			name: "dst nil",
			args: args{
				src: &bundlev1.Bundle{},
			},
			wantErr: true,
		},
		{
			name: "identic",
			args: args{
				src: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{},
				},
				dst: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{},
				},
			},
			wantErr: false,
			want:    []DiffItem{},
		},
		{
			name: "new package",
			args: args{
				src: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{},
				},
				dst: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{
						{
							Name: "app/test",
							Secrets: &bundlev1.SecretChain{
								Data: []*bundlev1.KV{
									{
										Key:   "key1",
										Value: MustPack("payload"),
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
			want: []DiffItem{
				{Operation: Add, Type: "package", Path: "app/test"},
				{Operation: Add, Type: "secret", Path: "app/test#key1", Value: "payload"},
			},
		},
		{
			name: "package removed",
			args: args{
				src: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{
						{
							Name: "app/test",
							Secrets: &bundlev1.SecretChain{
								Data: []*bundlev1.KV{
									{
										Key:   "key1",
										Value: MustPack("payload"),
									},
								},
							},
						},
					},
				},
				dst: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{},
				},
			},
			wantErr: false,
			want: []DiffItem{
				{Operation: Remove, Type: "package", Path: "app/test"},
			},
		},
		{
			name: "secret added",
			args: args{
				src: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{
						{
							Name: "app/test",
							Secrets: &bundlev1.SecretChain{
								Data: []*bundlev1.KV{
									{
										Key:   "key1",
										Value: MustPack("payload"),
									},
								},
							},
						},
					},
				},
				dst: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{
						{
							Name: "app/test",
							Secrets: &bundlev1.SecretChain{
								Data: []*bundlev1.KV{
									{
										Key:   "key1",
										Value: MustPack("payload"),
									},
									{
										Key:   "key2",
										Value: MustPack("newpayload"),
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
			want: []DiffItem{
				{Operation: Add, Type: "secret", Path: "app/test#key2", Value: "newpayload"},
			},
		},
		{
			name: "secret removed",
			args: args{
				src: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{
						{
							Name: "app/test",
							Secrets: &bundlev1.SecretChain{
								Data: []*bundlev1.KV{
									{
										Key:   "key1",
										Value: MustPack("payload"),
									},
									{
										Key:   "key2",
										Value: MustPack("newpayload"),
									},
								},
							},
						},
					},
				},
				dst: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{
						{
							Name: "app/test",
							Secrets: &bundlev1.SecretChain{
								Data: []*bundlev1.KV{
									{
										Key:   "key2",
										Value: MustPack("newpayload"),
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
			want: []DiffItem{
				{Operation: Remove, Type: "secret", Path: "app/test#key1"},
			},
		},
		{
			name: "secret replaced",
			args: args{
				src: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{
						{
							Name: "app/test",
							Secrets: &bundlev1.SecretChain{
								Data: []*bundlev1.KV{
									{
										Key:   "key2",
										Value: MustPack("oldpayload"),
									},
								},
							},
						},
					},
				},
				dst: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{
						{
							Name: "app/test",
							Secrets: &bundlev1.SecretChain{
								Data: []*bundlev1.KV{
									{
										Key:   "key2",
										Value: MustPack("newpayload"),
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
			want: []DiffItem{
				{Operation: Replace, Type: "secret", Path: "app/test#key2", Value: "newpayload"},
			},
		},
		{
			name: "no-op",
			args: args{
				src: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{
						{
							Name: "app/test",
							Secrets: &bundlev1.SecretChain{
								Data: []*bundlev1.KV{
									{
										Key:   "key2",
										Value: MustPack("oldpayload"),
									},
								},
							},
						},
					},
				},
				dst: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{
						{
							Name: "app/test",
							Secrets: &bundlev1.SecretChain{
								Data: []*bundlev1.KV{
									{
										Key:   "key2",
										Value: MustPack("oldpayload"),
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
			want:    []DiffItem{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Diff(tt.args.src, tt.args.dst)
			if (err != nil) != tt.wantErr {
				t.Errorf("Diff() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("%q. Diff():\n-got/+want\ndiff %s", tt.name, diff)
			}
		})
	}
}

func TestDiff_Fuzz(t *testing.T) {
	// Making sure the descrption never panics
	for i := 0; i < 50; i++ {
		f := fuzz.New()

		var (
			src bundlev1.Bundle
			dst bundlev1.Bundle
		)

		// Prepare arguments
		f.Fuzz(&src)
		f.Fuzz(&dst)

		// Execute
		Diff(&src, &dst)
	}
}
