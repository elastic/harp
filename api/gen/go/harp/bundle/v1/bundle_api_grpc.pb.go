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

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.3
// source: harp/bundle/v1/bundle_api.proto

package bundlev1

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// BundleAPIClient is the client API for BundleAPI service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BundleAPIClient interface {
	// GetSecret returns the matching RAW secret value according to requested path.
	GetSecret(ctx context.Context, in *GetSecretRequest, opts ...grpc.CallOption) (*GetSecretResponse, error)
}

type bundleAPIClient struct {
	cc grpc.ClientConnInterface
}

func NewBundleAPIClient(cc grpc.ClientConnInterface) BundleAPIClient {
	return &bundleAPIClient{cc}
}

func (c *bundleAPIClient) GetSecret(ctx context.Context, in *GetSecretRequest, opts ...grpc.CallOption) (*GetSecretResponse, error) {
	out := new(GetSecretResponse)
	err := c.cc.Invoke(ctx, "/harp.bundle.v1.BundleAPI/GetSecret", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BundleAPIServer is the server API for BundleAPI service.
// All implementations should embed UnimplementedBundleAPIServer
// for forward compatibility
type BundleAPIServer interface {
	// GetSecret returns the matching RAW secret value according to requested path.
	GetSecret(context.Context, *GetSecretRequest) (*GetSecretResponse, error)
}

// UnimplementedBundleAPIServer should be embedded to have forward compatible implementations.
type UnimplementedBundleAPIServer struct {
}

func (UnimplementedBundleAPIServer) GetSecret(context.Context, *GetSecretRequest) (*GetSecretResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSecret not implemented")
}

// UnsafeBundleAPIServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BundleAPIServer will
// result in compilation errors.
type UnsafeBundleAPIServer interface {
	mustEmbedUnimplementedBundleAPIServer()
}

func RegisterBundleAPIServer(s grpc.ServiceRegistrar, srv BundleAPIServer) {
	s.RegisterService(&BundleAPI_ServiceDesc, srv)
}

func _BundleAPI_GetSecret_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetSecretRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BundleAPIServer).GetSecret(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/harp.bundle.v1.BundleAPI/GetSecret",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BundleAPIServer).GetSecret(ctx, req.(*GetSecretRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// BundleAPI_ServiceDesc is the grpc.ServiceDesc for BundleAPI service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var BundleAPI_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "harp.bundle.v1.BundleAPI",
	HandlerType: (*BundleAPIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetSecret",
			Handler:    _BundleAPI_GetSecret_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "harp/bundle/v1/bundle_api.proto",
}
