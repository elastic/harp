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

package hcl

// Config is the configuration structure for bundle DSL.
type Config struct {
	Annotations map[string]string `hcl:"annotations,optional"`
	Labels      map[string]string `hcl:"labels,optional"`
	Packages    []Package         `hcl:"package,block"`
}

type Package struct {
	Path        string            `hcl:"path,label"`
	Description string            `hcl:"description"`
	Annotations map[string]string `hcl:"annotations,optional"`
	Labels      map[string]string `hcl:"labels,optional"`
	Secrets     map[string]string `hcl:"secrets"`
}
