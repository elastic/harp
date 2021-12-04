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

package container

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/awnumar/memguard"
	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/tasks"
)

func TestUnsealTask_Run(t *testing.T) {
	type fields struct {
		ContainerReader tasks.ReaderProvider
		OutputWriter    tasks.WriterProvider
		ContainerKey    *memguard.LockedBuffer
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
			name: "nil containerReader",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
			},
			wantErr: true,
		},
		{
			name: "nil outputWriter",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				OutputWriter:    nil,
			},
			wantErr: true,
		},
		{
			name: "nil containerKey",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				OutputWriter:    cmdutil.DiscardWriter(),
				ContainerKey:    nil,
			},
			wantErr: true,
		},
		{
			name: "containerReader error",
			fields: fields{
				ContainerReader: cmdutil.FileReader("non-existent.bundle"),
				OutputWriter:    cmdutil.DiscardWriter(),
				ContainerKey:    memguard.NewBuffer(32),
			},
			wantErr: true,
		},
		{
			name: "containerReader not a bundle",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.json"),
				OutputWriter:    cmdutil.DiscardWriter(),
				ContainerKey:    memguard.NewBuffer(32),
			},
			wantErr: true,
		},
		{
			name: "invalid container key",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				OutputWriter:    cmdutil.DiscardWriter(),
				ContainerKey:    memguard.NewBuffer(32),
			},
			wantErr: true,
		},
		{
			name: "outputWriter error",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.v1.sealed"),
				OutputWriter: func(ctx context.Context) (io.Writer, error) {
					return nil, errors.New("test")
				},
				ContainerKey: memguard.NewBufferFromBytes([]byte("v1.ck.MiVGh4KOmdzZbej17BZGChkCPZ9uK9uBWdPNU0GlBNg")),
			},
			wantErr: true,
		},
		{
			name: "outputWriter closed",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.v1.sealed"),
				OutputWriter: func(ctx context.Context) (io.Writer, error) {
					return cmdutil.NewClosedWriter(), nil
				},
				ContainerKey: memguard.NewBufferFromBytes([]byte("v1.ck.MiVGh4KOmdzZbej17BZGChkCPZ9uK9uBWdPNU0GlBNg")),
			},
			wantErr: true,
		},
		{
			name: "v2 without prefix",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.v2.sealed"),
				OutputWriter:    cmdutil.DiscardWriter(),
				ContainerKey:    memguard.NewBufferFromBytes([]byte("v2.ck.dAYx4CeTMRGKfpFHA7Q926qMz8imo1VJIToMw9uvH7HfPJTRpLUSMUS07JAdV-1R")),
			},
			wantErr: true,
		},
		// ---------------------------------------------------------------------
		{
			name: "valid - v1",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.v1.sealed"),
				OutputWriter:    cmdutil.DiscardWriter(),
				ContainerKey:    memguard.NewBufferFromBytes([]byte("v1.ck.MiVGh4KOmdzZbej17BZGChkCPZ9uK9uBWdPNU0GlBNg")),
			},
			wantErr: false,
		},
		{
			name: "valid - v1 - with identity recovery key",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.v1.sealed"),
				OutputWriter:    cmdutil.DiscardWriter(),
				ContainerKey:    memguard.NewBufferFromBytes([]byte("v1.ck.IO6bCjACnqsCP0ahT--CVBhryzhe-ZFroVzn5Dx3D0U")),
			},
			wantErr: false,
		},
		{
			name: "valid - v1 - with identity recovery key with prefix",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.v1.sealed"),
				OutputWriter:    cmdutil.DiscardWriter(),
				ContainerKey:    memguard.NewBufferFromBytes([]byte("v1.ck.IO6bCjACnqsCP0ahT--CVBhryzhe-ZFroVzn5Dx3D0U")),
			},
			wantErr: false,
		},
		{
			name: "valid - v2",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.v2.sealed"),
				OutputWriter:    cmdutil.DiscardWriter(),
				ContainerKey:    memguard.NewBufferFromBytes([]byte("v2.ck.P5l8Li3hRAsmCv4DPAPGr5VUMi4MGUsiSki1IDqIb0y6neJIU7VPBKqqhE0UR-x4")),
			},
			wantErr: false,
		},
		{
			name: "valid - v2 - with identity recovery key",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.v2.sealed"),
				OutputWriter:    cmdutil.DiscardWriter(),
				ContainerKey:    memguard.NewBufferFromBytes([]byte("v2.ck.VHJBdjBLWWJsMktxQ285ZoFXc5G4HY_0qSMZAibGlchUmqt915byglIOGeel-5X5")),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &UnsealTask{
				ContainerReader: tt.fields.ContainerReader,
				OutputWriter:    tt.fields.OutputWriter,
				ContainerKey:    tt.fields.ContainerKey,
			}
			if err := tr.Run(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("UnsealTask.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
