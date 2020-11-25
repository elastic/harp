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
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"

	csov1 "github.com/elastic/harp/api/gen/go/cso/v1"
)

func TestCso_Interpret_Text(t *testing.T) {
	txtTemplates := Text()

	testCases := []struct {
		desc     string
		input    *csov1.Secret
		wantErr  bool
		expected string
	}{
		{
			desc:    "Meta path",
			wantErr: false,
			input: &csov1.Secret{
				RingLevel: csov1.RingLevel_RING_LEVEL_META,
				Path: &csov1.Secret_Meta{
					Meta: &csov1.Meta{
						Key: "cso/revision",
					},
				},
			},
			expected: `Give me a secret named "cso/revision"`,
		},
		{
			desc:    "Infra path",
			wantErr: false,
			input: &csov1.Secret{
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
			expected: `Give me an infrastructure secret named "adminconsole/root_creds", for "aws" cloud provider, concerning account "security", located in region "us-east-1", for service "rds".`,
		},
		{
			desc:    "Platform path",
			wantErr: false,
			input: &csov1.Secret{
				RingLevel: csov1.RingLevel_RING_LEVEL_PLATFORM,
				Path: &csov1.Secret_Platform{
					Platform: &csov1.Platform{
						Stage:       csov1.QualityLevel_QUALITY_LEVEL_INVALID,
						Name:        "customer-1",
						Region:      "us-east-1",
						ServiceName: "adminconsole",
						Key:         "database/creds",
					},
				},
			},
			expected: `Give me a platform secret named "database/creds", for a service named "adminconsole", located in region "us-east-1", part of a "QUALITY_LEVEL_INVALID" platform named "customer-1".`,
		},
		{
			desc:    "Product path",
			wantErr: false,
			input: &csov1.Secret{
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
			expected: `Give me a product secret named "http/jwt_hmac", concerning the component "server", for a product named "ece", in version "v1.0.0".`,
		},
		{
			desc:    "Application path",
			wantErr: false,
			input: &csov1.Secret{
				RingLevel: csov1.RingLevel_RING_LEVEL_APPLICATION,
				Path: &csov1.Secret_Application{
					Application: &csov1.Application{
						Stage:          csov1.QualityLevel_QUALITY_LEVEL_INVALID,
						PlatformName:   "customer-1",
						ProductName:    "ece",
						ProductVersion: "v1.0.0",
						ComponentName:  "server",
						Key:            "http/jwt_hmac",
					},
				},
			},
			expected: `Give me an application secret named "http/jwt_hmac", concerning the component "server", for a product named "ece", in version "v1.0.0", running on a "QUALITY_LEVEL_INVALID" platform named "customer-1".`,
		},
		{
			desc:    "Artifact path",
			wantErr: false,
			input: &csov1.Secret{
				RingLevel: csov1.RingLevel_RING_LEVEL_ARTIFACT,
				Path: &csov1.Secret_Artifact{
					Artifact: &csov1.Artifact{
						Type: "docker",
						Id:   "sha256:fab2dded59dd0c2894dd9dbae71418f565be5bd0d8fd82365c16aec41c7e367f",
						Key:  "attestations/snyk_report",
					},
				},
			},
			expected: `Give me an artifact secret named "attestations/snyk_report", concerning the "docker" artifact with ID "sha256:fab2dded59dd0c2894dd9dbae71418f565be5bd0d8fd82365c16aec41c7e367f".`,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			var buf bytes.Buffer
			err := Interpret(tC.input, txtTemplates, &buf)
			if (err != nil) != tC.wantErr {
				t.Errorf("error: got %v, but not error expected", err)
			}
			if tC.wantErr {
				return
			}

			got := buf.String()
			if diff := cmp.Diff(got, tC.expected); diff != "" {
				t.Errorf("%q. Interpret():\n-got/+want\ndiff %s", tC.desc, diff)
			}
		})
	}
}
