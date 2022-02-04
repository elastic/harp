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

package kv

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	vaultApi "github.com/hashicorp/vault/api"

	"github.com/elastic/harp/pkg/vault/logical"
)

func Test_KVV1_List(t *testing.T) {
	type args struct {
		ctx  context.Context
		path string
	}
	tests := []struct {
		name    string
		prepare func(*logical.MockLogical)
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "blank",
			args: args{
				ctx:  context.Background(),
				path: "",
			},
			wantErr: true,
		},
		{
			name: "query error",
			args: args{
				ctx:  context.Background(),
				path: "secrets/application/foo",
			},
			prepare: func(logical *logical.MockLogical) {
				logical.EXPECT().List("secrets/application/foo").Return(&vaultApi.Secret{}, fmt.Errorf("foo"))
			},
			wantErr: true,
		},
		{
			name: "nil secret",
			args: args{
				ctx:  context.Background(),
				path: "secrets/application/foo",
			},
			prepare: func(logical *logical.MockLogical) {
				logical.EXPECT().List("secrets/application/foo").Return(nil, nil)
			},
			wantErr: false,
		},
		{
			name: "nil secret data",
			args: args{
				ctx:  context.Background(),
				path: "secrets/application/foo",
			},
			prepare: func(logical *logical.MockLogical) {
				logical.EXPECT().List("secrets/application/foo").Return(&vaultApi.Secret{
					Data: nil,
				}, nil)
			},
			wantErr: true,
		},
		{
			name: "missing keys data",
			args: args{
				ctx:  context.Background(),
				path: "secrets/application/foo",
			},
			prepare: func(logical *logical.MockLogical) {
				logical.EXPECT().List("secrets/application/foo").Return(&vaultApi.Secret{
					Data: SecretData{},
				}, nil)
			},
			wantErr: true,
		},
		{
			name: "invalid keys type",
			args: args{
				ctx:  context.Background(),
				path: "secrets/application/foo",
			},
			prepare: func(logical *logical.MockLogical) {
				logical.EXPECT().List("secrets/application/foo").Return(&vaultApi.Secret{
					Data: SecretData{
						"keys": 1,
					},
				}, nil)
			},
			wantErr: true,
		},
		{
			name: "unclean",
			args: args{
				ctx:  context.Background(),
				path: "    /secrets/application/foo/   ",
			},
			prepare: func(logical *logical.MockLogical) {
				logical.EXPECT().List("secrets/application/foo").Return(&vaultApi.Secret{
					Data: SecretData{
						"keys": []interface{}{},
					},
				}, nil)
			},
			wantErr: false,
			want:    []string{},
		},
		{
			name: "valid",
			args: args{
				ctx:  context.Background(),
				path: "secrets/application/foo",
			},
			prepare: func(logical *logical.MockLogical) {
				logical.EXPECT().List("secrets/application/foo").Return(&vaultApi.Secret{
					Data: SecretData{
						"keys": []interface{}{"secrets/application/foo/secret-1", "secrets/application/foo/secret-2"},
					},
				}, nil)
			},
			wantErr: false,
			want: []string{
				"secrets/application/foo/secret-1",
				"secrets/application/foo/secret-2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Arm mocks
			logicalMock := logical.NewMockLogical(ctrl)

			// Prepare mocks
			if tt.prepare != nil {
				tt.prepare(logicalMock)
			}

			// Service
			underTest := V1(logicalMock, "secrets/")
			got, err := underTest.List(tt.args.ctx, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("vaultClient.List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("vaultClient.List() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_KVV1_Read(t *testing.T) {
	type args struct {
		ctx  context.Context
		path string
	}
	tests := []struct {
		name     string
		prepare  func(*logical.MockLogical)
		args     args
		wantData SecretData
		wantMeta SecretMetadata
		wantErr  bool
	}{
		{
			name: "nil",
			args: args{
				ctx:  context.Background(),
				path: "",
			},
			wantErr: true,
		},
		{
			name: "query error",
			args: args{
				ctx:  context.Background(),
				path: "secrets/application/foo",
			},
			prepare: func(logical *logical.MockLogical) {
				logical.EXPECT().Read("secrets/application/foo").Return(&vaultApi.Secret{}, fmt.Errorf("foo"))
			},
			wantErr: true,
		},
		{
			name: "nil secret",
			args: args{
				ctx:  context.Background(),
				path: "secrets/application/foo",
			},
			prepare: func(logical *logical.MockLogical) {
				logical.EXPECT().Read("secrets/application/foo").Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "nil secret data",
			args: args{
				ctx:  context.Background(),
				path: "secrets/application/foo",
			},
			prepare: func(logical *logical.MockLogical) {
				logical.EXPECT().Read("secrets/application/foo").Return(&vaultApi.Secret{
					Data: nil,
				}, nil)
			},
			wantErr: true,
		},
		{
			name: "valid",
			args: args{
				ctx:  context.Background(),
				path: "secrets/application/foo",
			},
			prepare: func(logical *logical.MockLogical) {
				logical.EXPECT().Read("secrets/application/foo").Return(&vaultApi.Secret{
					Data: SecretData{
						"key": "value",
					},
				}, nil)
			},
			wantErr: false,
			wantData: SecretData{
				"key": "value",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Arm mocks
			logicalMock := logical.NewMockLogical(ctrl)

			// Prepare mocks
			if tt.prepare != nil {
				tt.prepare(logicalMock)
			}

			// Service
			underTest := V1(logicalMock, "secrets/")
			gotData, gotMeta, err := underTest.Read(tt.args.ctx, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("vaultClient.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(gotData, tt.wantData) {
				t.Errorf("vaultClient.Read() = %v, want %v", gotData, tt.wantData)
			}
			if !tt.wantErr && !reflect.DeepEqual(gotMeta, tt.wantMeta) {
				t.Errorf("vaultClient.Read() = %v, want %v", gotMeta, tt.wantMeta)
			}
		})
	}
}

func Test_vaultClient_Write(t *testing.T) {
	type args struct {
		ctx  context.Context
		path string
		data map[string]interface{}
	}
	tests := []struct {
		name    string
		prepare func(*logical.MockLogical)
		args    args
		wantErr bool
	}{
		{
			name: "blank",
			args: args{
				ctx:  context.Background(),
				path: "",
			},
			wantErr: true,
		},
		{
			name: "query error",
			args: args{
				ctx:  context.Background(),
				path: "secrets/application/foo",
			},
			prepare: func(logical *logical.MockLogical) {
				logical.EXPECT().Write("secrets/application/foo", gomock.Any()).Return(&vaultApi.Secret{}, fmt.Errorf("foo"))
			},
			wantErr: true,
		},
		{
			name: "valid",
			args: args{
				ctx:  context.Background(),
				path: "secrets/application/foo",
			},
			prepare: func(logical *logical.MockLogical) {
				logical.EXPECT().Write("secrets/application/foo", gomock.Any()).Return(&vaultApi.Secret{
					Data: SecretData{
						"key": "value",
					},
				}, nil)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Arm mocks
			logicalMock := logical.NewMockLogical(ctrl)

			// Prepare mocks
			if tt.prepare != nil {
				tt.prepare(logicalMock)
			}

			// Service
			underTest := V1(logicalMock, "secrets/")
			err := underTest.Write(tt.args.ctx, tt.args.path, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("vaultClient.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
