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

	"github.com/elastic/harp/pkg/vault/cubbyhole"
	"github.com/elastic/harp/pkg/vault/kv"
	"github.com/elastic/harp/pkg/vault/transit"
)

// -----------------------------------------------------------------------------

// ServiceFactory defines Vault client cervice contract.
type ServiceFactory interface {
	KV(mountPath string) (kv.Service, error)
	Transit(mounthPath, keyName string) (transit.Service, error)
	Cubbyhole(mountPath string) (cubbyhole.Service, error)
}

// -----------------------------------------------------------------------------

// DefaultClient initialize a Vault client and wrap it in a Service factory.
func DefaultClient() (ServiceFactory, error) {
	// Initialize default config
	conf := api.DefaultConfig()

	// Initialize vault client
	vaultClient, err := api.NewClient(conf)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize vault client: %w", err)
	}

	// Delegate to other constructor.
	return FromVaultClient(vaultClient)
}

// FromVaultClient wraps an existing Vault client as a Service factory.
func FromVaultClient(vaultClient *api.Client) (ServiceFactory, error) {
	// Return wrapped client.
	return &client{
		Client: vaultClient,
	}, nil
}

// -----------------------------------------------------------------------------

// Client wrpas original Vault client instance to provide service factory.
type client struct {
	*api.Client
}

func (c *client) KV(mountPath string) (kv.Service, error) {
	return kv.New(c.Client, mountPath)
}

func (c *client) Transit(mountPath, keyName string) (transit.Service, error) {
	return transit.New(c.Client, mountPath, keyName)
}

func (c *client) Cubbyhole(mountPath string) (cubbyhole.Service, error) {
	return cubbyhole.New(c.Client, mountPath)
}
