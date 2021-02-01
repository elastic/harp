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

package vault

import (
	"fmt"

	"github.com/hashicorp/vault/api"
)

// -----------------------------------------------------------------------------

// CheckAuthentication verifies that the connection to vault is setup correctly
// by retrieving information about the configured token
func CheckAuthentication(client *api.Client) ([]string, error) {
	tokenInfo, tokenErr := client.Auth().Token().LookupSelf()
	if tokenErr != nil {
		return nil, fmt.Errorf("error connecting to vault: %w", tokenErr)
	}

	tokenPolicies, polErr := tokenInfo.TokenPolicies()
	if polErr != nil {
		return nil, fmt.Errorf("error looking up token policies: %w", tokenErr)
	}

	// No error
	return tokenPolicies, nil
}
