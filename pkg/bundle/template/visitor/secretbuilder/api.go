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

package secretbuilder

import (
	"errors"
	"fmt"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle/template/visitor"
	"github.com/elastic/harp/pkg/template/engine"
)

// New returns a secret builder visitor instance.
func New(result *bundlev1.Bundle, templateCtx engine.Context) visitor.TemplateVisitor {
	return &secretBuilder{
		bundle:          result,
		templateContext: templateCtx,
	}
}

// -----------------------------------------------------------------------------

type secretBuilder struct {
	bundle          *bundlev1.Bundle
	templateContext engine.Context
	err             error
}

//nolint:gocyclo,gocognit,funlen // refactoring later
func (sb *secretBuilder) Visit(t *bundlev1.Template) {
	results := make(chan *bundlev1.Package)

	// Check arguments
	if t == nil {
		sb.err = errors.New("template is nil")
		return
	}
	if t.Spec == nil {
		sb.err = errors.New("template spec nil")
		return
	}
	if t.Spec.Namespaces == nil {
		sb.err = errors.New("template spec namespace nil")
		return
	}

	go func() {
		defer close(results)

		// Infrastructure secrets
		if t.Spec.Namespaces.Infrastructure != nil {
			for _, obj := range t.Spec.Namespaces.Infrastructure {
				// Initialize a infrastructure visitor
				v := infrastructure(results, sb.templateContext)

				// Traverse the object-tree
				visitor.InfrastructureDecorator(obj).Accept(v)

				// Get result
				if err := v.Error(); err != nil {
					sb.err = err
					return
				}
			}
		}

		// Platform secrets
		if t.Spec.Namespaces.Platform != nil {
			for _, obj := range t.Spec.Namespaces.Platform {
				// Check selector
				if t.Spec.Selector == nil {
					sb.err = fmt.Errorf("selector is mandatory for platform secrets")
					return
				}

				// Initialize a infrastructure visitor
				v, err := platform(results, sb.templateContext, t.Spec.Selector.Quality, t.Spec.Selector.Platform)
				if err != nil {
					sb.err = err
					return
				}

				// Traverse the object-tree
				visitor.PlatformDecorator(obj).Accept(v)

				// Get result
				if err := v.Error(); err != nil {
					sb.err = err
					return
				}
			}
		}

		// Product secrets
		if t.Spec.Namespaces.Product != nil {
			for _, obj := range t.Spec.Namespaces.Product {
				// Check selector
				if t.Spec.Selector == nil {
					sb.err = fmt.Errorf("selector is mandatory for product secrets")
					return
				}

				// Initialize a infrastructure visitor
				v, err := product(results, sb.templateContext, t.Spec.Selector.Product, t.Spec.Selector.Version)
				if err != nil {
					sb.err = err
					return
				}

				// Traverse the object-tree
				visitor.ProductDecorator(obj).Accept(v)

				// Get result
				if err := v.Error(); err != nil {
					sb.err = err
					return
				}
			}
		}

		// Application secrets
		if t.Spec.Namespaces.Application != nil {
			for _, obj := range t.Spec.Namespaces.Application {
				// Check selector
				if t.Spec.Selector == nil {
					sb.err = fmt.Errorf("selector is mandatory for application secrets")
					return
				}

				// Initialize a infrastructure visitor
				v, err := application(results, sb.templateContext, t.Spec.Selector.Quality, t.Spec.Selector.Platform, t.Spec.Selector.Product, t.Spec.Selector.Version)
				if err != nil {
					sb.err = err
					return
				}

				// Traverse the object-tree
				visitor.ApplicationDecorator(obj).Accept(v)

				// Get result
				if err := v.Error(); err != nil {
					sb.err = err
					return
				}
			}
		}
	}()

	// Pull all packages
	for p := range results {
		sb.bundle.Packages = append(sb.bundle.Packages, p)
	}
}

func (sb *secretBuilder) Error() error {
	return sb.err
}
