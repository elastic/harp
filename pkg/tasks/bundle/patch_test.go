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
	"errors"
	"io"
	"testing"

	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/tasks"
)

func TestPatchTask_Run(t *testing.T) {
	type fields struct {
		PatchReader     tasks.ReaderProvider
		ContainerReader tasks.ReaderProvider
		OutputWriter    tasks.WriterProvider
		Values          map[string]interface{}
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
			name: "nil patchReader",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				PatchReader:     nil,
			},
			wantErr: true,
		},
		{
			name: "nil outputWriter",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				PatchReader:     cmdutil.FileReader("../../../test/fixtures/patch/valid/path-cleaner.yaml"),
				OutputWriter:    nil,
			},
			wantErr: true,
		},
		{
			name: "containerReader error",
			fields: fields{
				ContainerReader: cmdutil.FileReader("non-existent.bundle"),
				PatchReader:     cmdutil.FileReader("../../../test/fixtures/patch/valid/path-cleaner.yaml"),
				OutputWriter:    cmdutil.DiscardWriter(),
			},
			wantErr: true,
		},
		{
			name: "containerReader - not a bundle",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.json"),
				PatchReader:     cmdutil.FileReader("../../../test/fixtures/patch/valid/path-cleaner.yaml"),
				OutputWriter:    cmdutil.DiscardWriter(),
			},
			wantErr: true,
		},
		{
			name: "patchReader error",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				PatchReader:     cmdutil.FileReader("non-existent.yaml"),
				OutputWriter:    cmdutil.DiscardWriter(),
			},
			wantErr: true,
		},
		{
			name: "patchReader - not a yaml",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				PatchReader:     cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				OutputWriter:    cmdutil.DiscardWriter(),
			},
			wantErr: true,
		},
		{
			name: "outputWriter error",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				PatchReader:     cmdutil.FileReader("../../../test/fixtures/patch/valid/path-cleaner.yaml"),
				OutputWriter: func(ctx context.Context) (io.Writer, error) {
					return nil, errors.New("test")
				},
			},
			wantErr: true,
		},
		{
			name: "outputWriter closed",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				PatchReader:     cmdutil.FileReader("../../../test/fixtures/patch/valid/path-cleaner.yaml"),
				OutputWriter: func(ctx context.Context) (io.Writer, error) {
					return cmdutil.NewClosedWriter(), nil
				},
			},
			wantErr: true,
		},
		// ---------------------------------------------------------------------
		{
			name: "valid",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				PatchReader:     cmdutil.FileReader("../../../test/fixtures/patch/valid/path-cleaner.yaml"),
				OutputWriter:    cmdutil.DiscardWriter(),
			},
			wantErr: false,
		},
		{
			name: "empty patch",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				PatchReader:     cmdutil.FileReader("../../../test/fixtures/patch/valid/empty.yaml"),
				OutputWriter:    cmdutil.DiscardWriter(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &PatchTask{
				PatchReader:     tt.fields.PatchReader,
				ContainerReader: tt.fields.ContainerReader,
				OutputWriter:    tt.fields.OutputWriter,
				Values:          tt.fields.Values,
			}
			if err := tr.Run(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("PatchTask.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
