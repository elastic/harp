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

package signature

import (
	"context"
)

var (
	contextKeyDetachedSignature      = contextKey("detachedSignature")
	contextKeyInputHash              = contextKey("inputHash")
	contextKeyDeterministicSignature = contextKey("deterministic")
)

// -----------------------------------------------------------------------------

type contextKey string

func (c contextKey) String() string {
	return "github.com/elastic/harp/pkg/sdk/value/signature/" + string(c)
}

// -----------------------------------------------------------------------------

func WithDetachedSignature(ctx context.Context, value bool) context.Context {
	return context.WithValue(ctx, contextKeyDetachedSignature, value)
}

func IsDetached(ctx context.Context) bool {
	value, ok := ctx.Value(contextKeyDetachedSignature).(bool)
	if !ok {
		return false
	}
	return value
}

func WithInputPreHashed(ctx context.Context, value bool) context.Context {
	return context.WithValue(ctx, contextKeyInputHash, value)
}

func IsInputPreHashed(ctx context.Context) bool {
	value, ok := ctx.Value(contextKeyInputHash).(bool)
	if !ok {
		return false
	}
	return value
}

func WithDetermisticSignature(ctx context.Context, value bool) context.Context {
	return context.WithValue(ctx, contextKeyDeterministicSignature, value)
}

func IsDeterministic(ctx context.Context) bool {
	value, ok := ctx.Value(contextKeyDeterministicSignature).(bool)
	if !ok {
		return false
	}
	return value
}
