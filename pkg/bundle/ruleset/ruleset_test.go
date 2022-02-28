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

package ruleset

import (
	"reflect"
	"testing"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
)

func TestValidate(t *testing.T) {
	type args struct {
		spec *bundlev1.RuleSet
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
				spec: &bundlev1.RuleSet{
					ApiVersion: "foo",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid kind",
			args: args{
				spec: &bundlev1.RuleSet{
					ApiVersion: "harp.elastic.co/v1",
					Kind:       "foo",
				},
			},
			wantErr: true,
		},
		{
			name: "nil meta",
			args: args{
				spec: &bundlev1.RuleSet{
					ApiVersion: "harp.elastic.co/v1",
					Kind:       "RuleSet",
				},
			},
			wantErr: true,
		},
		{
			name: "meta name not defined",
			args: args{
				spec: &bundlev1.RuleSet{
					ApiVersion: "harp.elastic.co/v1",
					Kind:       "RuleSet",
					Meta:       &bundlev1.RuleSetMeta{},
				},
			},
			wantErr: true,
		},
		{
			name: "nil spec",
			args: args{
				spec: &bundlev1.RuleSet{
					ApiVersion: "harp.elastic.co/v1",
					Kind:       "RuleSet",
					Meta:       &bundlev1.RuleSetMeta{},
				},
			},
			wantErr: true,
		},
		{
			name: "no action patch",
			args: args{
				spec: &bundlev1.RuleSet{
					ApiVersion: "harp.elastic.co/v1",
					Kind:       "RuleSet",
					Meta:       &bundlev1.RuleSetMeta{},
					Spec:       &bundlev1.RuleSetSpec{},
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
		spec *bundlev1.RuleSet
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
				spec: &bundlev1.RuleSet{
					ApiVersion: "harp.elastic.co/v1",
					Kind:       "RuleSet",
					Meta:       &bundlev1.RuleSetMeta{},
					Spec:       &bundlev1.RuleSetSpec{},
				},
			},
			wantErr: false,
			want:    "yM_TR6rMWW7BGA1Ms-U3WK6E4Xax5qRBMjK4VQLyZmQ",
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
