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
	"fmt"
	"strings"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle/template/visitor"
	csov1 "github.com/elastic/harp/pkg/cso/v1"
	"github.com/elastic/harp/pkg/template/engine"
)

type applicationSecretBuilder struct {
	results         chan *bundlev1.Package
	templateContext engine.Context

	// Context
	quality   string
	platform  string
	product   string
	version   string
	component string
	err       error
}

// -----------------------------------------------------------------------------

// Infrastructure returns a visitor instance to generate secretpath
// and values.
func application(results chan *bundlev1.Package, templateContext engine.Context, quality, platform, product, version string) (visitor.ApplicationVisitor, error) {
	// Parse selector values
	platformQuality, err := engine.RenderContext(templateContext, quality)
	if err != nil {
		return nil, fmt.Errorf("unable to render platform.quality: %w", err)
	}
	if strings.TrimSpace(platformQuality) == "" {
		return nil, fmt.Errorf("quality selector must not be empty")
	}

	platformName, err := engine.RenderContext(templateContext, platform)
	if err != nil {
		return nil, fmt.Errorf("unable to render platform.name: %w", err)
	}
	if strings.TrimSpace(platformName) == "" {
		return nil, fmt.Errorf("platform selector must not be empty")
	}

	productName, err := engine.RenderContext(templateContext, product)
	if err != nil {
		return nil, fmt.Errorf("unable to render product.name: %w", err)
	}
	if strings.TrimSpace(productName) == "" {
		return nil, fmt.Errorf("product selector must not be empty")
	}

	productVersion, err := engine.RenderContext(templateContext, version)
	if err != nil {
		return nil, fmt.Errorf("unable to render product.version: %w", err)
	}
	if strings.TrimSpace(productVersion) == "" {
		return nil, fmt.Errorf("version selector must not be empty")
	}

	return &applicationSecretBuilder{
		results:         results,
		templateContext: templateContext,
		quality:         platformQuality,
		platform:        platformName,
		product:         productName,
		version:         productVersion,
	}, nil
}

// -----------------------------------------------------------------------------

func (b *applicationSecretBuilder) Error() error {
	return b.err
}

func (b *applicationSecretBuilder) VisitForComponent(obj *bundlev1.ApplicationComponentNS) {
	// Check arguments
	if obj == nil {
		return
	}

	// Set context values
	b.component, b.err = engine.RenderContext(b.templateContext, obj.Name)
	if b.err != nil {
		return
	}

	for _, item := range obj.Secrets {
		// Check arguments
		if item == nil {
			continue
		}

		// Parse suffix with template engine
		suffix, err := engine.RenderContext(b.templateContext, item.Suffix)
		if err != nil {
			b.err = fmt.Errorf("unable to merge template is suffix '%s'", item.Suffix)
			return
		}

		// Generate secret suffix
		secretPath, err := csov1.RingApplication.Path(b.quality, b.platform, b.product, b.version, b.component, suffix)
		if err != nil {
			b.err = err
			return
		}

		// Prepare template model
		tmplModel := &struct {
			Quality   string
			Platform  string
			Product   string
			Version   string
			Component string
			Secret    *bundlev1.SecretSuffix
		}{
			Quality:   b.quality,
			Platform:  b.platform,
			Product:   b.product,
			Version:   b.version,
			Component: b.component,
			Secret:    item,
		}

		// Compile template
		p, err := parseSecretTemplate(b.templateContext, secretPath, item, tmplModel)
		if err != nil {
			b.err = err
			return
		}

		// Add package to collection
		b.results <- p
	}
}
