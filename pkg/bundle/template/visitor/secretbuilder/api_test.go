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
	"testing"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/template/engine"
	fuzz "github.com/google/gofuzz"
)

func TestVisit_Fuzz(t *testing.T) {
	// Making sure the descrption never panics
	for i := 0; i < 50; i++ {
		f := fuzz.New()

		v := New(&bundlev1.Bundle{}, engine.NewContext())

		tmpl := &bundlev1.Template{
			ApiVersion: "harp.elastic.co/v1",
			Kind:       "BundleTemplate",
			Spec: &bundlev1.TemplateSpec{
				Selector: &bundlev1.Selector{
					Quality:     "production",
					Product:     "harp",
					Application: "harp",
					Version:     "v1.0.0",
					Platform:    "test",
					Component:   "cli",
				},
				Namespaces: &bundlev1.Namespaces{},
			},
		}

		// Infrastructure
		f.Fuzz(&tmpl.Spec.Namespaces.Infrastructure)
		// Platform
		f.Fuzz(&tmpl.Spec.Namespaces.Platform)
		// Product
		f.Fuzz(&tmpl.Spec.Namespaces.Product)
		// Application
		f.Fuzz(&tmpl.Spec.Namespaces.Application)

		v.Visit(tmpl)
	}
}

func TestVisit_Template_Fuzz(t *testing.T) {
	// Making sure the descrption never panics
	for i := 0; i < 50; i++ {
		f := fuzz.New()

		v := New(&bundlev1.Bundle{}, engine.NewContext())

		tmpl := &bundlev1.Template{
			ApiVersion: "harp.elastic.co/v1",
			Kind:       "BundleTemplate",
			Spec: &bundlev1.TemplateSpec{
				Selector: &bundlev1.Selector{
					Quality:     "production",
					Product:     "harp",
					Application: "harp",
					Version:     "v1.0.0",
					Platform:    "test",
					Component:   "cli",
				},
				Namespaces: &bundlev1.Namespaces{
					Infrastructure: []*bundlev1.InfrastructureNS{
						{
							Provider: "aws",
							Account:  "test",
							Regions: []*bundlev1.InfrastructureRegionNS{
								{
									Name: "eu-central-1",
									Services: []*bundlev1.InfrastructureServiceNS{
										{
											Name: "ssh",
											Type: "ec2",
											Secrets: []*bundlev1.SecretSuffix{
												{
													Suffix: "test",
												},
											},
										},
									},
								},
							},
						},
					},
					Platform: []*bundlev1.PlatformRegionNS{
						{
							Region: "eu-central-1",
							Components: []*bundlev1.PlatformComponentNS{
								{
									Type: "rds",
									Name: "postgres-1",
									Secrets: []*bundlev1.SecretSuffix{
										{
											Suffix: "test",
										},
									},
								},
							},
						},
					},
					Product: []*bundlev1.ProductComponentNS{
						{
							Type: "service",
							Name: "rest-api",
							Secrets: []*bundlev1.SecretSuffix{
								{
									Suffix: "test",
								},
							},
						},
					},
					Application: []*bundlev1.ApplicationComponentNS{
						{
							Type: "service",
							Name: "web",
							Secrets: []*bundlev1.SecretSuffix{
								{
									Suffix: "test",
								},
							},
						},
					},
				},
			},
		}

		// Infrastructure
		f.Fuzz(&tmpl.Spec.Namespaces.Infrastructure[0].Regions[0].Services[0].Secrets[0].Template)
		// Platform
		f.Fuzz(&tmpl.Spec.Namespaces.Platform[0].Components[0].Secrets[0].Template)
		// Product
		f.Fuzz(&tmpl.Spec.Namespaces.Product[0].Secrets[0].Template)
		// Application
		f.Fuzz(&tmpl.Spec.Namespaces.Application[0].Secrets[0].Template)

		v.Visit(tmpl)
	}
}

func TestVisit_Content_Fuzz(t *testing.T) {
	// Making sure the descrption never panics
	for i := 0; i < 50; i++ {
		f := fuzz.New().NilChance(0).NumElements(1, 5)

		v := New(&bundlev1.Bundle{}, engine.NewContext())

		tmpl := &bundlev1.Template{
			ApiVersion: "harp.elastic.co/v1",
			Kind:       "BundleTemplate",
			Spec: &bundlev1.TemplateSpec{
				Selector: &bundlev1.Selector{
					Quality:     "production",
					Product:     "harp",
					Application: "harp",
					Version:     "v1.0.0",
					Platform:    "test",
					Component:   "cli",
				},
				Namespaces: &bundlev1.Namespaces{
					Infrastructure: []*bundlev1.InfrastructureNS{
						{
							Provider: "aws",
							Account:  "test",
							Regions: []*bundlev1.InfrastructureRegionNS{
								{
									Name: "eu-central-1",
									Services: []*bundlev1.InfrastructureServiceNS{
										{
											Name: "ssh",
											Type: "ec2",
											Secrets: []*bundlev1.SecretSuffix{
												{
													Suffix:  "test",
													Content: map[string]string{},
												},
											},
										},
									},
								},
							},
						},
					},
					Platform: []*bundlev1.PlatformRegionNS{
						{
							Region: "eu-central-1",
							Components: []*bundlev1.PlatformComponentNS{
								{
									Type: "rds",
									Name: "postgres-1",
									Secrets: []*bundlev1.SecretSuffix{
										{
											Suffix:  "test",
											Content: map[string]string{},
										},
									},
								},
							},
						},
					},
					Product: []*bundlev1.ProductComponentNS{
						{
							Type: "service",
							Name: "rest-api",
							Secrets: []*bundlev1.SecretSuffix{
								{
									Suffix:  "test",
									Content: map[string]string{},
								},
							},
						},
					},
					Application: []*bundlev1.ApplicationComponentNS{
						{
							Type: "service",
							Name: "web",
							Secrets: []*bundlev1.SecretSuffix{
								{
									Suffix:  "test",
									Content: map[string]string{},
								},
							},
						},
					},
				},
			},
		}

		// Infrastructure
		f.Fuzz(&tmpl.Spec.Namespaces.Infrastructure[0].Regions[0].Services[0].Secrets[0].Content)
		// Platform
		f.Fuzz(&tmpl.Spec.Namespaces.Platform[0].Components[0].Secrets[0].Content)
		// Product
		f.Fuzz(&tmpl.Spec.Namespaces.Product[0].Secrets[0].Content)
		// Application
		f.Fuzz(&tmpl.Spec.Namespaces.Application[0].Secrets[0].Content)

		v.Visit(tmpl)
	}
}
