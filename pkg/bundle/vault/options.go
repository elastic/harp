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

package vault

import (
	"fmt"
	"regexp"
)

type options struct {
	prefix       string
	withMetadata bool
	exclusions   []*regexp.Regexp
	includes     []*regexp.Regexp
}

// Option defines the functional pattern for bundle operation settings.
type Option func(*options) error

// WithExcludePath register a path exclusion regexp
func WithExcludePath(value string) Option {
	return func(opts *options) error {
		// Compile RegExp first
		r, err := regexp.Compile(value)
		if err != nil {
			return fmt.Errorf("unable to compile `%s` as a valid regexp: %w", value, err)
		}

		// Append to exclusions
		opts.exclusions = append(opts.exclusions, r)

		// No error
		return nil
	}
}

// WithIncludePath register a path inclusion regexp
func WithIncludePath(value string) Option {
	return func(opts *options) error {
		// Compile RegExp first
		r, err := regexp.Compile(value)
		if err != nil {
			return fmt.Errorf("unable to compile `%s` as a valid regexp: %w", value, err)
		}

		// Append to exclusions
		opts.includes = append(opts.includes, r)

		// No error
		return nil
	}
}

// WithPrefix add a prefix to path value
func WithPrefix(value string) Option {
	return func(opts *options) error {
		opts.prefix = value
		// No error
		return nil
	}
}

// WithMetadata add package metadata as secret value to be exported in Vault.
func WithMetadata(value bool) Option {
	return func(opts *options) error {
		opts.withMetadata = value
		// No error
		return nil
	}
}
