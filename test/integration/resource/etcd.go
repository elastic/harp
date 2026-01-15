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
	"fmt"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Etcd creates a test etcd server inside a Docker container.
func Etcd(ctx context.Context, tb testing.TB) string {
	pool, err := dockertest.NewPool("")
	if err != nil {
		tb.Fatalf("couldn't connect to docker: %v", err)
		return ""
	}
	pool.MaxWait = 10 * time.Second

	// Start zookeeper server
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "quay.io/coreos/etcd",
		Tag:        "v3.5.1",
		Cmd: []string{
			"/usr/local/bin/etcd",
			"--data-dir=/etcd-data",
			"--name=node1",
			"--initial-advertise-peer-urls=http://0.0.0.0:2380",
			"--listen-peer-urls=http://0.0.0.0:2380",
			"--advertise-client-urls=http://0.0.0.0:2379",
			"--listen-client-urls=http://0.0.0.0:2379",
			"--initial-cluster=node1=http://0.0.0.0:2380",
		},
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

	etcURI := fmt.Sprintf("http://127.0.0.1:%s", resource.GetPort("2379/tcp"))

	// Wait until connection is ready
	if err := pool.Retry(func() (err error) {
		if _, errClient := clientv3.New(clientv3.Config{
			Endpoints:   []string{etcURI},
			DialTimeout: 5 * time.Second,
		}); errClient != nil {
			return fmt.Errorf("unable to connect to etcd3 server: %w", errClient)
		}

		// Check connection state
		return nil
	}); err != nil {
		tb.Fatalf("zk server never ready: %v", err)
		return ""
	}

	// Return connection uri
	return etcURI
}
