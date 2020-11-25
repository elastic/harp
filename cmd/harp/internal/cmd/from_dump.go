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
	"io/ioutil"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle"
	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
)

var (
	fromDumpOutputPath string
	fromDumpInputPath  string
)

// -----------------------------------------------------------------------------

var fromDumpCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dump",
		Short: "Import from bundle dump output as a secret container",
		Run:   runfromDump,
	}

	// Parameters
	cmd.Flags().StringVar(&fromDumpInputPath, "in", "", "JSON input file ('-' for stdin or filename)")
	cmd.Flags().StringVar(&fromDumpOutputPath, "out", "", "Container output ('-' for stdout or filename)")

	return cmd
}

func runfromDump(cmd *cobra.Command, args []string) {
	ctx, cancel := cmdutil.Context(cmd.Context(), "harp-from-dump", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
	defer cancel()

	// Create input reader
	reader, err := cmdutil.Reader(fromDumpInputPath)
	if err != nil {
		log.For(ctx).Fatal("unable to open input file", zap.Error(err), zap.String("path", fromDumpInputPath))
	}

	// Create output writer
	writer, err := cmdutil.Writer(fromDumpOutputPath)
	if err != nil {
		log.For(ctx).Fatal("unable to open output bundle", zap.Error(err), zap.String("path", fromDumpOutputPath))
	}

	// Drain input content
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		log.For(ctx).Fatal("unable to read input content", zap.Error(err))
	}

	// unmarshall from JSON
	var secrets bundlev1.Bundle
	if err = protojson.Unmarshal(content, &secrets); err != nil {
		log.For(ctx).Fatal("unable to decode JSON bundle", zap.Error(err))
	}

	// Dump all content
	if err := bundle.Dump(writer, &secrets); err != nil {
		log.For(ctx).Fatal("unable to dump bundle content", zap.Error(err))
	}
}
