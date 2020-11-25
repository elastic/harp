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

import "testing"

func Test_RingPath_Meta(t *testing.T) {
	testCases := []struct {
		desc     string
		input    []string
		expected string
		wantErr  bool
	}{
		{
			desc:    "nil",
			input:   nil,
			wantErr: true,
		},
		{
			desc:    "empty",
			input:   []string{},
			wantErr: true,
		},
		{
			desc:    "not enough items",
			input:   []string{""},
			wantErr: true,
		},
		{
			desc:     "meta/vault/authentication/oidc/okta/client_credentials",
			input:    []string{"vault", "authentication", "oidc", "okta", "client_credentials"},
			expected: "meta/vault/authentication/oidc/okta/client_credentials",
			wantErr:  false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got, err := RingMeta.Path(tC.input...)
			if tC.wantErr != (err != nil) {
				t.Errorf("unexpected error, got : %v", err)
			}
			if tC.wantErr {
				return
			}
			if got != tC.expected {
				t.Errorf("expected '%s', got '%s'", tC.expected, got)
			}
		})
	}
}

func Test_RingPath_Infra(t *testing.T) {
	testCases := []struct {
		desc     string
		input    []string
		expected string
		wantErr  bool
	}{
		{
			desc:    "nil",
			input:   nil,
			wantErr: true,
		},
		{
			desc:    "empty",
			input:   []string{},
			wantErr: true,
		},
		{
			desc:    "not enough items",
			input:   []string{""},
			wantErr: true,
		},
		{
			desc:    "invalid cso path",
			input:   []string{"aws", "ecsecurity", "rds", "adminconsole", "accounts", "root_admin"},
			wantErr: true,
		},
		{
			desc:     "valid cso path",
			input:    []string{"aws", "ecsecurity", "us-east-1", "rds", "adminconsole", "accounts", "root_admin"},
			expected: "infra/aws/ecsecurity/us-east-1/rds/adminconsole/accounts/root_admin",
			wantErr:  false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got, err := RingInfra.Path(tC.input...)
			if tC.wantErr != (err != nil) {
				t.Errorf("unexpected error, got : %v", err)
			}
			if tC.wantErr {
				return
			}
			if got != tC.expected {
				t.Errorf("expected '%s', got '%s'", tC.expected, got)
			}
		})
	}
}

func Test_RingPath_Platform(t *testing.T) {
	testCases := []struct {
		desc     string
		input    []string
		expected string
		wantErr  bool
	}{
		{
			desc:    "nil",
			input:   nil,
			wantErr: true,
		},
		{
			desc:    "empty",
			input:   []string{},
			wantErr: true,
		},
		{
			desc:    "not enough items",
			input:   []string{""},
			wantErr: true,
		},
		{
			desc:    "invalid cso path",
			input:   []string{"invalid", "customer-1", "eu-central-1", "database", "accounts", "billing_account"},
			wantErr: true,
		},
		{
			desc:     "valid cso path",
			input:    []string{"production", "customer-1", "eu-central-1", "database", "accounts", "billing_account"},
			expected: "platform/production/customer-1/eu-central-1/database/accounts/billing_account",
			wantErr:  false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got, err := RingPlatform.Path(tC.input...)
			if tC.wantErr != (err != nil) {
				t.Errorf("unexpected error, got : %v", err)
			}
			if tC.wantErr {
				return
			}
			if got != tC.expected {
				t.Errorf("expected '%s', got '%s'", tC.expected, got)
			}
		})
	}
}

func Test_RingPath_Product(t *testing.T) {
	testCases := []struct {
		desc     string
		input    []string
		expected string
		wantErr  bool
	}{
		{
			desc:    "nil",
			input:   nil,
			wantErr: true,
		},
		{
			desc:    "empty",
			input:   []string{},
			wantErr: true,
		},
		{
			desc:    "not enough items",
			input:   []string{""},
			wantErr: true,
		},
		{
			desc:    "invalid cso path",
			input:   []string{"eck", "invalid", "licensing", "private_signing_key"},
			wantErr: true,
		},
		{
			desc:     "valid cso path (with semver prefix)",
			input:    []string{"eck", "v1.0.0", "licensing", "private_signing_key"},
			expected: "product/eck/v1.0.0/licensing/private_signing_key",
			wantErr:  false,
		},
		{
			desc:     "valid cso path",
			input:    []string{"eck", "1.0.0", "licensing", "private_signing_key"},
			expected: "product/eck/1.0.0/licensing/private_signing_key",
			wantErr:  false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got, err := RingProduct.Path(tC.input...)
			if tC.wantErr != (err != nil) {
				t.Errorf("unexpected error, got : %v", err)
			}
			if tC.wantErr {
				return
			}
			if got != tC.expected {
				t.Errorf("expected '%s', got '%s'", tC.expected, got)
			}
		})
	}
}

func Test_RingPath_Application(t *testing.T) {
	testCases := []struct {
		desc     string
		input    []string
		expected string
		wantErr  bool
	}{
		{
			desc:    "nil",
			input:   nil,
			wantErr: true,
		},
		{
			desc:    "empty",
			input:   []string{},
			wantErr: true,
		},
		{
			desc:    "not enough items",
			input:   []string{""},
			wantErr: true,
		},
		{
			desc:    "invalid cso path",
			input:   []string{"invalid", "customer-1", "eck", "v1.0.0", "authentication", "http", "cookie_hmac_seed"},
			wantErr: true,
		},
		{
			desc:     "valid cso path",
			input:    []string{"production", "customer-1", "eck", "v1.0.0", "authentication", "http", "cookie_hmac_seed"},
			expected: "app/production/customer-1/eck/v1.0.0/authentication/http/cookie_hmac_seed",
			wantErr:  false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got, err := RingApplication.Path(tC.input...)
			if tC.wantErr != (err != nil) {
				t.Errorf("unexpected error, got : %v", err)
			}
			if tC.wantErr {
				return
			}
			if got != tC.expected {
				t.Errorf("expected '%s', got '%s'", tC.expected, got)
			}
		})
	}
}

func Test_RingPath_Artifact(t *testing.T) {
	testCases := []struct {
		desc     string
		input    []string
		expected string
		wantErr  bool
	}{
		{
			desc:    "nil",
			input:   nil,
			wantErr: true,
		},
		{
			desc:    "empty",
			input:   []string{},
			wantErr: true,
		},
		{
			desc:    "not enough items",
			input:   []string{""},
			wantErr: true,
		},
		{
			desc:     "valid cso path",
			input:    []string{"docker", "123456789", "attestations", "snyk_report"},
			expected: "artifact/docker/123456789/attestations/snyk_report",
			wantErr:  false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got, err := RingArtifact.Path(tC.input...)
			if tC.wantErr != (err != nil) {
				t.Errorf("unexpected error, got : %v", err)
			}
			if tC.wantErr {
				return
			}
			if got != tC.expected {
				t.Errorf("expected '%s', got '%s'", tC.expected, got)
			}
		})
	}
}
