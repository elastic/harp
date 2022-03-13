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

package template

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/harp/pkg/tasks"
	tplcmdutil "github.com/elastic/harp/pkg/template/cmdutil"
)

// ValueTask implements value object generation task.
type ValueTask struct {
	OutputWriter tasks.WriterProvider
	ValueFiles   []string
	Values       []string
	StringValues []string
	FileValues   []string
}

// Run the task.
func (t *ValueTask) Run(ctx context.Context) error {
	// Load values
	valueOpts := tplcmdutil.ValueOptions{
		ValueFiles:   t.ValueFiles,
		Values:       t.Values,
		StringValues: t.StringValues,
		FileValues:   t.FileValues,
	}
	values, err := valueOpts.MergeValues()
	if err != nil {
		return fmt.Errorf("unable to process values: %w", err)
	}

	// Create output writer
	writer, err := t.OutputWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to create output writer: %w", err)
	}

	// Write rendered content
	if err := json.NewEncoder(writer).Encode(values); err != nil {
		return fmt.Errorf("unable to dump values as JSON: %w", err)
	}

	return nil
}
