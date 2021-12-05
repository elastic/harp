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

	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/value"
	"github.com/elastic/harp/pkg/sdk/value/encryption"
	"github.com/elastic/harp/pkg/sdk/value/identity"
	"github.com/elastic/harp/pkg/sdk/value/mock"
	"github.com/elastic/harp/pkg/tasks"

	// Imported for tests
	_ "github.com/elastic/harp/pkg/sdk/value/encryption/jwe"
)

func TestRecoverTask_Run(t *testing.T) {
	type fields struct {
		JSONReader   tasks.ReaderProvider
		OutputWriter tasks.WriterProvider
		Transformer  value.Transformer
		JSONOutput   bool
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
			name: "nil jsonReader",
			fields: fields{
				JSONReader: nil,
			},
			wantErr: true,
		},
		{
			name: "nil outputWriter",
			fields: fields{
				JSONReader:   cmdutil.FileReader("../../../test/fixtures/identity/security.v1.json"),
				OutputWriter: nil,
			},
			wantErr: true,
		},
		{
			name: "nil transformer",
			fields: fields{
				JSONReader:   cmdutil.FileReader("../../../test/fixtures/identity/security.v1.json"),
				OutputWriter: cmdutil.DiscardWriter(),
				Transformer:  nil,
			},
			wantErr: true,
		},
		{
			name: "jsonReader error",
			fields: fields{
				JSONReader:   cmdutil.FileReader("non-existent.json"),
				OutputWriter: cmdutil.DiscardWriter(),
				Transformer:  identity.Transformer(),
			},
			wantErr: true,
		},
		{
			name: "jsonReader not an identity",
			fields: fields{
				JSONReader:   cmdutil.FileReader("../../../test/fixtures/bundles/complete.json"),
				OutputWriter: cmdutil.DiscardWriter(),
				Transformer:  identity.Transformer(),
			},
			wantErr: true,
		},
		{
			name: "transformer error",
			fields: fields{
				JSONReader:   cmdutil.FileReader("../../../test/fixtures/identity/security.v1.json"),
				OutputWriter: cmdutil.DiscardWriter(),
				Transformer:  mock.Transformer(errors.New("test")),
			},
			wantErr: true,
		},
		{
			name: "outputWriter error",
			fields: fields{
				JSONReader: cmdutil.FileReader("../../../test/fixtures/identity/security.v1.json"),
				OutputWriter: func(ctx context.Context) (io.Writer, error) {
					return nil, errors.New("test")
				},
				Transformer: encryption.Must(encryption.FromKey("jwe:pbes2-hs512-a256kw:test")),
			},
			wantErr: true,
		},
		{
			name: "outputWriter closed",
			fields: fields{
				JSONReader:  cmdutil.FileReader("../../../test/fixtures/identity/security.v1.json"),
				Transformer: encryption.Must(encryption.FromKey("jwe:pbes2-hs512-a256kw:test")),
				OutputWriter: func(ctx context.Context) (io.Writer, error) {
					return cmdutil.NewClosedWriter(), nil
				},
			},
			wantErr: true,
		},
		{
			name: "outputWriter closed - json",
			fields: fields{
				JSONReader:  cmdutil.FileReader("../../../test/fixtures/identity/security.v1.json"),
				Transformer: encryption.Must(encryption.FromKey("jwe:pbes2-hs512-a256kw:test")),
				OutputWriter: func(ctx context.Context) (io.Writer, error) {
					return cmdutil.NewClosedWriter(), nil
				},
				JSONOutput: true,
			},
			wantErr: true,
		},
		// ---------------------------------------------------------------------
		{
			name: "valid - v1",
			fields: fields{
				JSONReader:   cmdutil.FileReader("../../../test/fixtures/identity/security.v1.json"),
				OutputWriter: cmdutil.DiscardWriter(),
				Transformer:  encryption.Must(encryption.FromKey("jwe:pbes2-hs512-a256kw:test")),
			},
			wantErr: false,
		},
		{
			name: "valid - v1 - json output",
			fields: fields{
				JSONReader:   cmdutil.FileReader("../../../test/fixtures/identity/security.v2.json"),
				OutputWriter: cmdutil.DiscardWriter(),
				Transformer:  encryption.Must(encryption.FromKey("jwe:pbes2-hs512-a256kw:test")),
				JSONOutput:   true,
			},
			wantErr: false,
		},
		{
			name: "valid - v2",
			fields: fields{
				JSONReader:   cmdutil.FileReader("../../../test/fixtures/identity/security.v2.json"),
				OutputWriter: cmdutil.DiscardWriter(),
				Transformer:  encryption.Must(encryption.FromKey("jwe:pbes2-hs512-a256kw:test")),
			},
			wantErr: false,
		},
		{
			name: "valid - v2 - json output",
			fields: fields{
				JSONReader:   cmdutil.FileReader("../../../test/fixtures/identity/security.v2.json"),
				OutputWriter: cmdutil.DiscardWriter(),
				Transformer:  encryption.Must(encryption.FromKey("jwe:pbes2-hs512-a256kw:test")),
				JSONOutput:   true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &RecoverTask{
				JSONReader:   tt.fields.JSONReader,
				OutputWriter: tt.fields.OutputWriter,
				Transformer:  tt.fields.Transformer,
				JSONOutput:   tt.fields.JSONOutput,
			}
			if err := tr.Run(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("RecoverTask.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
