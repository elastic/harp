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

package value

import (
	"context"

	"github.com/elastic/harp/pkg/sdk/value"
	"github.com/elastic/harp/pkg/server/storage"
)

// Transformer returns a value transformer decorator.
func Transformer(transformer value.Transformer, revert bool) func(storage.Engine) storage.Engine {
	// Return decorator constructor
	return func(engine storage.Engine) storage.Engine {
		return &transformerDecorator{
			next:        engine,
			transformer: transformer,
			revert:      revert,
		}
	}
}

// -----------------------------------------------------------------------------

type transformerDecorator struct {
	next        storage.Engine
	transformer value.Transformer
	revert      bool
}

func (d *transformerDecorator) Get(ctx context.Context, id string) ([]byte, error) {
	// Delegate to original storage engine
	secret, err := d.next.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if d.revert {
		// Apply reverse fonction
		return d.transformer.From(ctx, secret)
	}

	// Delegate to transformer
	return d.transformer.To(ctx, secret)
}
