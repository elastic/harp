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
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle/pipeline"
	"github.com/elastic/harp/pkg/sdk/log"
)

func main() {
	var (
		// Initialize an execution context
		ctx = context.Background()
	)

	// Run the pipeline
	if err := pipeline.Run(ctx,
		pipeline.PackageProcessor(packageRemapper), // Package processor
	); err != nil {
		log.For(ctx).Fatal("unable to process bundle", zap.Error(err))
	}
}

// -----------------------------------------------------------------------------

func packageRemapper(ctx pipeline.Context, p *bundlev1.Package) error {

	// Remapping condition
	if !strings.HasPrefix(p.Name, "services/production/global/clusters/") {
		// Skip path remapping
		return nil
	}

	// Remap secret path
	p.Name = fmt.Sprintf("app/production/global/clusters/v1.0.0/bootstrap/%s", strings.TrimPrefix(p.Name, "services/production/global/clusters/"))

	// No error
	return nil
}
