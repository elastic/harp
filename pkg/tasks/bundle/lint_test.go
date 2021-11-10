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

package bundle

import (
	"context"
	"testing"

	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/tasks"
)

func TestLintTask_Run(t *testing.T) {
	type fields struct {
		ContainerReader tasks.ReaderProvider
		RuleSetReader   tasks.ReaderProvider
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "nil",
			wantErr: true,
		},
		{
			name: "nil ruleSetReader",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				RuleSetReader:   nil,
			},
			wantErr: true,
		},
		{
			name: "containerReader error",
			fields: fields{
				ContainerReader: cmdutil.FileReader("non-existent.bundle"),
				RuleSetReader:   cmdutil.FileReader("../../../test/fixtures/ruleset/valid/cso.yaml"),
			},
			wantErr: true,
		},
		{
			name: "containerReader - not a bundle",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.json"),
				RuleSetReader:   cmdutil.FileReader("../../../test/fixtures/ruleset/valid/cso.yaml"),
			},
			wantErr: true,
		},
		{
			name: "ruleSetReader error",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				RuleSetReader:   cmdutil.FileReader("non-existent.yaml"),
			},
			wantErr: true,
		},
		{
			name: "containerReader - not a yaml",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				RuleSetReader:   cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
			},
			wantErr: true,
		},
		// ---------------------------------------------------------------------
		{
			name: "valid",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				RuleSetReader:   cmdutil.FileReader("../../../test/fixtures/ruleset/valid/cso.yaml"),
			},
			wantErr: false,
		},
		{
			name: "rule violation",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				RuleSetReader:   cmdutil.FileReader("../../../test/fixtures/ruleset/valid/database-secret-validator.yaml"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &LintTask{
				ContainerReader: tt.fields.ContainerReader,
				RuleSetReader:   tt.fields.RuleSetReader,
			}
			if err := tr.Run(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("LintTask.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
