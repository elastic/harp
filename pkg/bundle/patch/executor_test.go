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

package patch

import (
	"testing"

	fuzz "github.com/google/gofuzz"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
)

func Test_executeRule_Fuzz(t *testing.T) {
	// Making sure the executeRule never panics
	for i := 0; i < 50; i++ {
		f := fuzz.New()

		// Prepare arguments
		values := map[string]interface{}{}
		spec := &bundlev1.Patch{
			ApiVersion: "harp.elastic.co/v1",
			Kind:       "BundlePatch",
			Meta: &bundlev1.PatchMeta{
				Name: "test-patch",
			},
			Spec: &bundlev1.PatchSpec{
				Rules: []*bundlev1.PatchRule{
					{
						Package:  &bundlev1.PatchPackage{},
						Selector: &bundlev1.PatchSelector{},
					},
				},
			},
		}
		p := bundlev1.Package{
			Name: "foo",
			Secrets: &bundlev1.SecretChain{
				Data: []*bundlev1.KV{
					{
						Key:   "k1",
						Value: []byte("v1"),
					},
				},
			},
		}

		var patchName string

		f.Fuzz(&patchName)
		f.Fuzz(&spec.Spec.Rules[0])

		// Execute
		executeRule(spec.Spec.Rules[0], &p, values)
	}
}

func Test_compileSelector_Fuzz(t *testing.T) {
	// Making sure the compileSelector never panics
	for i := 0; i < 50; i++ {
		f := fuzz.New()

		// Prepare arguments
		values := map[string]interface{}{}
		spec := &bundlev1.Patch{
			ApiVersion: "harp.elastic.co/v1",
			Kind:       "BundlePatch",
			Meta: &bundlev1.PatchMeta{
				Name: "test-patch",
			},
			Spec: &bundlev1.PatchSpec{
				Rules: []*bundlev1.PatchRule{
					{
						Package:  &bundlev1.PatchPackage{},
						Selector: &bundlev1.PatchSelector{},
					},
				},
			},
		}

		f.Fuzz(&spec.Spec.Rules[0].Selector)

		// Execute
		compileSelector(spec.Spec.Rules[0].Selector, values)
	}
}

func Test_applyPackagePatch_Fuzz(t *testing.T) {
	// Making sure the applyPatchPackage never panics
	for i := 0; i < 500; i++ {
		f := fuzz.New()

		// Prepare arguments
		values := map[string]interface{}{}
		spec := &bundlev1.Patch{
			ApiVersion: "harp.elastic.co/v1",
			Kind:       "BundlePatch",
			Meta: &bundlev1.PatchMeta{
				Name: "test-patch",
			},
			Spec: &bundlev1.PatchSpec{
				Rules: []*bundlev1.PatchRule{
					{
						Package:  &bundlev1.PatchPackage{},
						Selector: &bundlev1.PatchSelector{},
					},
				},
			},
		}
		p := bundlev1.Package{
			Name: "foo",
			Secrets: &bundlev1.SecretChain{
				Data: []*bundlev1.KV{
					{
						Key:   "k1",
						Value: []byte("v1"),
					},
				},
			},
		}

		f.Fuzz(&spec.Spec.Rules[0].Package)

		// Execute
		applyPackagePatch(&p, spec.Spec.Rules[0].Package, values)
	}
}

func Test_applySecretPatch_Fuzz(t *testing.T) {
	// Making sure the applyPatchPackage never panics
	for i := 0; i < 500; i++ {
		f := fuzz.New()

		// Prepare arguments
		values := map[string]interface{}{
			"foo": "test",
		}
		spec := &bundlev1.Patch{
			ApiVersion: "harp.elastic.co/v1",
			Kind:       "BundlePatch",
			Meta: &bundlev1.PatchMeta{
				Name: "test-patch",
			},
			Spec: &bundlev1.PatchSpec{
				Rules: []*bundlev1.PatchRule{
					{
						Package: &bundlev1.PatchPackage{},
						Selector: &bundlev1.PatchSelector{
							MatchPath: &bundlev1.PatchSelectorMatchPath{
								Strict: "foo",
							},
						},
					},
				},
			},
		}
		file := bundlev1.Bundle{
			Packages: []*bundlev1.Package{
				{
					Name: "foo",
					Secrets: &bundlev1.SecretChain{
						Data: []*bundlev1.KV{
							{
								Key:   "k1",
								Value: []byte("v1"),
							},
						},
					},
				},
			},
		}

		f.Fuzz(&spec.Spec.Rules[0].Package.Data)
		f.Fuzz(&file.Packages[0].Secrets)

		// Execute
		applySecretPatch(file.Packages[0].Secrets, spec.Spec.Rules[0].Package.Data, values)
	}
}

func Test_applySecretKVPatch_Fuzz(t *testing.T) {
	// Making sure the applySecretKVPatch never panics
	for i := 0; i < 500; i++ {
		f := fuzz.New()

		// Prepare arguments
		values := map[string]interface{}{}
		spec := &bundlev1.PatchOperation{
			Add:    map[string]string{},
			Remove: []string{},
			Update: map[string]string{},
		}
		file := bundlev1.Bundle{
			Packages: []*bundlev1.Package{
				{
					Name: "foo",
					Secrets: &bundlev1.SecretChain{
						Data: []*bundlev1.KV{
							{
								Key:   "k1",
								Value: []byte("v1"),
							},
						},
					},
				},
			},
		}

		f.Fuzz(&file.Packages[0].Secrets.Data)
		f.Fuzz(&spec)

		// Execute
		applySecretKVPatch(file.Packages[0].Secrets.Data, spec, values)
	}
}
