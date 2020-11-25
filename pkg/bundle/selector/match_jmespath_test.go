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
	"testing"

	fuzz "github.com/google/gofuzz"
	"github.com/jmespath/go-jmespath"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
)

func Test_matchJMESPath_IsSatisfiedBy(t *testing.T) {
	type fields struct {
		exp *jmespath.JMESPath
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
			fields: fields{
				exp: jmespath.MustCompile("true"),
			},
			args: args{
				object: struct{}{},
			},
			want: false,
		},
		{
			name: "supported type: empty object",
			fields: fields{
				exp: jmespath.MustCompile("true"),
			},
			args: args{
				object: &bundlev1.Package{},
			},
			want: false,
		},
		{
			name: "supported type: nil exp",
			args: args{
				object: &bundlev1.Package{},
			},
			want: false,
		},
		{
			name: "supported type: not matching",
			fields: fields{
				exp: jmespath.MustCompile("name=='test'"),
			},
			args: args{
				object: &bundlev1.Package{
					Name: "foo",
				},
			},
			want: false,
		},
		{
			name: "supported type: annotations query with nil",
			fields: fields{
				exp: jmespath.MustCompile("annotations.patched=='foo'"),
			},
			args: args{
				object: &bundlev1.Package{
					Name: "foo",
				},
			},
			want: false,
		},
		{
			name: "supported type: annotations not matching",
			fields: fields{
				exp: jmespath.MustCompile("annotations.patched=='foo'"),
			},
			args: args{
				object: &bundlev1.Package{
					Name: "foo",
					Annotations: map[string]string{
						"patched": "true",
					},
				},
			},
			want: false,
		},
		{
			name: "supported type: annotations not matching with same type",
			fields: fields{
				exp: jmespath.MustCompile("annotations.patched=='true'"),
			},
			args: args{
				object: &bundlev1.Package{
					Name: "foo",
					Annotations: map[string]string{
						"patched": "false",
					},
				},
			},
			want: false,
		},
		{
			name: "supported type: annotations matching",
			fields: fields{
				exp: jmespath.MustCompile("annotations.patched=='true'"),
			},
			args: args{
				object: &bundlev1.Package{
					Name: "foo",
					Annotations: map[string]string{
						"patched": "true",
					},
				},
			},
			want: true,
		},
		{
			name: "supported type: name matching",
			fields: fields{
				exp: jmespath.MustCompile("name=='foo'"),
			},
			args: args{
				object: &bundlev1.Package{
					Name: "foo",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &jmesPathMatcher{
				exp: tt.fields.exp,
			}
			if got := s.IsSatisfiedBy(tt.args.object); got != tt.want {
				t.Errorf("jmesPathMatcher.IsSatisfiedBy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_matchJMESPath_IsSatisfiedBy_Fuzz(t *testing.T) {
	// Making sure the function never panics
	for i := 0; i < 50; i++ {
		f := fuzz.New()

		// Prepare arguments
		var (
			expr *jmespath.JMESPath
		)

		f.Fuzz(&expr)

		// Execute
		s := &jmesPathMatcher{
			exp: expr,
		}
		s.IsSatisfiedBy(&bundlev1.Package{
			Name: "foo",
		})
	}
}
