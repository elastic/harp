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

type platformSecretBuilder struct {
	results         chan *bundlev1.Package
	templateContext engine.Context

	// Context
	quality   string
	name      string
	region    string
	component string
	err       error
}

// -----------------------------------------------------------------------------

// Infrastructure returns a visitor instance to generate secretpath
// and values.
func platform(results chan *bundlev1.Package, templateContext engine.Context, quality, name string) (visitor.PlatformVisitor, error) {
	// Parse selector values
	platformQuality, err := engine.RenderContext(templateContext, quality)
	if err != nil {
		return nil, fmt.Errorf("unable to render platform.quality: %w", err)
	}
	if strings.TrimSpace(platformQuality) == "" {
		return nil, fmt.Errorf("quality selector must not be empty")
	}

	platformName, err := engine.RenderContext(templateContext, name)
	if err != nil {
		return nil, fmt.Errorf("unable to render platform.name: %w", err)
	}
	if strings.TrimSpace(platformName) == "" {
		return nil, fmt.Errorf("platform selector must not be empty")
	}

	return &platformSecretBuilder{
		results:         results,
		templateContext: templateContext,
		quality:         platformQuality,
		name:            platformName,
	}, nil
}

// -----------------------------------------------------------------------------

func (b *platformSecretBuilder) Error() error {
	return b.err
}

func (b *platformSecretBuilder) VisitForRegion(obj *bundlev1.PlatformRegionNS) {
	// Check arguments
	if obj == nil {
		return
	}

	// Set context values
	b.region, b.err = engine.RenderContext(b.templateContext, obj.Region)
	if b.err != nil {
		return
	}

	// Iterates over all components
	for _, item := range obj.Components {
		b.VisitForComponent(item)
	}
}

func (b *platformSecretBuilder) VisitForComponent(obj *bundlev1.PlatformComponentNS) {
	// Check arguments
	if obj == nil {
		return
	}

	// Set context value
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
		secretPath, err := csov1.RingPlatform.Path(b.quality, b.name, b.region, b.component, suffix)
		if err != nil {
			b.err = err
			return
		}

		// Prepare template model
		tmplModel := &struct {
			Quality   string
			Name      string
			Region    string
			Component string
			Secret    *bundlev1.SecretSuffix
		}{
			Quality:   b.quality,
			Name:      b.name,
			Region:    b.region,
			Component: b.component,
			Secret:    item,
		}

		// Compile template
		p, err := parseSecretTemplate(b.templateContext, csov1.RingPlatform, secretPath, item, tmplModel)
		if err != nil {
			b.err = err
			return
		}

		// Add package to collection
		b.results <- p
	}
}
