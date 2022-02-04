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

package consul

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/harp/pkg/kv"
	"github.com/elastic/harp/pkg/kv/consul/mock"
)

func Test_consulDriver_withNilClient(t *testing.T) {
	underTest := Store(nil)
	assert.NotNil(t, underTest)

	kp, err := underTest.Get(context.Background(), "test")
	assert.Nil(t, kp)
	assert.Error(t, err)

	err = underTest.Put(context.Background(), "test", []byte(""))
	assert.Error(t, err)

	err = underTest.Delete(context.Background(), "test")
	assert.Error(t, err)

	kps, err := underTest.List(context.Background(), "test")
	assert.Nil(t, kps)
	assert.Error(t, err)
}

func Test_consulDriver_Get(t *testing.T) {
	type args struct {
		in0 context.Context
		key string
	}
	tests := []struct {
		name    string
		args    args
		prepare func(*mock.MockClient)
		want    *kv.Pair
		wantErr bool
	}{
		{
			name: "get error",
			args: args{
				in0: context.Background(),
				key: "application/production/test",
			},
			prepare: func(client *mock.MockClient) {
				client.EXPECT().Get("application/production/test", gomock.Any()).Return(nil, nil, fmt.Errorf("test"))
			},
			wantErr: true,
		},
		{
			name: "not found",
			args: args{
				in0: context.Background(),
				key: "application/production/test",
			},
			prepare: func(client *mock.MockClient) {
				client.EXPECT().Get("application/production/test", gomock.Any()).Return(nil, nil, nil)
			},
			wantErr: true,
		},
		// ---------------------------------------------------------------------
		{
			name: "valid",
			args: args{
				in0: context.Background(),
				key: "application/production/test",
			},
			prepare: func(client *mock.MockClient) {
				client.EXPECT().Get("application/production/test", gomock.Any()).Return(&api.KVPair{
					Key:   "application/production/test",
					Value: []byte("{}"),
				}, &api.QueryMeta{}, nil)
			},
			wantErr: false,
			want: &kv.Pair{
				Key:   "application/production/test",
				Value: []byte("{}"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Arm mocks
			consul := mock.NewMockClient(ctrl)

			// Prepare mocks
			if tt.prepare != nil {
				tt.prepare(consul)
			}

			d := &consulDriver{
				client: consul,
			}
			got, err := d.Get(tt.args.in0, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("consulDriver.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("consulDriver.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_consulDriver_Put(t *testing.T) {
	type args struct {
		in0   context.Context
		key   string
		value []byte
	}
	tests := []struct {
		name    string
		args    args
		prepare func(*mock.MockClient)
		wantErr bool
	}{
		{
			name: "put error",
			args: args{
				in0:   context.Background(),
				key:   "application/production/test",
				value: []byte("{}"),
			},
			prepare: func(client *mock.MockClient) {
				client.EXPECT().Put(gomock.Any(), nil).Return(nil, fmt.Errorf("test"))
			},
			wantErr: true,
		},
		// ---------------------------------------------------------------------
		{
			name: "valid",
			args: args{
				in0:   context.Background(),
				key:   "application/production/test",
				value: []byte("{}"),
			},
			prepare: func(client *mock.MockClient) {
				client.EXPECT().Put(gomock.Any(), nil).Return(&api.WriteMeta{}, nil)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Arm mocks
			consul := mock.NewMockClient(ctrl)

			// Prepare mocks
			if tt.prepare != nil {
				tt.prepare(consul)
			}

			d := &consulDriver{
				client: consul,
			}
			err := d.Put(tt.args.in0, tt.args.key, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("consulDriver.Put() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_consulDriver_Delete(t *testing.T) {
	type args struct {
		in0 context.Context
		key string
	}
	tests := []struct {
		name    string
		args    args
		prepare func(*mock.MockClient)
		wantErr bool
	}{
		{
			name: "not found",
			args: args{
				in0: context.Background(),
				key: "application/production/test",
			},
			prepare: func(client *mock.MockClient) {
				client.EXPECT().Get("application/production/test", gomock.Any()).Return(nil, nil, nil)
			},
			wantErr: true,
		},
		{
			name: "get error",
			args: args{
				in0: context.Background(),
				key: "application/production/test",
			},
			prepare: func(client *mock.MockClient) {
				client.EXPECT().Get("application/production/test", gomock.Any()).Return(nil, nil, fmt.Errorf("test"))
			},
			wantErr: true,
		},
		{
			name: "put error",
			args: args{
				in0: context.Background(),
				key: "application/production/test",
			},
			prepare: func(client *mock.MockClient) {
				client.EXPECT().Get("application/production/test", gomock.Any()).Return(&api.KVPair{}, &api.QueryMeta{}, nil)
				client.EXPECT().Delete(gomock.Any(), nil).Return(nil, fmt.Errorf("test"))
			},
			wantErr: true,
		},
		// ---------------------------------------------------------------------
		{
			name: "valid",
			args: args{
				in0: context.Background(),
				key: "application/production/test",
			},
			prepare: func(client *mock.MockClient) {
				client.EXPECT().Get("application/production/test", gomock.Any()).Return(&api.KVPair{}, &api.QueryMeta{}, nil)
				client.EXPECT().Delete(gomock.Any(), nil).Return(&api.WriteMeta{}, nil)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Arm mocks
			consul := mock.NewMockClient(ctrl)

			// Prepare mocks
			if tt.prepare != nil {
				tt.prepare(consul)
			}

			d := &consulDriver{
				client: consul,
			}
			err := d.Delete(tt.args.in0, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("consulDriver.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_consulDriver_Exists(t *testing.T) {
	type args struct {
		in0 context.Context
		key string
	}
	tests := []struct {
		name    string
		args    args
		prepare func(*mock.MockClient)
		want    bool
		wantErr bool
	}{
		{
			name: "get error",
			args: args{
				in0: context.Background(),
				key: "application/production/test",
			},
			prepare: func(client *mock.MockClient) {
				client.EXPECT().Get("application/production/test", gomock.Any()).Return(nil, nil, fmt.Errorf("test"))
			},
			wantErr: true,
		},
		// ---------------------------------------------------------------------
		{
			name: "not found",
			args: args{
				in0: context.Background(),
				key: "application/production/test",
			},
			prepare: func(client *mock.MockClient) {
				client.EXPECT().Get("application/production/test", gomock.Any()).Return(nil, nil, nil)
			},
			wantErr: false,
			want:    false,
		},
		{
			name: "valid",
			args: args{
				in0: context.Background(),
				key: "application/production/test",
			},
			prepare: func(client *mock.MockClient) {
				client.EXPECT().Get("application/production/test", gomock.Any()).Return(&api.KVPair{
					Key:   "application/production/test",
					Value: []byte("{}"),
				}, &api.QueryMeta{}, nil)
			},
			wantErr: false,
			want:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Arm mocks
			consul := mock.NewMockClient(ctrl)

			// Prepare mocks
			if tt.prepare != nil {
				tt.prepare(consul)
			}

			d := &consulDriver{
				client: consul,
			}
			got, err := d.Exists(tt.args.in0, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("consulDriver.Exists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("consulDriver.Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_consulDriver_List(t *testing.T) {
	type args struct {
		in0 context.Context
		key string
	}
	tests := []struct {
		name    string
		args    args
		prepare func(*mock.MockClient)
		want    []*kv.Pair
		wantErr bool
	}{
		{
			name: "list error",
			args: args{
				in0: context.Background(),
				key: "application/production",
			},
			prepare: func(client *mock.MockClient) {
				client.EXPECT().List("application/production", nil).Return(nil, nil, fmt.Errorf("test"))
			},
			wantErr: true,
		},
		{
			name: "empty result",
			args: args{
				in0: context.Background(),
				key: "application/production",
			},
			prepare: func(client *mock.MockClient) {
				client.EXPECT().List("application/production", nil).Return([]*api.KVPair{}, &api.QueryMeta{}, nil)
			},
			wantErr: true,
		},
		// ---------------------------------------------------------------------
		{
			name: "valid",
			args: args{
				in0: context.Background(),
				key: "application/production",
			},
			prepare: func(client *mock.MockClient) {
				client.EXPECT().List("application/production", nil).Return([]*api.KVPair{
					{
						Key: "application/production",
					},
					{
						Key:   "application/production/test",
						Value: []byte("{}"),
					},
				}, &api.QueryMeta{}, nil)
			},
			wantErr: false,
			want: []*kv.Pair{
				{
					Key:   "application/production/test",
					Value: []byte("{}"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Arm mocks
			consul := mock.NewMockClient(ctrl)

			// Prepare mocks
			if tt.prepare != nil {
				tt.prepare(consul)
			}

			d := &consulDriver{
				client: consul,
			}
			got, err := d.List(tt.args.in0, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("consulDriver.List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("consulDriver.List() = %s", diff)
			}
		})
	}
}

func Test_consulDriver_Close(t *testing.T) {
	underTest := Store(nil)
	assert.NotNil(t, underTest)
	err := underTest.Close()
	assert.NoError(t, err)
}
