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

package engine

// Context describes engine rendering context contract.
type Context interface {
	Name() string
	StrictMode() bool
	Delims() (string, string)
	SecretReaders() []SecretReaderFunc
	Values() Values
	Files() Files
}

// -----------------------------------------------------------------------------

// ContextOption defines context functional builder function
type ContextOption func(*context)

// WithName sets the template name.
func WithName(value string) ContextOption {
	return func(ctx *context) {
		ctx.name = value
	}
}

// WithStrictMode enable or disable strict rendering mode.
func WithStrictMode(value bool) ContextOption {
	return func(ctx *context) {
		ctx.strictMode = value
	}
}

// WithDelims defines used delimiters for rendering engine.
func WithDelims(left, right string) ContextOption {
	return func(ctx *context) {
		ctx.delimLeft = left
		ctx.delimRight = right
	}
}

// WithSecretReaders defines secret resolver functions used by `secret` template
// function.
func WithSecretReaders(values ...SecretReaderFunc) ContextOption {
	return func(ctx *context) {
		if len(values) > 0 {
			ctx.secretReaders = values
		}
	}
}

// WithValues defines template values injected via CLI.
func WithValues(values Values) ContextOption {
	return func(ctx *context) {
		ctx.values = values
	}
}

// WithFiles defines file collection.
func WithFiles(files Files) ContextOption {
	return func(ctx *context) {
		ctx.files = files
	}
}

// NewContext returns a template rendering context.
func NewContext(opts ...ContextOption) Context {
	defaultContext := &context{
		delimLeft:     "{{",
		delimRight:    "}}",
		name:          "root",
		secretReaders: []SecretReaderFunc{},
		strictMode:    true,
	}

	// Apply functions
	for _, opt := range opts {
		opt(defaultContext)
	}

	// Return modified context
	return defaultContext
}

// -----------------------------------------------------------------------------

// Context describes rendering context.
type context struct {
	name          string
	strictMode    bool
	delimLeft     string
	delimRight    string
	secretReaders []SecretReaderFunc
	values        Values
	files         Files
}

// Name returns template name
func (ctx *context) Name() string {
	return ctx.name
}

// StrictMode retruns strict mode status of template engine.
func (ctx *context) StrictMode() bool {
	return ctx.strictMode
}

// Delims returns left and right delimiters used to compile the template.
func (ctx *context) Delims() (left, right string) {
	return ctx.delimLeft, ctx.delimRight
}

// SecretReaders returns secret reader function called by `secret` template function.
func (ctx *context) SecretReaders() []SecretReaderFunc {
	return ctx.secretReaders
}

// Values returns binded values from rendering context.
func (ctx *context) Values() Values {
	return ctx.values
}

// Files returns binded files from rendering context.
func (ctx *context) Files() Files {
	return ctx.files
}
