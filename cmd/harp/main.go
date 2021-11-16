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
	"math/rand"
	"time"

	"github.com/elastic/harp/cmd/harp/internal/cmd"
	"github.com/elastic/harp/pkg/sdk/log"

	// Register encryption transformers
	_ "github.com/elastic/harp/pkg/sdk/value/encryption/aead"
	_ "github.com/elastic/harp/pkg/sdk/value/encryption/fernet"
	_ "github.com/elastic/harp/pkg/sdk/value/encryption/jwe"
	_ "github.com/elastic/harp/pkg/sdk/value/encryption/secretbox"
	_ "github.com/elastic/harp/pkg/vault"
)

func init() {
	// Set default timezone to UTC
	time.Local = time.UTC

	// Initialize random number generator
	rand.Seed(time.Now().Unix())
}

func main() {
	if err := cmd.Execute(); err != nil {
		log.CheckErr("Unable to complete command execution", err)
	}
}
