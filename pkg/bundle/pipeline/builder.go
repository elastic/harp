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

package pipeline

import "io"

// Options defines default options.
type Options struct {
	input         io.Reader
	output        io.Writer
	disableOutput bool
	fpf           FileProcessorFunc
	ppf           PackageProcessorFunc
	cpf           ChainProcessorFunc
	kpf           KVProcessorFunc
}

// Option represents option function
type Option func(*Options)

// InputReader defines the input reader used to retrieve the bundle content.
func InputReader(value io.Reader) Option {
	return func(opts *Options) {
		opts.input = value
	}
}

// OutputWriter defines where the bundle will be written after process execution.
func OutputWriter(value io.Writer) Option {
	return func(opts *Options) {
		opts.output = value
	}
}

// OutputDisabled assign the value to disableOutput option.
func OutputDisabled() Option {
	return func(opts *Options) {
		opts.disableOutput = true
	}
}

// FileProcessor assign the file object processor.
func FileProcessor(f FileProcessorFunc) Option {
	return func(opts *Options) {
		opts.fpf = f
	}
}

// PackageProcessor assign the package object processor.
func PackageProcessor(f PackageProcessorFunc) Option {
	return func(opts *Options) {
		opts.ppf = f
	}
}

// ChainProcessor assign the chain object processor.
func ChainProcessor(f ChainProcessorFunc) Option {
	return func(opts *Options) {
		opts.cpf = f
	}
}

// KVProcessor assign the KV object processor.
func KVProcessor(f KVProcessorFunc) Option {
	return func(opts *Options) {
		opts.kpf = f
	}
}
