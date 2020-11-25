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

package identity

import (
	"context"

	"github.com/elastic/harp/pkg/sdk/value"
)

type identityTransformer struct {
}

// Transformer returns a non-operation transformer
func Transformer() (value.Transformer, error) {
	return &identityTransformer{}, nil
}

// -----------------------------------------------------------------------------

func (t *identityTransformer) From(_ context.Context, in []byte) ([]byte, error) {
	return in, nil
}

func (t *identityTransformer) To(_ context.Context, in []byte) ([]byte, error) {
	return in, nil
}
