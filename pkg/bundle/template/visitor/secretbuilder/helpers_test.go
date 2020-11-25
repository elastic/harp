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

package secretbuilder

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	csov1 "github.com/elastic/harp/pkg/cso/v1"
	"github.com/elastic/harp/pkg/template/engine"
)

func TestSuffix(t *testing.T) {
	type args struct {
		templateContext engine.Context
		ring            csov1.Ring
		secretPath      string
		item            *bundlev1.SecretSuffix
		data            interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name:    "nil",
			args:    args{},
			wantErr: true,
		},
		{
			name: "invalid template function",
			args: args{
				templateContext: engine.NewContext(),
				ring:            csov1.RingInfra,
				secretPath:      "infra/aws/foo/us-east-1/rds/database/root_credentials",
				item: &bundlev1.SecretSuffix{
					Template: `{{ foo }}`,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid json",
			args: args{
				templateContext: engine.NewContext(),
				ring:            csov1.RingInfra,
				secretPath:      "infra/aws/foo/us-east-1/rds/database/root_credentials",
				item: &bundlev1.SecretSuffix{
					Template: `{"foo`,
				},
			},
			wantErr: true,
		},
		{
			name: "valid",
			args: args{
				templateContext: engine.NewContext(),
				ring:            csov1.RingInfra,
				secretPath:      "infra/aws/foo/us-east-1/rds/database/root_credentials",
				item: &bundlev1.SecretSuffix{
					Template: `{"foo":"123456"}`,
				},
			},
			wantErr: false,
			want: map[string]interface{}{
				"foo": "123456",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderSuffix(tt.args.templateContext, tt.args.secretPath, tt.args.item, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Suffix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("Suffix() = %v, want %v", got, tt.want)
			}
		})
	}
}
