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

package cmdutil

import (
	"context"

	"github.com/gosimple/slug"

	"github.com/elastic/harp/build/version"
	"github.com/elastic/harp/pkg/sdk/log"
)

// Context initializes a command context.
func Context(ctx context.Context, name string, debug bool, logLevel string) (context.Context, context.CancelFunc) {
	// Context to attach all goroutines
	ctx, cancel := context.WithCancel(ctx)

	// Initialize logger
	log.Setup(ctx,
		&log.Options{
			Debug:    debug,
			LogLevel: logLevel,
			AppName:  slug.Make(name),
			AppID:    version.ID(),
			Version:  version.Version,
			Revision: version.Revision,
		},
	)

	// Return context
	return ctx, cancel
}
