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
	"github.com/elastic/harp/pkg/sdk/value"
	"github.com/elastic/harp/pkg/sdk/value/encryption"
	"github.com/elastic/harp/pkg/sdk/value/identity"
	"github.com/elastic/harp/pkg/sdk/value/mock"
	"github.com/elastic/harp/pkg/tasks"
)

func TestEncryptTask_Run(t *testing.T) {
	type fields struct {
		ContainerReader   tasks.ReaderProvider
		OutputWriter      tasks.WriterProvider
		BundleTransformer value.Transformer
		TransformerMap    map[string]value.Transformer
		SkipUnresolved    bool
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
			name: "nil outputWriter",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				OutputWriter:    nil,
			},
			wantErr: true,
		},
		{
			name: "nil transformers",
			fields: fields{
				ContainerReader:   cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				OutputWriter:      cmdutil.DiscardWriter(),
				BundleTransformer: nil,
				TransformerMap:    nil,
			},
			wantErr: true,
		},
		{
			name: "containerReader error",
			fields: fields{
				ContainerReader:   cmdutil.FileReader("non-existent.bundle"),
				OutputWriter:      cmdutil.DiscardWriter(),
				BundleTransformer: identity.Transformer(),
			},
			wantErr: true,
		},
		{
			name: "containerReader - not a bundle",
			fields: fields{
				ContainerReader:   cmdutil.FileReader("../../../test/fixtures/bundles/complete.json"),
				OutputWriter:      cmdutil.DiscardWriter(),
				BundleTransformer: identity.Transformer(),
			},
			wantErr: true,
		},
		{
			name: "bundle transformer error",
			fields: fields{
				ContainerReader:   cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				OutputWriter:      cmdutil.DiscardWriter(),
				BundleTransformer: mock.Transformer(errors.New("test")),
			},
			wantErr: true,
		},
		{
			name: "empty annotation transformer map",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				OutputWriter:    cmdutil.DiscardWriter(),
				TransformerMap:  map[string]value.Transformer{},
			},
			wantErr: true,
		},
		{
			name: "annotation transformer error",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				OutputWriter:    cmdutil.DiscardWriter(),
				TransformerMap: map[string]value.Transformer{
					"test": mock.Transformer(errors.New("test")),
				},
			},
			wantErr: true,
		},
		{
			name: "outputWriter error",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.aes-gcm.bundle"),
				OutputWriter: func(ctx context.Context) (io.Writer, error) {
					return nil, errors.New("test")
				},
				BundleTransformer: encryption.Must(encryption.FromKey("aes-gcm:5OSpiJUr_XS2M1_vvTBeGg==")),
			},
			wantErr: true,
		},
		{
			name: "outputWriter closed",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.aes-gcm.bundle"),
				OutputWriter: func(ctx context.Context) (io.Writer, error) {
					return cmdutil.NewClosedWriter(), nil
				},
				BundleTransformer: encryption.Must(encryption.FromKey("aes-gcm:5OSpiJUr_XS2M1_vvTBeGg==")),
			},
			wantErr: true,
		},
		// ---------------------------------------------------------------------
		{
			name: "valid",
			fields: fields{
				ContainerReader:   cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				OutputWriter:      cmdutil.DiscardWriter(),
				BundleTransformer: encryption.Must(encryption.FromKey("aes-gcm:5OSpiJUr_XS2M1_vvTBeGg==")),
			},
			wantErr: false,
		},
		{
			name: "valid - unused key alias",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				OutputWriter:    cmdutil.DiscardWriter(),
				TransformerMap: map[string]value.Transformer{
					"not-used": encryption.Must(encryption.FromKey("aes-gcm:5OSpiJUr_XS2M1_vvTBeGg==")),
				},
				SkipUnresolved: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &EncryptTask{
				ContainerReader:   tt.fields.ContainerReader,
				OutputWriter:      tt.fields.OutputWriter,
				BundleTransformer: tt.fields.BundleTransformer,
				TransformerMap:    tt.fields.TransformerMap,
				SkipUnresolved:    tt.fields.SkipUnresolved,
			}
			if err := tr.Run(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("EncryptTask.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
