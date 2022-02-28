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
	"encoding/json"
	"fmt"

	"github.com/elastic/harp/pkg/container/oci/schema"
)

func NewConfig() schema.Config {
	return &Config{
		V: schema.V1,
	}
}

// -----------------------------------------------------------------------------

type Config struct {
	V              schema.Version `json:"co.elastic.harp.oci.version"`
	ContainerFiles []string       `json:"co.elastic.harp.oci.containers"`
	TemplateFiles  []string       `json:"co.elastic.harp.oci.templates"`
}

// Containers returns the current image container filenames.
func (c *Config) Containers() []string {
	return c.ContainerFiles
}

func (c *Config) SetContainers(containers []string) {
	c.ContainerFiles = containers
}

// Containers returns the current image template archive filenames.
func (c *Config) Templates() []string {
	return c.TemplateFiles
}

func (c *Config) SetTemplates(templates []string) {
	c.TemplateFiles = templates
}

// -----------------------------------------------------------------------------

// ParseConfig parses the given input as a v1.Config.
func ParseConfig(data []byte) (schema.Config, error) {
	var c Config
	err := json.Unmarshal(data, &c)
	if err != nil {
		return nil, fmt.Errorf("error parsing OCI image config: %w", err)
	}
	if c.V != schema.V1 {
		return nil, fmt.Errorf("invalid config version")
	}

	// No error
	return &c, nil
}
