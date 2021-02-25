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

package ext

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gobwas/glob"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/interpreter/functions"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	csov1 "github.com/elastic/harp/pkg/cso/v1"
	htypes "github.com/elastic/harp/pkg/sdk/types"
)

// Packages exported package operations.
func Packages() cel.EnvOption {
	return cel.Lib(packageLib{})
}

type packageLib struct{}

func (packageLib) CompileOptions() []cel.EnvOption {
	return []cel.EnvOption{
		cel.Declarations(
			decls.NewVar("p", harpPackageObjectType),
			decls.NewFunction("match_path",
				decls.NewInstanceOverload("package_match_path_string",
					[]*exprpb.Type{harpPackageObjectType, decls.String},
					decls.Bool,
				),
			),
			decls.NewFunction("has_secret",
				decls.NewInstanceOverload("package_has_secret_string",
					[]*exprpb.Type{harpPackageObjectType, decls.String},
					decls.Bool,
				),
			),
			decls.NewFunction("has_all_secrets",
				decls.NewInstanceOverload("package_has_all_secrets_list",
					[]*exprpb.Type{harpPackageObjectType, decls.NewListType(decls.String)},
					decls.Bool,
				),
			),
			decls.NewFunction("is_cso_compliant",
				decls.NewInstanceOverload("package_is_cso_compliant",
					[]*exprpb.Type{harpPackageObjectType},
					decls.Bool,
				),
			),
			decls.NewFunction("secret",
				decls.NewInstanceOverload("package_secret_string",
					[]*exprpb.Type{harpPackageObjectType, decls.String},
					harpKVObjectType,
				),
			),
		),
	}
}

func (packageLib) ProgramOptions() []cel.ProgramOption {
	// Register types
	reg, err := types.NewRegistry(
		&bundlev1.KV{},
	)
	if err != nil {
		panic(fmt.Errorf("unable to register types: %w", err))
	}

	return []cel.ProgramOption{
		cel.Functions(
			&functions.Overload{
				Operator: "package_match_path_string",
				Binary:   celPackageMatchPath,
			},
			&functions.Overload{
				Operator: "package_has_secret_string",
				Binary:   celPackageHasSecret,
			},
			&functions.Overload{
				Operator: "package_has_all_secrets_list",
				Binary:   celPackageHasAllSecrets,
			},
			&functions.Overload{
				Operator: "package_is_cso_compliant",
				Unary:    celPackageIsCSOCompliant,
			},
			&functions.Overload{
				Operator: "package_secret_string",
				Binary:   celPackageGetSecret(reg),
			},
		),
	}
}

// -----------------------------------------------------------------------------

func celPackageMatchPath(lhs, rhs ref.Val) ref.Val {
	x, _ := lhs.ConvertToNative(reflect.TypeOf(&bundlev1.Package{}))
	p := x.(*bundlev1.Package)

	pathTyped := rhs.(types.String)
	path := pathTyped.Value().(string)

	return types.Bool(glob.MustCompile(path).Match(p.Name))
}

func celPackageHasSecret(lhs, rhs ref.Val) ref.Val {
	x, _ := lhs.ConvertToNative(reflect.TypeOf(&bundlev1.Package{}))
	p := x.(*bundlev1.Package)
	secretTyped := rhs.(types.String)
	secretName := secretTyped.Value().(string)

	// No secret data
	if p.Secrets == nil || p.Secrets.Data == nil || len(p.Secrets.Data) == 0 {
		return types.Bool(false)
	}

	// Look for secret name
	for _, k := range p.Secrets.Data {
		if strings.EqualFold(k.Key, secretName) {
			return types.Bool(true)
		}
	}

	return types.Bool(false)
}

func celPackageHasAllSecrets(lhs, rhs ref.Val) ref.Val {
	x, _ := lhs.ConvertToNative(reflect.TypeOf(&bundlev1.Package{}))
	p := x.(*bundlev1.Package)
	secretsTyped, _ := rhs.ConvertToNative(reflect.TypeOf([]string{}))
	secretNames := secretsTyped.([]string)

	// No secret data
	if p.Secrets == nil || p.Secrets.Data == nil || len(p.Secrets.Data) == 0 {
		return types.Bool(false)
	}

	sa := htypes.StringArray(secretNames)

	secretMap := map[string]*bundlev1.KV{}
	for _, k := range p.Secrets.Data {
		if !sa.Contains(k.Key) {
			return types.Bool(false)
		}
		secretMap[k.Key] = k
	}

	// Look for secret name
	for _, k := range secretNames {
		if _, ok := secretMap[k]; !ok {
			return types.Bool(false)
		}
	}

	return types.Bool(true)
}

func celPackageIsCSOCompliant(lhs ref.Val) ref.Val {
	x, _ := lhs.ConvertToNative(reflect.TypeOf(&bundlev1.Package{}))
	p := x.(*bundlev1.Package)

	if err := csov1.Validate(p.Name); err != nil {
		return types.Bool(false)
	}

	return types.Bool(true)
}

func celPackageGetSecret(reg ref.TypeAdapter) func(lhs, rhs ref.Val) ref.Val {
	return func(lhs, rhs ref.Val) ref.Val {
		x, _ := lhs.ConvertToNative(reflect.TypeOf(&bundlev1.Package{}))
		p := x.(*bundlev1.Package)
		secretTyped := rhs.(types.String)
		secretName := secretTyped.Value().(string)

		// No secret data
		if p.Secrets == nil || p.Secrets.Data == nil || len(p.Secrets.Data) == 0 {
			return types.Bool(false)
		}

		// Look for secret name
		for _, k := range p.Secrets.Data {
			if strings.EqualFold(k.Key, secretName) {
				return reg.NativeToValue(k)
			}
		}

		return nil
	}
}
