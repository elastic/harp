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

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/tasks"
)

var (
	opt = cmp.FilterPath(
		func(p cmp.Path) bool {
			// Remove ignoring of the fields below once go-cmp is able to ignore generated fields.
			// See https://github.com/google/go-cmp/issues/153
			ignoreXXXCache :=
				p.String() == "XXX_sizecache" ||
					p.String() == "Packages.XXX_sizecache" ||
					p.String() == "Packages.Secrets.XXX_sizecache" ||
					p.String() == "Packages.Secrets.Data.XXX_sizecache"
			return ignoreXXXCache
		}, cmp.Ignore())

	ignoreOpts = []cmp.Option{
		cmpopts.IgnoreUnexported(bundlev1.Bundle{}),
		cmpopts.IgnoreUnexported(bundlev1.Package{}),
		cmpopts.IgnoreUnexported(bundlev1.SecretChain{}),
		cmpopts.IgnoreUnexported(bundlev1.KV{}),
		opt,
	}
)

func TestFilterTask_Run(t *testing.T) {
	type fields struct {
		ContainerReader tasks.ReaderProvider
		OutputWriter    tasks.WriterProvider
		ReverseLogic    bool
		KeepPaths       []string
		ExcludePaths    []string
		JMESPath        string
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
			name: "containerReader error",
			fields: fields{
				ContainerReader: cmdutil.FileReader("non-existent.bundle"),
				OutputWriter:    cmdutil.DiscardWriter(),
			},
			wantErr: true,
		},
		{
			name: "containerReader - not a bundle",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.json"),
				OutputWriter:    cmdutil.DiscardWriter(),
			},
			wantErr: true,
		},
		{
			name: "outputWriter error",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
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
				OutputWriter: func(ctx context.Context) (io.Writer, error) {
					return cmdutil.NewClosedWriter(), nil
				},
			},
			wantErr: true,
		},
		{
			name: "keep - invalid regexp",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				OutputWriter:    cmdutil.DiscardWriter(),
				KeepPaths:       []string{"(["},
			},
			wantErr: true,
		},
		{
			name: "exclude - invalid regexp",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				OutputWriter:    cmdutil.DiscardWriter(),
				ExcludePaths:    []string{"(["},
			},
			wantErr: true,
		},
		{
			name: "jmespath - invalid",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				OutputWriter:    cmdutil.DiscardWriter(),
				JMESPath:        ".",
			},
			wantErr: true,
		},
		// ---------------------------------------------------------------------
		{
			name: "valid - noop",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				OutputWriter:    cmdutil.DiscardWriter(),
			},
			wantErr: false,
		},
		{
			name: "valid - keep",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				OutputWriter:    cmdutil.DiscardWriter(),
				KeepPaths:       []string{"app/*"},
			},
			wantErr: false,
		},
		{
			name: "valid - exclude",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				OutputWriter:    cmdutil.DiscardWriter(),
				ExcludePaths:    []string{"^product/*"},
			},
			wantErr: false,
		},
		{
			name: "valid - jmespath",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				OutputWriter:    cmdutil.DiscardWriter(),
				JMESPath:        "labels.okta == 'true'",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &FilterTask{
				ContainerReader: tt.fields.ContainerReader,
				OutputWriter:    tt.fields.OutputWriter,
				ReverseLogic:    tt.fields.ReverseLogic,
				KeepPaths:       tt.fields.KeepPaths,
				ExcludePaths:    tt.fields.ExcludePaths,
				JMESPath:        tt.fields.JMESPath,
			}
			if err := tr.Run(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("FilterTask.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFilterTask_keepFilter(t *testing.T) {
	type fields struct {
		ContainerReader tasks.ReaderProvider
		OutputWriter    tasks.WriterProvider
		ReverseLogic    bool
		KeepPaths       []string
		ExcludePaths    []string
		JMESPath        string
	}
	type args struct {
		in           []*bundlev1.Package
		paths        []string
		reverseLogic bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*bundlev1.Package
		wantErr bool
	}{
		{
			name:    "nil",
			wantErr: false,
			want:    nil,
		},
		{
			name: "empty packages",
			args: args{
				in: []*bundlev1.Package{},
			},
			wantErr: false,
			want:    []*bundlev1.Package{},
		},
		{
			name: "empty paths",
			args: args{
				in: []*bundlev1.Package{
					{
						Name: "app/production/test",
					},
				},
				paths: []string{},
			},
			wantErr: false,
			want: []*bundlev1.Package{
				{
					Name: "app/production/test",
				},
			},
		},
		{
			name: "invalid regexp",
			args: args{
				paths: []string{
					`[(`,
				},
				in: []*bundlev1.Package{
					{
						Name: "app/production/test",
					},
				},
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "empty package collection",
			args: args{
				in: []*bundlev1.Package{},
				paths: []string{
					"^app/*",
				},
			},
			want:    []*bundlev1.Package{},
			wantErr: false,
		},
		{
			name: "valid",
			args: args{
				in: []*bundlev1.Package{
					{
						Name: "app/production/test",
					},
					{
						Name: "product/security/test",
					},
				},
				paths: []string{
					"^app/*",
				},
			},
			want: []*bundlev1.Package{
				{
					Name: "app/production/test",
				},
			},
			wantErr: false,
		},
		{
			name: "valid - reverse",
			args: args{
				in: []*bundlev1.Package{
					{
						Name: "app/production/test",
					},
					{
						Name: "product/security/test",
					},
				},
				paths: []string{
					"^app/*",
				},
				reverseLogic: true,
			},
			want: []*bundlev1.Package{
				{
					Name: "product/security/test",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &FilterTask{
				ContainerReader: tt.fields.ContainerReader,
				OutputWriter:    tt.fields.OutputWriter,
				ReverseLogic:    tt.fields.ReverseLogic,
				KeepPaths:       tt.fields.KeepPaths,
				ExcludePaths:    tt.fields.ExcludePaths,
				JMESPath:        tt.fields.JMESPath,
			}
			got, err := tr.keepFilter(tt.args.in, tt.args.paths, tt.args.reverseLogic)
			if (err != nil) != tt.wantErr {
				t.Errorf("FilterTask.keepFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want, ignoreOpts...); diff != "" {
				t.Errorf("%q. FilterTask.keepFilter():\n-got/+want\ndiff %s", tt.name, diff)
			}
		})
	}
}

func TestFilterTask_excludeFilter(t *testing.T) {
	type fields struct {
		ContainerReader tasks.ReaderProvider
		OutputWriter    tasks.WriterProvider
		ReverseLogic    bool
		KeepPaths       []string
		ExcludePaths    []string
		JMESPath        string
	}
	type args struct {
		in           []*bundlev1.Package
		paths        []string
		reverseLogic bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*bundlev1.Package
		wantErr bool
	}{
		{
			name:    "nil",
			wantErr: false,
			want:    nil,
		},
		{
			name: "empty packages",
			args: args{
				in: []*bundlev1.Package{},
			},
			wantErr: false,
			want:    []*bundlev1.Package{},
		},
		{
			name: "empty paths",
			args: args{
				in: []*bundlev1.Package{
					{
						Name: "app/production/test",
					},
				},
				paths: []string{},
			},
			wantErr: false,
			want: []*bundlev1.Package{
				{
					Name: "app/production/test",
				},
			},
		},
		{
			name: "invalid regexp",
			args: args{
				in: []*bundlev1.Package{
					{
						Name: "app/production/test",
					},
				},
				paths: []string{
					"[(",
				},
			},
			wantErr: true,
		},
		{
			name: "empty package collection",
			args: args{
				in: []*bundlev1.Package{},
				paths: []string{
					"^app/*",
				},
			},
			want:    []*bundlev1.Package{},
			wantErr: false,
		},
		{
			name: "valid",
			args: args{
				in: []*bundlev1.Package{
					{
						Name: "app/production/test",
					},
					{
						Name: "product/security/test",
					},
				},
				paths: []string{
					"^app/*",
				},
			},
			want: []*bundlev1.Package{
				{
					Name: "product/security/test",
				},
			},
			wantErr: false,
		},
		{
			name: "valid - reverse",
			args: args{
				in: []*bundlev1.Package{
					{
						Name: "app/production/test",
					},
					{
						Name: "product/security/test",
					},
				},
				paths: []string{
					"^app/*",
				},
				reverseLogic: true,
			},
			want: []*bundlev1.Package{
				{
					Name: "app/production/test",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &FilterTask{
				ContainerReader: tt.fields.ContainerReader,
				OutputWriter:    tt.fields.OutputWriter,
				ReverseLogic:    tt.fields.ReverseLogic,
				KeepPaths:       tt.fields.KeepPaths,
				ExcludePaths:    tt.fields.ExcludePaths,
				JMESPath:        tt.fields.JMESPath,
			}
			got, err := tr.excludeFilter(tt.args.in, tt.args.paths, tt.args.reverseLogic)
			if (err != nil) != tt.wantErr {
				t.Errorf("FilterTask.excludeFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got, ignoreOpts...); diff != "" {
				t.Errorf("%q. FilterTask.excludeFilter():\n-got/+want\ndiff %s", tt.name, diff)
			}
		})
	}
}

func TestFilterTask_jmespathFilter(t *testing.T) {
	type fields struct {
		ContainerReader tasks.ReaderProvider
		OutputWriter    tasks.WriterProvider
		ReverseLogic    bool
		KeepPaths       []string
		ExcludePaths    []string
		JMESPath        string
	}
	type args struct {
		in           []*bundlev1.Package
		filter       string
		reverseLogic bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*bundlev1.Package
		wantErr bool
	}{
		{
			name:    "nil",
			wantErr: false,
		},
		{
			name: "empty packages",
			args: args{
				in: []*bundlev1.Package{},
			},
			wantErr: false,
			want:    []*bundlev1.Package{},
		},
		{
			name: "empty filter",
			args: args{
				in: []*bundlev1.Package{
					{
						Name: "app/production/test",
					},
				},
				filter: "",
			},
			wantErr: false,
			want: []*bundlev1.Package{
				{
					Name: "app/production/test",
				},
			},
		},
		{
			name: "empty package collection",
			args: args{
				in:     []*bundlev1.Package{},
				filter: "labels.vendor == 'true'",
			},
			want:    []*bundlev1.Package{},
			wantErr: false,
		},
		{
			name: "invalid query",
			args: args{
				in: []*bundlev1.Package{
					{
						Name: "app/production/test",
						Labels: map[string]string{
							"vendor": "true",
						},
					},
				},
				filter: ".",
			},
			wantErr: true,
		},
		{
			name: "valid",
			args: args{
				in: []*bundlev1.Package{
					{
						Name: "app/production/test",
						Labels: map[string]string{
							"vendor": "true",
						},
					},
					{
						Name: "product/security/test",
					},
				},
				filter: "labels.vendor == 'true'",
			},
			want: []*bundlev1.Package{
				{
					Name: "app/production/test",
					Labels: map[string]string{
						"vendor": "true",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid - reverse",
			args: args{
				in: []*bundlev1.Package{
					{
						Name: "app/production/test",
						Labels: map[string]string{
							"vendor": "true",
						},
					},
					{
						Name: "product/security/test",
					},
				},
				filter:       "labels.vendor == 'true'",
				reverseLogic: true,
			},
			want: []*bundlev1.Package{
				{
					Name: "product/security/test",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &FilterTask{
				ContainerReader: tt.fields.ContainerReader,
				OutputWriter:    tt.fields.OutputWriter,
				ReverseLogic:    tt.fields.ReverseLogic,
				KeepPaths:       tt.fields.KeepPaths,
				ExcludePaths:    tt.fields.ExcludePaths,
				JMESPath:        tt.fields.JMESPath,
			}
			got, err := tr.jmespathFilter(tt.args.in, tt.args.filter, tt.args.reverseLogic)
			if (err != nil) != tt.wantErr {
				t.Errorf("FilterTask.jmespathFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got, ignoreOpts...); diff != "" {
				t.Errorf("%q. FilterTask.jmespathFilter():\n-got/+want\ndiff %s", tt.name, diff)
			}
		})
	}
}
