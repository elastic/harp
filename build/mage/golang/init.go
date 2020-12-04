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

package golang

import (
	"os"
	"runtime"
	"time"

	"github.com/fatih/color"

	"github.com/elastic/harp/pkg/sdk/types"
)

// Keep only last 2 versions
var goVersions = []string{
	"go1.15.5",
	"go1.15.6",
}

func init() {
	// Set default timezone to UTC
	time.Local = time.UTC

	if !types.StringArray(goVersions).Contains(runtime.Version()) {
		color.HiRed("#############################################################################################")
		color.HiRed("")
		color.HiRed("Your golang compiler (%s) must be updated to %s to successfully compile all tools.", runtime.Version(), goVersions)
		color.HiRed("")
		color.HiRed("#############################################################################################")
		os.Exit(-1)
	}
}
