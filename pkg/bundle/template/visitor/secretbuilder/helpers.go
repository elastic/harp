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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle/secret"
	csov1 "github.com/elastic/harp/pkg/cso/v1"
	"github.com/elastic/harp/pkg/sdk/types"
	"github.com/elastic/harp/pkg/template/engine"
)

func parseSecretTemplate(templateContext engine.Context, ring csov1.Ring, secretPath string, item *bundlev1.SecretSuffix, data interface{}) (*bundlev1.Package, error) {
	// Prepare secret chain
	chain, err := buildSecretChain(templateContext, secretPath, item, data)
	if err != nil {
		return nil, fmt.Errorf("unable to build secret chain for path '%s': %w", secretPath, err)
	}

	// No error
	return buildPackage(templateContext, secretPath, chain, item)
}

func buildSecretChain(templateContext engine.Context, secretPath string, item *bundlev1.SecretSuffix, data interface{}) (*bundlev1.SecretChain, error) {
	// Check arguments
	if types.IsNil(templateContext) {
		return nil, errors.New("unable to process with nil context")
	}
	if secretPath == "" {
		return nil, errors.New("unable to process with blank secret path")
	}
	if item == nil {
		return nil, errors.New("unable to process with nil secret suffix")
	}

	// Extract generated secret value
	kv, err := renderSuffix(templateContext, secretPath, item, data)
	if err != nil {
		return nil, fmt.Errorf("unable to render secret suffix (path:%s suffix:%s): %w", secretPath, item.Suffix, err)
	}

	// Prepare secret list
	chain := &bundlev1.SecretChain{
		Version: uint32(0),
		Labels: map[string]string{
			"generated": "true",
		},
		Annotations: map[string]string{
			"creationDate": fmt.Sprintf("%d", time.Now().UTC().Unix()),
			"description":  item.Description,
			"template":     item.Template,
		},
		Data:            make([]*bundlev1.KV, 0),
		NextVersion:     nil,
		PreviousVersion: nil,
	}

	// Check vendor status
	if item.Vendor {
		chain.Labels["vendor"] = "true"
	}

	// Iterate over K/V
	for key, value := range kv {
		// Skip empty key
		if key == "" {
			continue
		}

		// Pack secret value
		secretBody, err := secret.Pack(value)
		if err != nil {
			return nil, fmt.Errorf("unable to pack secret value for path '%s': %w", secretPath, err)
		}

		// Add secret to package
		chain.Data = append(chain.Data, &bundlev1.KV{
			Key:   key,
			Type:  fmt.Sprintf("%T", value),
			Value: secretBody,
		})
	}

	// No error
	return chain, nil
}

// suffix is a function used for suffix template compiler.
func renderSuffix(templateContext engine.Context, secretPath string, item *bundlev1.SecretSuffix, data interface{}) (map[string]interface{}, error) {
	// Check input
	if types.IsNil(templateContext) {
		return nil, errors.New("unable to process with nil context")
	}
	if item == nil {
		return nil, fmt.Errorf("unable to process nil item")
	}
	if len(item.Content) == 0 && item.Template == "" {
		return nil, fmt.Errorf("content or template property must be defined")
	}

	kv := map[string]interface{}{}

	if item.Template != "" {
		payload, err := engine.RenderContextWithData(templateContext, item.Template, data)
		if err != nil {
			return nil, fmt.Errorf("unable to render suffix template: %w", err)
		}

		// Parse generated JSON
		if !json.Valid([]byte(payload)) {
			return nil, fmt.Errorf("unable to validate generated json for secret path '%s': %s", secretPath, payload)
		}

		// Extract payload as K/V
		if err := json.Unmarshal([]byte(payload), &kv); err != nil {
			return nil, fmt.Errorf("unable to assemble secret package for secret path '%s': %w", secretPath, err)
		}
	}

	if len(item.Content) > 0 {
		for filename, content := range item.Content {
			// Render filename
			renderedFilename, err := engine.RenderContextWithData(templateContext, filename, data)
			if err != nil {
				return nil, fmt.Errorf("unable to render filename template: %w", err)
			}

			// Render content
			payload, err := engine.RenderContextWithData(templateContext, content, data)
			if err != nil {
				return nil, fmt.Errorf("unable to render file content template: %w", err)
			}

			// Assign result
			kv[renderedFilename] = payload
		}
	}

	// No error
	return kv, nil
}

func buildPackage(templateContext engine.Context, secretPath string, chain *bundlev1.SecretChain, item *bundlev1.SecretSuffix) (*bundlev1.Package, error) {
	// Check arguments
	if types.IsNil(templateContext) {
		return nil, errors.New("unable to process with nil context")
	}
	if secretPath == "" {
		return nil, errors.New("unable to process with blank secret path")
	}
	if chain == nil {
		return nil, errors.New("unable to process with nil secret chain")
	}
	if item == nil {
		return nil, errors.New("unable to process with nil secret suffix")
	}

	// Evaluate annotation values
	if item.Annotations != nil {
		for k, v := range item.Annotations {
			// Evaluate using template engine
			renderedValue, err := engine.RenderContext(templateContext, v)
			if err != nil {
				return nil, fmt.Errorf("unable to render annotations value '%s' of '%s': %w", k, secretPath, err)
			}

			item.Annotations[k] = renderedValue
		}
	}

	// Evaluate labels values
	if item.Labels != nil {
		for k, v := range item.Labels {
			// Evaluate using template engine
			renderedValue, err := engine.RenderContext(templateContext, v)
			if err != nil {
				return nil, fmt.Errorf("unable to render label value '%s' of '%s': %w", k, secretPath, err)
			}

			item.Labels[k] = renderedValue
		}
	}

	// Assemble final secret package
	return &bundlev1.Package{
		Name:        secretPath,
		Secrets:     chain,
		Annotations: item.GetAnnotations(),
		Labels:      item.GetLabels(),
	}, nil
}
