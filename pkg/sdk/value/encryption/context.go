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

package encryption

import "context"

type contextKey string

func (c contextKey) String() string {
	return "github.com/elastic/harp/pkg/sdk/value/encryption#" + string(c)
}

var (
	contextKeyNonce = contextKey("nonce")
	contextKeyAAD   = contextKey("aad")
)

func WithNonce(ctx context.Context, value []byte) context.Context {
	return context.WithValue(ctx, contextKeyNonce, value)
}

// Nonce gets the nonce value from the context.
func Nonce(ctx context.Context) ([]byte, bool) {
	nonce, ok := ctx.Value(contextKeyNonce).([]byte)
	return nonce, ok
}

func WithAdditionalData(ctx context.Context, value []byte) context.Context {
	return context.WithValue(ctx, contextKeyAAD, value)
}

// AdditionalData gets the aad value from the context.
func AdditionalData(ctx context.Context) ([]byte, bool) {
	aad, ok := ctx.Value(contextKeyAAD).([]byte)
	return aad, ok
}
