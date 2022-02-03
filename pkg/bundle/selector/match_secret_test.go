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

package selector

import (
	"regexp"
	"testing"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	fuzz "github.com/google/gofuzz"
)

func Test_matchSecret_IsSatisfiedBy(t *testing.T) {
	type fields struct {
		strict string
		regex  *regexp.Regexp
	}
	type args struct {
		object interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "nil",
			want: false,
		},
		{
			name: "empty",
			args: args{},
			want: false,
		},
		{
			name: "not supported type",
			args: args{
				object: struct{}{},
			},
			want: false,
		},
		{
			name: "supported type: path nil",
			args: args{
				object: &bundlev1.Package{},
			},
			want: false,
		},
		{
			name: "supported type: path empty",
			args: args{
				object: &bundlev1.Package{
					Name: "",
				},
			},
			want: false,
		},
		{
			name: "supported type: strict mode not match",
			fields: fields{
				strict: "bar",
			},
			args: args{
				object: &bundlev1.Package{
					Name: "foo",
				},
			},
			want: false,
		},
		{
			name: "supported type: strict mode with match",
			fields: fields{
				strict: "foo",
			},
			args: args{
				object: &bundlev1.Package{
					Name: "foo",
					Secrets: &bundlev1.SecretChain{
						Data: []*bundlev1.KV{
							{
								Key: "foo",
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "supported type: regexp mode not match",
			fields: fields{
				regex: regexp.MustCompile("bar"),
			},
			args: args{
				object: &bundlev1.Package{
					Name: "foo",
					Secrets: &bundlev1.SecretChain{
						Data: []*bundlev1.KV{
							{
								Key: "foo",
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "supported type: regexp mode with match",
			fields: fields{
				regex: regexp.MustCompile("foo"),
			},
			args: args{
				object: &bundlev1.Package{
					Name: "foo",
					Secrets: &bundlev1.SecretChain{
						Data: []*bundlev1.KV{
							{
								Key: "foo",
							},
						},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &matchSecret{
				strict: tt.fields.strict,
				regex:  tt.fields.regex,
			}
			if got := s.IsSatisfiedBy(tt.args.object); got != tt.want {
				t.Errorf("matchSecret.IsSatisfiedBy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_matchSecret_IsSatisfiedBy_Fuzz(t *testing.T) {
	// Making sure the function never panics
	for i := 0; i < 50; i++ {
		f := fuzz.New()

		// Prepare arguments
		var (
			strict string
			re     *regexp.Regexp
		)

		f.Fuzz(&strict)
		// f.Fuzz(&re)

		// Execute
		s := &matchSecret{
			strict: strict,
			regex:  re,
		}
		s.IsSatisfiedBy(&bundlev1.Package{
			Name: "foo",
		})
	}
}
