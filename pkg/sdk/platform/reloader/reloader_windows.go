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

// +build windows

package reloader

import (
	"context"
	"net"

	"github.com/oklog/run"

	"github.com/elastic/harp/pkg/sdk/log"
)

// UnsupportedReloader is the file descriptor reloader mock for Windows.
type UnsupportedReloader struct {
}

// Create a descriptor reloader.
func Create(ctx context.Context) Reloader {
	log.For(ctx).Warn("graceful reload is not supported on this platform")
	return &UnsupportedReloader{}
}

// Listen create a listener socket.
func (t *UnsupportedReloader) Listen(network, address string) (net.Listener, error) {
	return net.Listen(network, address)
}

// SetupGracefulRestart does nothing on Windows.
func (t *UnsupportedReloader) SetupGracefulRestart(context context.Context, group run.Group) {
	// no-op since it isn't supported
}
