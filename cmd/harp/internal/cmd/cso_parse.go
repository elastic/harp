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
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"

	csov1 "github.com/elastic/harp/pkg/cso/v1"
	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
)

var (
	csoParsePath   string
	csoParseAsText bool
)

// -----------------------------------------------------------------------------

var csoParseCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "parse",
		Aliases: []string{"p"},
		Short:   "Parse given CSO path",
		Run:     runCSOParse,
	}

	// Parameters
	cmd.Flags().StringVar(&csoParsePath, "path", "", "Path to parse")
	cmd.Flags().BoolVar(&csoParseAsText, "text", false, "Display path component as text")

	return cmd
}

func runCSOParse(cmd *cobra.Command, args []string) {
	ctx, cancel := cmdutil.Context(cmd.Context(), "harp-cso-parse", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
	defer cancel()

	// Validate and pack secret path first
	s, err := csov1.Pack(csoParsePath)
	if err != nil {
		log.For(ctx).Fatal("unable to validate given path as a compliant CSO path", zap.Error(err), zap.String("path", csoParsePath))
	}

	if csoParseAsText {
		if err := csov1.Interpret(s, csov1.Text(), os.Stdout); err != nil {
			log.For(ctx).Fatal("unable to generate textual interpretation of given path", zap.Error(err), zap.String("path", csoParsePath))
		}
	} else {
		// Override values as nil
		s.Value = nil

		// Marshal using protojson
		out, err := protojson.Marshal(s)
		if err != nil {
			log.For(ctx).Fatal("unable to generate json interpretation of given path", zap.Error(err), zap.String("path", csoParsePath))
		}

		// Dump in stdout
		fmt.Fprintf(os.Stdout, "%s", string(out))
	}
}
