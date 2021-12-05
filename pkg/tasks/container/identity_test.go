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
	"github.com/elastic/harp/pkg/sdk/value/identity"
	"github.com/elastic/harp/pkg/sdk/value/mock"
	"github.com/elastic/harp/pkg/tasks"
)

func TestIdentityTask_Run(t *testing.T) {
	type fields struct {
		OutputWriter tasks.WriterProvider
		Description  string
		Transformer  value.Transformer
		Version      IdentityVersion
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
				OutputWriter: nil,
			},
			wantErr: true,
		},
		{
			name: "nil transformer",
			fields: fields{
				OutputWriter: cmdutil.DiscardWriter(),
				Transformer:  nil,
			},
			wantErr: true,
		},
		{
			name: "blank description",
			fields: fields{
				OutputWriter: cmdutil.DiscardWriter(),
				Transformer:  identity.Transformer(),
				Description:  "",
			},
			wantErr: true,
		},
		{
			name: "transformer error",
			fields: fields{
				OutputWriter: cmdutil.DiscardWriter(),
				Transformer:  mock.Transformer(errors.New("test")),
				Description:  "test",
			},
			wantErr: true,
		},
		{
			name: "outputWriter error",
			fields: fields{
				OutputWriter: func(ctx context.Context) (io.Writer, error) {
					return nil, errors.New("test")
				},
				Description: "test",
				Transformer: identity.Transformer(),
			},
			wantErr: true,
		},
		{
			name: "outputWriter closed",
			fields: fields{
				OutputWriter: func(ctx context.Context) (io.Writer, error) {
					return cmdutil.NewClosedWriter(), nil
				},
				Description: "test",
				Transformer: identity.Transformer(),
			},
			wantErr: true,
		},
		{
			name: "version unspecified",
			fields: fields{
				OutputWriter: cmdutil.DiscardWriter(),
				Description:  "test",
				Transformer:  identity.Transformer(),
			},
			wantErr: true,
		},
		// ---------------------------------------------------------------------
		{
			name: "valid - v1",
			fields: fields{
				OutputWriter: cmdutil.DiscardWriter(),
				Description:  "test",
				Transformer:  identity.Transformer(),
				Version:      LegacyIdentity,
			},
			wantErr: false,
		},
		{
			name: "valid - v2",
			fields: fields{
				OutputWriter: cmdutil.DiscardWriter(),
				Description:  "test",
				Transformer:  identity.Transformer(),
				Version:      ModernIdentity,
			},
			wantErr: false,
		},
		{
			name: "valid - v3",
			fields: fields{
				OutputWriter: cmdutil.DiscardWriter(),
				Description:  "test",
				Transformer:  identity.Transformer(),
				Version:      NISTIdentity,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &IdentityTask{
				OutputWriter: tt.fields.OutputWriter,
				Description:  tt.fields.Description,
				Transformer:  tt.fields.Transformer,
				Version:      tt.fields.Version,
			}
			if err := tr.Run(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("IdentityTask.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
