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
	"bytes"
	"io"
	"os"
	"testing"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestFromDump(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    *bundlev1.Bundle
		wantErr bool
	}{
		{
			name:    "nil",
			wantErr: true,
		},
		{
			name: "nil reader",
			args: args{
				r: nil,
			},
			wantErr: true,
		},
		{
			name: "closed reader",
			args: args{
				r: func() io.Reader {
					f, err := os.Open("../../test/fixtures/bundles/empty.json")
					assert.NoError(t, err)
					f.Close()
					return f
				}(),
			},
			wantErr: true,
		},
		{
			name: "proto unmarshal error",
			args: args{
				r: bytes.NewReader([]byte{0x00}),
			},
			wantErr: true,
		},
		// ---------------------------------------------------------------------
		{
			name: "empty",
			args: args{
				r: func() io.Reader {
					f, err := os.Open("../../test/fixtures/bundles/empty.json")
					assert.NoError(t, err)
					return f
				}(),
			},
			wantErr: false,
			want:    &bundlev1.Bundle{},
		},
		{
			name: "complete",
			args: args{
				r: func() io.Reader {
					f, err := os.Open("../../test/fixtures/bundles/complete.json")
					assert.NoError(t, err)
					return f
				}(),
			},
			wantErr: false,
			want: &bundlev1.Bundle{
				Labels: map[string]string{
					"test": "true",
				},
				Annotations: map[string]string{
					"harp.elastic.co/v1/testing#bundlePurpose": "test",
				},
				Packages: []*bundlev1.Package{
					{
						Labels: map[string]string{
							"okta": "true",
						},
						Annotations: map[string]string{
							"infosec.elastic.co/v1/SecretPolicy#rotationMethod": "rundeck",
							"infosec.elastic.co/v1/SecretPolicy#rotationPeriod": "180d",
							"infosec.elastic.co/v1/SecretPolicy#serviceType":    "authentication",
							"infosec.elastic.co/v1/SecretPolicy#severity":       "high",
							"infra.elastic.co/v1/Rundeck#jobName":               "rotate-adminconsole-okta-api-key",
							"harp.elastic.co/v1/package#encryptionKeyAlias":     "test",
						},
						Name: "app/production/customer1/ece/v1.0.0/adminconsole/authentication/otp/okta_api_key",
						Secrets: &bundlev1.SecretChain{
							Labels: map[string]string{
								"vendor": "true",
							},
							Annotations: map[string]string{
								"creationDate": "1636452457",
								"description":  "Okta API Key for OTP validation",
								"template":     "{\n  \"API_KEY\": \"{{ .Values.vendor.okta.api_key }}\"\n}",
							},
							Data: []*bundlev1.KV{
								{
									Key:   "API_KEY",
									Type:  "string",
									Value: []byte("0\x1b\x02\x01\x01\x13\x16okta-foo-api-123456789"),
								},
							},
						},
					},
					{
						Labels: map[string]string{
							"database": "postgresql",
						},
						Annotations: map[string]string{
							"infosec.elastic.co/v1/SecretPolicy#rotationPeriod": "on-new-version",
						},
						Name: "app/production/customer1/ece/v1.0.0/adminconsole/database/usage_credentials",
						Secrets: &bundlev1.SecretChain{
							Data: []*bundlev1.KV{
								{
									Key:   "host",
									Type:  "string",
									Value: []byte("0=\x02\x01\x01\x138sample-instance.abc2defghije.us-west-2.rds.amazonaws.com"),
								},
								{
									Key:   "port",
									Type:  "string",
									Value: []byte("0\t\x02\x01\x01\x13\x045432"),
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromDump(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromDump() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got, ignoreOpts...); diff != "" {
				t.Errorf("%q. Bundle.FromDump():\n-got/+want\ndiff %s", tt.name, diff)
			}
		})
	}
}
