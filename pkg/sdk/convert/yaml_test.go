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

package convert

import (
	"reflect"
	"testing"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/google/go-cmp/cmp"
)

func Test_PBtoYAML(t *testing.T) {
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

	expectedOutput := []byte("apiVersion: harp.elastic.co/v1\nkind: BundlePatch\nmeta:\n  name: test-patch\nspec:\n  rules:\n  - package: {}\n    selector: {}\n")

	out, err := PBtoYAML(spec)
	if err != nil {
		t.Error(err)
	}

	if report := cmp.Diff(string(out), string(expectedOutput)); report != "" {
		t.Errorf("unexpected conversion output:\n%v", report)
	}
}

func Test_convertMapStringInterface(t *testing.T) {
	type args struct {
		val interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name:    "nil",
			wantErr: false,
		},
		{
			name: "map[interface{}]interface{}",
			args: args{
				val: map[interface{}]interface{}{
					"abc":  1234,
					"true": 12.56,
				},
			},
			wantErr: false,
			want: map[string]interface{}{
				"abc":  1234,
				"true": 12.56,
			},
		},
		{
			name: "[]interface{}",
			args: args{
				val: []interface{}{
					"abc", "true",
				},
			},
			wantErr: false,
			want:    []interface{}{"abc", "true"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertMapStringInterface(tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertMapStringInterface() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertMapStringInterface() = %v, want %v", got, tt.want)
			}
		})
	}
}
