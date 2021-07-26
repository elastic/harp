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

//go:build !windows
// +build !windows

package reloader

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/cloudflare/tableflip"
	"github.com/oklog/run"

	"github.com/elastic/harp/pkg/sdk/log"
)

// TableflipReloader deleagtes socket reloading to tableflip library which his
// not windows compatible.
type TableflipReloader struct {
	*tableflip.Upgrader
}

// Create a descriptor reload based on tableflip.
func Create(ctx context.Context) Reloader {
	upg, _ := tableflip.New(tableflip.Options{})

	// Do an upgrade on SIGHUP
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGHUP)
		for range ch {
			log.For(ctx).Warn("Graceful reloading socket descriptor")
			_ = upg.Upgrade()
		}
	}()

	return &TableflipReloader{upg}
}

// SetupGracefulRestart arms the graceful restart handler.
func (t *TableflipReloader) SetupGracefulRestart(ctx context.Context, group run.Group) {
	ctx, cancel := context.WithCancel(ctx)

	// Register an actor, i.e. an execute and interrupt func, that
	// terminates when graceful restart is initiated and the child process
	// signals to be ready, or the parent context is canceled.
	group.Add(func() error {
		// Tell the parent we are ready
		err := t.Ready()
		if err != nil {
			return err
		}

		select {
		case <-t.Exit(): // Wait for child to be ready (or application shutdown)
			return nil

		case <-ctx.Done():
			return ctx.Err()
		}
	}, func(error) {
		cancel()
		t.Stop()
	})
}
