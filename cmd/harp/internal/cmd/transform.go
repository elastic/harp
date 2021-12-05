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
	"github.com/spf13/cobra"
)

// -----------------------------------------------------------------------------

var transformCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transform",
		Short: "Transform input value using encryption transformers",
	}

	// Deprecated
	cmd.AddCommand(transformEncryptionCmd())

	// Add commands
	cmd.AddCommand(transformEncryptCmd())
	cmd.AddCommand(transformDecryptCmd())
	cmd.AddCommand(transformSignCmd())
	cmd.AddCommand(transformVerifyCmd())

	return cmd
}
