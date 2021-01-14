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

package v1

import (
	"fmt"
	"io"
	"text/template"

	csov1 "github.com/elastic/harp/api/gen/go/cso/v1"
)

// Text returns english text templates.
func Text() map[csov1.RingLevel]string {
	return map[csov1.RingLevel]string{
		csov1.RingLevel_RING_LEVEL_META:           `Give me a secret named "{{ .Key }}"`,
		csov1.RingLevel_RING_LEVEL_INFRASTRUCTURE: `Give me an infrastructure secret named "{{ .Key }}", for "{{ .CloudProvider }}" cloud provider, concerning account "{{ .AccountId }}", located in region "{{ .Region }}", for service "{{ .ServiceName }}".`,
		csov1.RingLevel_RING_LEVEL_PLATFORM:       `Give me a platform secret named "{{ .Key }}", for a service named "{{ .ServiceName }}", located in region "{{ .Region }}", part of a "{{ .Stage }}" platform named "{{ .Name }}".`,
		csov1.RingLevel_RING_LEVEL_PRODUCT:        `Give me a product secret named "{{ .Key }}", concerning the component "{{ .ComponentName }}", for a product named "{{ .Name }}", in version "{{ .Version }}".`,
		csov1.RingLevel_RING_LEVEL_APPLICATION:    `Give me an application secret named "{{ .Key }}", concerning the component "{{ .ComponentName }}", for a product named "{{ .ProductName }}", in version "{{ .ProductVersion }}", running on a "{{ .Stage }}" platform named "{{.PlatformName}}".`,
		csov1.RingLevel_RING_LEVEL_ARTIFACT:       `Give me an artifact secret named "{{ .Key }}", concerning the "{{ .Type }}" artifact with ID "{{ .Id }}".`,
	}
}

// Interpret returns the cso secret interpretation.
func Interpret(secret *csov1.Secret, templates map[csov1.RingLevel]string, w io.Writer) error {
	// Check arguments
	if secret == nil {
		return fmt.Errorf("unable to interpret nil secret")
	}

	// Retrieve template according to ring level
	tBody, ok := templates[secret.RingLevel]
	if !ok {
		return fmt.Errorf("unable to retrieve temlate")
	}

	// Compile template
	t, err := template.New("interpret").Parse(tBody)
	if err != nil {
		return fmt.Errorf("unable to compile template: %w", err)
	}

	// Merge data
	switch secret.RingLevel {
	case csov1.RingLevel_RING_LEVEL_META:
		err = t.Execute(w, secret.GetMeta())
	case csov1.RingLevel_RING_LEVEL_INFRASTRUCTURE:
		err = t.Execute(w, secret.GetInfrastructure())
	case csov1.RingLevel_RING_LEVEL_PLATFORM:
		err = t.Execute(w, secret.GetPlatform())
	case csov1.RingLevel_RING_LEVEL_PRODUCT:
		err = t.Execute(w, secret.GetProduct())
	case csov1.RingLevel_RING_LEVEL_APPLICATION:
		err = t.Execute(w, secret.GetApplication())
	case csov1.RingLevel_RING_LEVEL_ARTIFACT:
		err = t.Execute(w, secret.GetArtifact())
	case csov1.RingLevel_RING_LEVEL_INVALID, csov1.RingLevel_RING_LEVEL_UNKNOWN:
		err = fmt.Errorf("invalid secret ring %v", secret.RingLevel)
	default:
		err = fmt.Errorf("invalid secret ring %v", secret.RingLevel)
	}

	// Return error
	if err != nil {
		return fmt.Errorf("unable to interpret CSO path: %w", err)
	}

	// No error
	return nil
}
