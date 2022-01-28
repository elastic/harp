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

package v1

import (
	"testing"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	csov1 "github.com/elastic/harp/api/gen/go/cso/v1"
)

var cmpOpts = []cmp.Option{
	cmpopts.IgnoreUnexported(wrappers.StringValue{}),
	cmpopts.IgnoreUnexported(csov1.Secret{}),
	cmpopts.IgnoreUnexported(csov1.Meta{}),
	cmpopts.IgnoreUnexported(csov1.Secret_Meta{}),
	cmpopts.IgnoreUnexported(csov1.Secret_Infrastructure{}),
	cmpopts.IgnoreUnexported(csov1.Secret_Platform{}),
	cmpopts.IgnoreUnexported(csov1.Secret_Product{}),
	cmpopts.IgnoreUnexported(csov1.Secret_Application{}),
	cmpopts.IgnoreUnexported(csov1.Secret_Artifact{}),
	cmpopts.IgnoreUnexported(csov1.Infrastructure{}),
	cmpopts.IgnoreUnexported(csov1.Platform{}),
	cmpopts.IgnoreUnexported(csov1.Product{}),
	cmpopts.IgnoreUnexported(csov1.Application{}),
	cmpopts.IgnoreUnexported(csov1.Artifact{}),
	cmp.FilterPath(
		func(p cmp.Path) bool {
			return p.String() == "Value"
		}, cmp.Ignore()),
}

func TestCso_Pack(t *testing.T) {
	testCases := []struct {
		desc     string
		path     string
		value    interface{}
		expected *csov1.Secret
		wantErr  bool
	}{
		{
			desc:    "Empty path",
			path:    "",
			wantErr: true,
		},
		{
			desc:    "Invalid path",
			path:    "/toto",
			wantErr: true,
		},
		{
			desc: "Meta path",
			path: "meta/cso/revision",
			value: map[string]interface{}{
				"rev": "6",
			},
			wantErr: false,
			expected: &csov1.Secret{
				RingLevel: csov1.RingLevel_RING_LEVEL_META,
				Path: &csov1.Secret_Meta{
					Meta: &csov1.Meta{
						Key: "cso/revision",
					},
				},
			},
		},
		{
			desc: "Infra path",
			path: "infra/aws/security/us-east-1/rds/adminconsole/root_creds",
			value: map[string]interface{}{
				"user":     "foo",
				"password": "bar",
			},
			wantErr: false,
			expected: &csov1.Secret{
				RingLevel: csov1.RingLevel_RING_LEVEL_INFRASTRUCTURE,
				Path: &csov1.Secret_Infrastructure{
					Infrastructure: &csov1.Infrastructure{
						CloudProvider: "aws",
						AccountId:     "security",
						Region:        "us-east-1",
						ServiceName:   "rds",
						Key:           "adminconsole/root_creds",
					},
				},
			},
		},
		{
			desc: "Platform path",
			path: "platform/dev/customer-1/us-east-1/adminconsole/database/creds",
			value: map[string]interface{}{
				"user":     "foo",
				"password": "bar",
			},
			wantErr: false,
			expected: &csov1.Secret{
				RingLevel: csov1.RingLevel_RING_LEVEL_PLATFORM,
				Path: &csov1.Secret_Platform{
					Platform: &csov1.Platform{
						Stage:       csov1.QualityLevel_QUALITY_LEVEL_DEV,
						Name:        "customer-1",
						Region:      "us-east-1",
						ServiceName: "adminconsole",
						Key:         "database/creds",
					},
				},
			},
		},
		{
			desc: "Product path",
			path: "product/ece/v1.0.0/server/http/jwt_hmac",
			value: map[string]interface{}{
				"user":     "foo",
				"password": "bar",
			},
			wantErr: false,
			expected: &csov1.Secret{
				RingLevel: csov1.RingLevel_RING_LEVEL_PRODUCT,
				Path: &csov1.Secret_Product{
					Product: &csov1.Product{
						Name:          "ece",
						Version:       "v1.0.0",
						ComponentName: "server",
						Key:           "http/jwt_hmac",
					},
				},
			},
		},
		{
			desc: "Application path",
			path: "app/dev/customer-1/ece/v1.0.0/server/http/jwt_hmac",
			value: map[string]interface{}{
				"user":     "foo",
				"password": "bar",
			},
			wantErr: false,
			expected: &csov1.Secret{
				RingLevel: csov1.RingLevel_RING_LEVEL_APPLICATION,
				Path: &csov1.Secret_Application{
					Application: &csov1.Application{
						Stage:          csov1.QualityLevel_QUALITY_LEVEL_DEV,
						PlatformName:   "customer-1",
						ProductName:    "ece",
						ProductVersion: "v1.0.0",
						ComponentName:  "server",
						Key:            "http/jwt_hmac",
					},
				},
			},
		},
		{
			desc: "Artifact path",
			path: "artifact/docker/sha256:fab2dded59dd0c2894dd9dbae71418f565be5bd0d8fd82365c16aec41c7e367f/attestations/snyk_report",
			value: map[string]interface{}{
				"user":     "foo",
				"password": "bar",
			},
			wantErr: false,
			expected: &csov1.Secret{
				RingLevel: csov1.RingLevel_RING_LEVEL_ARTIFACT,
				Path: &csov1.Secret_Artifact{
					Artifact: &csov1.Artifact{
						Type: "docker",
						Id:   "sha256:fab2dded59dd0c2894dd9dbae71418f565be5bd0d8fd82365c16aec41c7e367f",
						Key:  "attestations/snyk_report",
					},
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got, err := Pack(tC.path)
			if (err != nil) != tC.wantErr {
				t.Errorf("error: got %v, but not error expected", err)
			}
			if tC.wantErr {
				return
			}
			if diff := cmp.Diff(got, tC.expected, cmpOpts...); diff != "" {
				t.Errorf("%q. Pack():\n-got/+want\ndiff %s", tC.desc, diff)
			}
		})
	}
}
