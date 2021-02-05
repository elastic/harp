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
	"testing"
)

func Test_extractVersion(t *testing.T) {
	type args struct {
		packagePath string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   uint32
		wantErr bool
	}{
		{
			name:    "blank",
			wantErr: true,
		},
		{
			name: "no version",
			args: args{
				packagePath: "app/test",
			},
			wantErr: false,
			want:    "app/test",
			want1:   0,
		},
		{
			name: "with version",
			args: args{
				packagePath: "app/test?version=14",
			},
			wantErr: false,
			want:    "app/test",
			want1:   14,
		},
		{
			name: "with invalid version",
			args: args{
				packagePath: "app/test?version=azerty",
			},
			wantErr: true,
		},
		{
			name: "with invalid path",
			args: args{
				packagePath: "\n\t",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := extractVersion(tt.args.packagePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("exporter.extractVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("exporter.extractVersion() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("exporter.extractVersion() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
