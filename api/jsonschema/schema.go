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

package jsonschema

import (
	_ "embed"
)

//go:embed harp.bundle.v1/Bundle.json
var bundleV1BundleSchemaDefinition []byte

// BundleV1BundleSchema returns the `harp.bundle.v1.Bundle` jsonschema content.
func BundleV1BundleSchema() []byte {
	return bundleV1BundleSchemaDefinition
}

//go:embed harp.bundle.v1/Patch.json
var bundleV1PatchSchemaDefinition []byte

// BundleV1PatchSchema returns the `harp.bundle.v1.Patch` jsonschema content.
func BundleV1PatchSchema() []byte {
	return bundleV1PatchSchemaDefinition
}

//go:embed harp.bundle.v1/RuleSet.json
var bundleV1RuleSetSchemaDefinition []byte

// BundleV1RuleSetSchema returns the `harp.bundle.v1.RuleSet` jsonschema content.
func BundleV1RuleSetSchema() []byte {
	return bundleV1RuleSetSchemaDefinition
}

//go:embed harp.bundle.v1/Template.json
var bundleV1TemplateSchemaDefinition []byte

// BundleV1TemplateSchema returns the `harp.bundle.v1.Template` jsonschema content.
func BundleV1TemplateSchema() []byte {
	return bundleV1TemplateSchemaDefinition
}
