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

package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

// Consul creates a test consul server inside a Docker container.
func Consul(ctx context.Context, tb testing.TB) string {
	pool, err := dockertest.NewPool("")
	if err != nil {
		tb.Fatalf("couldn't connect to docker: %v", err)
		return ""
	}
	pool.MaxWait = 10 * time.Second

	// Prepare bootstrap configuration
	config := struct {
		Datacenter       string `json:"datacenter,omitempty"`
		ACLDatacenter    string `json:"acl_datacenter,omitempty"`
		ACLDefaultPolicy string `json:"acl_default_policy,omitempty"`
		ACLMasterToken   string `json:"acl_master_token,omitempty"`
	}{
		Datacenter:       "test",
		ACLDatacenter:    "test",
		ACLDefaultPolicy: "deny",
		ACLMasterToken:   "test",
	}

	// Encode configuration as JSON
	encodedConfig, errConfig := json.Marshal(config)
	if errConfig != nil {
		tb.Fatalf("couldn't serialize configuration as json: %v", errConfig)
		return ""
	}

	// Start zookeeper server
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "consul",
		Tag:        "1.10.3",
		Cmd:        []string{"agent", "-dev", "-client", "0.0.0.0"},
		Env:        []string{fmt.Sprintf("CONSUL_LOCAL_CONFIG=%s", encodedConfig)},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		tb.Fatalf("couldn't start resource: %v", err)
		return ""
	}

	// Set expiration
	if err := resource.Expire(15 * 60); err != nil {
		tb.Error("unable to set expiration value for the container")
	}

	// Cleanup function
	tb.Cleanup(func() {
		if err := pool.Purge(resource); err != nil {
			tb.Errorf("couldn't purge container: %v", err)
			return
		}
	})

	consulURI := fmt.Sprintf("localhost:%s", resource.GetPort("8500/tcp"))

	// Wait until connection is ready
	if err := pool.Retry(func() (err error) {
		config := api.DefaultConfig()
		config.Address = consulURI
		config.Token = "test"

		// Create client instance.
		client, err := api.NewClient(config)
		if err != nil {
			return fmt.Errorf("unable to connect to the server: %w", err)
		}

		// Try to write data.
		_, err = client.KV().Put(&api.KVPair{
			Key:   "ready",
			Value: []byte("ready"),
		}, nil)

		// Check connection state
		return err
	}); err != nil {
		tb.Fatalf("zk server never ready: %v", err)
		return ""
	}

	// Return connection uri
	return consulURI
}
