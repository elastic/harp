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

package vfs

import (
	"io/ioutil"
	"testing"

	"github.com/davecgh/go-spew/spew"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle/secret"
)

func mustPack(in []byte) []byte {
	out, err := secret.Pack(in)
	if err != nil {
		panic(err)
	}
	return out
}

var testBundle = &bundlev1.Bundle{
	Version: 1,
	Packages: []*bundlev1.Package{
		{
			Name: "infra/aws/ecsecurity/eu-central-1/rds/rundeck/admin_creds",
			Secrets: &bundlev1.SecretChain{
				Version: 0,
				Data: []*bundlev1.KV{
					{
						Key:   "database_root_password",
						Type:  "string",
						Value: mustPack([]byte("foo")),
					},
				},
			},
		},
	},
}

func TestBundle_FS_Initialization(t *testing.T) {
	fs, err := FromBundle(testBundle)
	if err != nil {
		t.Errorf("unable to initialize filesystem : %v", err)
		return
	}

	f, err := fs.Open("/infra/aws/ecsecurity/eu-central-1/rds/rundeck/admin_creds")
	if err != nil {
		t.Errorf("unable to open file from filesystem : %v", err)
		return
	}

	payload, err := ioutil.ReadAll(f)
	if err != nil {
		t.Errorf("unable to read file from filesystem : %v", err)
		return
	}

	spew.Dump(payload)
}
