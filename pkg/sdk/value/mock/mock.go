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

package mock

import (
	"context"

	"github.com/elastic/harp/pkg/sdk/value"
)

func Transformer(err error) value.Transformer {
	return &mockedTransformer{
		err: err,
	}
}

type mockedTransformer struct {
	err error
}

func (m *mockedTransformer) To(ctx context.Context, input []byte) ([]byte, error) {
	return input, m.err
}

func (m *mockedTransformer) From(ctx context.Context, input []byte) ([]byte, error) {
	return input, m.err
}
