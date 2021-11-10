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

package main

import (
	"fmt"
	"os"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle"
	"github.com/elastic/harp/pkg/bundle/secret"
)

func main() {
	b := &bundlev1.Bundle{
		Packages: []*bundlev1.Package{},
	}

	// Create 25000 packages
	for i := 0; i < 25000; i++ {
		p := &bundlev1.Package{
			Name: fmt.Sprintf("app/secret/large-bundle/%d", i),
			Secrets: &bundlev1.SecretChain{
				Data: []*bundlev1.KV{},
			},
		}

		for j := 0; j < 100; j++ {
			p.Secrets.Data = append(p.Secrets.Data, &bundlev1.KV{
				Key:   fmt.Sprintf("secret-%d", j),
				Value: secret.MustPack("test-value"),
			})
		}

		b.Packages = append(b.Packages, p)
	}

	// Save as a container in Stdout.
	if err := bundle.ToContainerWriter(os.Stdout, b); err != nil {
		panic(err)
	}
}
