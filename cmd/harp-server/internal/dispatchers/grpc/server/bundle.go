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

package server

import (
	"context"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/server/manager"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Bundle returns an gRPC requests handler for BundleAPI.
func Bundle(bm manager.Backend) bundlev1.BundleAPIServer {
	return &grpcBundleServer{
		bm: bm,
	}
}

type grpcBundleServer struct {
	bundlev1.UnimplementedBundleAPIServer
	bm manager.Backend
}

func (s *grpcBundleServer) GetSecret(ctx context.Context, req *bundlev1.GetSecretRequest) (*bundlev1.GetSecretResponse, error) {
	// Check arguments
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "request is nil")
	}
	if req.Namespace == "" {
		return nil, status.Errorf(codes.InvalidArgument, "namespace could not be blank")
	}
	if req.Path == "" {
		return nil, status.Errorf(codes.InvalidArgument, "path could not be blank")
	}

	// Delegate to engine to retrieve secret
	content, err := s.bm.GetSecret(ctx, req.Namespace, req.Path)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Secret '%s' could not be retrieved from '%s' namespace", req.Path, req.Namespace)
	}

	// Return result
	return &bundlev1.GetSecretResponse{
		Namespace: req.Namespace,
		Path:      req.Path,
		Content:   content,
	}, nil
}
