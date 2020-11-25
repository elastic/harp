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

package cmd

import (
	"fmt"
	"net/url"

	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"

	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
)

// -----------------------------------------------------------------------------

func bugCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bug",
		Short: "start a bug report",
		Long: `
	Bug opens the default browser and starts a new bug report.
	The report includes useful system information.
		`,
		Run: runBug,
	}
}

func runBug(cmd *cobra.Command, args []string) {
	ctx, cancel := cmdutil.Context(cmd.Context(), "harp-bug", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
	defer cancel()

	// No argument check
	if len(args) > 0 {
		log.For(ctx).Fatal("bug command takes no arguments")
	}

	// Prepare the report body
	body := cmdutil.BugReport()

	// Open the browser to issue creation form
	reportURL := "https://github.com/elastic/harp/issues/new?body=" + url.QueryEscape(body)
	if err := open.Run(reportURL); err != nil {
		fmt.Print("Please file a new issue at github.com/elastic/harp/issues/new using this template:\n\n")
		fmt.Print(body)
	}
}

// -----------------------------------------------------------------------------
