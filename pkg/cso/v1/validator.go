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
	"strings"

	semver "github.com/blang/semver/v4"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"

	"github.com/elastic/harp/pkg/sdk/types"
)

const (
	// Local is used to describe local infrastructure provider.
	reservedLocalProvider = "local"
	// Global is used to replace region component to indicate non-region bound secret suffix.
	reservedGlobalRegion = "global"
)

var validators = map[string]func([]string) error{
	"meta":     validateMeta,
	"infra":    validateInfra,
	"platform": validatePlatform,
	"product":  validateProduct,
	"app":      validateApplication,
	"artifact": validateArtifact,
}

// Validate path according to to CSO model
func Validate(path string) error {
	// Validate path
	if err := validation.Validate(path,
		validation.Required,
		is.PrintableASCII,
	); err != nil {
		return fmt.Errorf("unable to secret path: %w", err)
	}

	// Clean path first
	cleanPath := Clean(path)

	// Split path using '/'
	parts := strings.Split(cleanPath, "/")

	// Check path part count
	if len(parts) < 2 {
		return fmt.Errorf("invalid secret path, should contains more than 2 parts")
	}

	// Check validator according to given ring value
	v, ok := validators[parts[0]]
	if !ok {
		return fmt.Errorf("invalid ring value (%s)", parts[0])
	}

	// Delegate to ring validator
	return v(parts[1:])
}

// -----------------------------------------------------------------------------

func validateMeta(parts []string) error {
	// Validate parts count
	if len(parts) < 2 {
		return fmt.Errorf("invalid part count for meta secret path")
	}

	// Validate accounts
	if err := validation.Validate(parts[0],
		validation.Required,
		is.PrintableASCII,
	); err != nil {
		return fmt.Errorf("unable to validate meta path (%s): %w", parts[0], err)
	}

	// Meta has no constraints
	return nil
}

// -----------------------------------------------------------------------------

var cloudProviderRegions = map[string]types.StringArray{
	reservedLocalProvider: {},
	"aws": {
		"af-south-1",
		"ap-east-1",
		"ap-northeast-1",
		"ap-northeast-2",
		"ap-northeast-3",
		"ap-south-1",
		"ap-southeast-1",
		"ap-southeast-2",
		"ca-central-1",
		"eu-central-1",
		"eu-north-1",
		"eu-south-1",
		"eu-west-1",
		"eu-west-2",
		"eu-west-3",
		"me-south-1",
		"me-south-1",
		"sa-east-1",
		"us-east-1",
		"us-east-2",
		"us-west-1",
		"us-west-2",
	},
	"aws-us-gov": {
		"us-gov-east-1",
		"us-gov-west-1",
	},
	"gcp": {
		"asia-east1",
		"asia-east2",
		"asia-northeast1",
		"asia-northeast2",
		"asia-south1",
		"asia-southeast1",
		"australia-southeast1",
		"europe-north1",
		"europe-west1",
		"europe-west2",
		"europe-west3",
		"europe-west4",
		"europe-west6",
		"northamerica-northeast1",
		"southamerica-east1",
		"us-central1",
		"us-east1",
		"us-east4",
		"us-west1",
		"us-west2",
	},
	"azure": {
		"asia",
		"asiapacific",
		"australia",
		"australiacentral",
		"australiacentral2",
		"australiaeast",
		"australiasoutheast",
		"brazil",
		"brazilsouth",
		"brazilsoutheast",
		"canada",
		"canadacentral",
		"canadaeast",
		"centralindia",
		"centralus",
		"centraluseuap",
		"centralusstage",
		"eastasia",
		"eastasiastage",
		"eastus",
		"eastus2",
		"eastus2euap",
		"eastus2stage",
		"eastusstage",
		"europe",
		"france",
		"francecentral",
		"francesouth",
		"germany",
		"germanynorth",
		"germanywestcentral",
		"global",
		"india",
		"japan",
		"japaneast",
		"japanwest",
		"jioindiacentral",
		"jioindiawest",
		"korea",
		"koreacentral",
		"koreasouth",
		"northcentralus",
		"northcentralusstage",
		"northeurope",
		"norway",
		"norwayeast",
		"norwaywest",
		"southafrica",
		"southafricanorth",
		"southafricawest",
		"southcentralus",
		"southcentralusstage",
		"southeastasia",
		"southeastasiastage",
		"southindia",
		"swedencentral",
		"switzerland",
		"switzerlandnorth",
		"switzerlandwest",
		"uae",
		"uaecentral",
		"uaenorth",
		"uk",
		"uksouth",
		"ukwest",
		"unitedstates",
		"westcentralus",
		"westeurope",
		"westindia",
		"westus",
		"westus2",
		"westus2stage",
		"westus3",
		"westusstage",
	},
	"azure-us-gov": {
		"usgovarizona",
		"usgovvirginia",
		"usdodcentral",
		"usdodeast",
		"usgoviowa",
		"usgovtexas",
	},
	"ibm": {
		"au-syd",
		"in-che",
		"jp-osa",
		"jp-tok",
		"kr-seo",
		"eu-de",
		"eu-gb",
		"ca-tor",
		"us-south",
		"us-east",
		"br-sao",
	},
}

func validateInfra(parts []string) error {
	// Validate parts count
	if len(parts) < 4 {
		return fmt.Errorf("invalid part count for infrastructure secret path")
	}

	// Validate cloud provider
	r, ok := cloudProviderRegions[parts[0]]
	if !ok {
		return fmt.Errorf("cloud provider (%s) not supported", r)
	}

	// Validate accounts
	if err := validation.Validate(parts[1],
		validation.Required,
		is.PrintableASCII,
	); err != nil {
		return fmt.Errorf("unable to validate infrastructure cloud provider account (%s): %w", parts[1], err)
	}

	// Validate region if not local provider and not global region
	if parts[0] != reservedLocalProvider && parts[2] != reservedGlobalRegion && !r.Contains(parts[2]) {
		return fmt.Errorf("invalid region (%s) for account (%s) on cloud provider (%s)", parts[2], parts[1], parts[0])
	}

	// Infra has no more constraints
	return nil
}

// -----------------------------------------------------------------------------

var platformQualityLevels = types.StringArray{"production", "staging", "qa", "dev"}

func validatePlatform(parts []string) error {
	// Validate parts count
	if len(parts) < 5 {
		return fmt.Errorf("invalid part count for platform secret path")
	}

	// Validate quality grade level
	if !platformQualityLevels.Contains(parts[0]) {
		return fmt.Errorf("platform quality level (%s) is not supported", parts[0])
	}

	// Validate name
	if err := validation.Validate(parts[1],
		validation.Required,
		is.PrintableASCII,
	); err != nil {
		return fmt.Errorf("unable to validate platform name (%s): %w", parts[1], err)
	}

	// Validate platform region
	r := parts[2]
	if r != reservedGlobalRegion {
		regionFound := false
		for _, regions := range cloudProviderRegions {
			if regions.Contains(r) {
				regionFound = true
				break
			}
		}
		if !regionFound {
			return fmt.Errorf("unable to find a region matching (%s)", r)
		}
	}

	// Validate accounts
	if err := validation.Validate(parts[3],
		validation.Required,
		is.PrintableASCII,
	); err != nil {
		return fmt.Errorf("unable to validate platform service (%s): %w", parts[1], err)
	}

	// Platform has no more constraints
	return nil
}

// -----------------------------------------------------------------------------

func validateProduct(parts []string) error {
	// Validate parts count
	if len(parts) < 3 {
		return fmt.Errorf("invalid part count for product secret path")
	}

	// Extract product name
	if err := validation.Validate(parts[0],
		validation.Required,
		is.PrintableASCII,
	); err != nil {
		return fmt.Errorf("unable to validate product name (%s): %w", parts[0], err)
	}

	// check version as a semver compliant version
	if err := validateSemVer(parts[1]); err != nil {
		return fmt.Errorf("invalid product (%s) version (%s), semver not compliant: %w", parts[0], parts[1], err)
	}

	// Product has no more constraints
	return nil
}

// -----------------------------------------------------------------------------

func validateApplication(parts []string) error {
	// Validate parts count
	if len(parts) < 6 {
		return fmt.Errorf("invalid part count for application secret path")
	}

	// Validate quality grade level
	if !platformQualityLevels.Contains(parts[0]) {
		return fmt.Errorf("application quality level (%s) is not supported", parts[0])
	}

	// Validate platform name
	if err := validation.Validate(parts[1],
		validation.Required,
		is.PrintableASCII,
	); err != nil {
		return fmt.Errorf("unable to validate platform name (%s): %w", parts[1], err)
	}

	// Extract product name
	if err := validation.Validate(parts[2],
		validation.Required,
		is.PrintableASCII,
	); err != nil {
		return fmt.Errorf("unable to validate product name (%s): %w", parts[2], err)
	}

	// check version as a semver compliant version
	if err := validateSemVer(parts[3]); err != nil {
		return fmt.Errorf("invalid product (%s) version (%s), semver not compliant: %w", parts[2], parts[3], err)
	}

	// Extract component
	if err := validation.Validate(parts[4],
		validation.Required,
		is.PrintableASCII,
	); err != nil {
		return fmt.Errorf("invalid component (%s) for product (%s) version (%s), %w", parts[4], parts[3], parts[2], err)
	}

	// Product has no more constraints
	return nil
}

// -----------------------------------------------------------------------------

func validateArtifact(parts []string) error {
	// Validate parts count
	if len(parts) < 2 {
		return fmt.Errorf("invalid part count for artifact secret path")
	}

	// Validate type
	if err := validation.Validate(parts[1],
		validation.Required,
		is.PrintableASCII,
	); err != nil {
		return fmt.Errorf("unable to validate artifact type (%s): %w", parts[1], err)
	}

	// Artifact has no more constraints
	return nil
}

// -----------------------------------------------------------------------------

func validateSemVer(version string) error {
	// Clean input
	version = strings.TrimPrefix(strings.TrimSpace(strings.ToLower(version)), "v")

	// check version as a semver compliant version
	_, err := semver.Make(version)
	if err != nil {
		return fmt.Errorf("version '%s' has not a valid semver syntax: %w", version, err)
	}

	// No error
	return nil
}
