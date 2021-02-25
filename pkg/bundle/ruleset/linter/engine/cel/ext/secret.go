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
	"encoding/json"
	"fmt"
	"reflect"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/interpreter/functions"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle/secret"
)

// Secrets exported secret operations.
func Secrets() cel.EnvOption {
	return cel.Lib(secretLib{})
}

type secretLib struct{}

func (secretLib) CompileOptions() []cel.EnvOption {
	return []cel.EnvOption{
		cel.Declarations(
			decls.NewFunction("is_base64",
				decls.NewInstanceOverload("kv_is_base64",
					[]*exprpb.Type{harpKVObjectType},
					decls.Bool,
				),
			),
			decls.NewFunction("is_required",
				decls.NewInstanceOverload("kv_is_required",
					[]*exprpb.Type{harpKVObjectType},
					decls.Bool,
				),
			),
			decls.NewFunction("is_url",
				decls.NewInstanceOverload("kv_is_url",
					[]*exprpb.Type{harpKVObjectType},
					decls.Bool,
				),
			),
			decls.NewFunction("is_uuid",
				decls.NewInstanceOverload("kv_is_uuid",
					[]*exprpb.Type{harpKVObjectType},
					decls.Bool,
				),
			),
			decls.NewFunction("is_email",
				decls.NewInstanceOverload("kv_is_email",
					[]*exprpb.Type{harpKVObjectType},
					decls.Bool,
				),
			),
			decls.NewFunction("is_json",
				decls.NewInstanceOverload("kv_is_json",
					[]*exprpb.Type{harpKVObjectType},
					decls.Bool,
				),
			),
		),
	}
}

func (secretLib) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{
		cel.Functions(
			&functions.Overload{
				Operator: "kv_is_base64",
				Unary:    celValidatorBuilder(is.Base64),
			},
			&functions.Overload{
				Operator: "kv_is_required",
				Unary:    celValidatorBuilder(validation.Required),
			},
			&functions.Overload{
				Operator: "kv_is_url",
				Unary:    celValidatorBuilder(is.URL),
			},
			&functions.Overload{
				Operator: "kv_is_uuid",
				Unary:    celValidatorBuilder(is.UUID),
			},
			&functions.Overload{
				Operator: "kv_is_email",
				Unary:    celValidatorBuilder(is.EmailFormat),
			},
			&functions.Overload{
				Operator: "kv_is_json",
				Unary:    celValidatorBuilder(&jsonValidator{}),
			},
		),
	}
}

// -----------------------------------------------------------------------------

func celValidatorBuilder(rules ...validation.Rule) func(ref.Val) ref.Val {
	return func(lhs ref.Val) ref.Val {
		x, _ := lhs.ConvertToNative(reflect.TypeOf(&bundlev1.KV{}))
		p := x.(*bundlev1.KV)

		var out string
		if err := secret.Unpack(p.Value, &out); err != nil {
			return types.Bool(false)
		}

		if err := validation.Validate(out, rules...); err != nil {
			return types.Bool(false)
		}

		return types.Bool(true)
	}
}

// -----------------------------------------------------------------------------

var _ validation.Rule = (*jsonValidator)(nil)

type jsonValidator struct{}

func (v *jsonValidator) Validate(in interface{}) error {
	// Process input
	if data, ok := in.([]byte); ok {
		if !json.Valid(data) {
			return fmt.Errorf("invalid JSON payload")
		}
	}

	return fmt.Errorf("unable to validate JSON for %T type", in)
}
