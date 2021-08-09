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

package flatmap

import (
	"fmt"
	"path"
	"reflect"

	"github.com/elastic/harp/pkg/bundle"
)

// -----------------------------------------------------------------------------

// Flatten takes a structure and turns into a flat map[string]string.
func Flatten(thing map[string]interface{}) map[string]bundle.KV {
	result := make(map[string]string)

	// Flatten recursively the map
	for k, raw := range thing {
		flatten(result, k, reflect.ValueOf(raw))
	}

	// Unpack leaf as secrets
	jsonMap := map[string]bundle.KV{}
	for k, v := range result {
		// Get last element as secret name
		packageName, secretName := path.Split(k)

		// Remove trailing path separator
		packageName = path.Clean(packageName)

		// Check if package already is registered
		p, ok := jsonMap[packageName]
		if !ok {
			p = bundle.KV{}
		}

		// Assign secret
		p[secretName] = v

		// Re-assign to map
		jsonMap[packageName] = p
	}

	// Return json map
	return jsonMap
}

func flatten(result map[string]string, prefix string, v reflect.Value) {
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() {
			result[prefix] = "true"
		} else {
			result[prefix] = "false"
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		result[prefix] = fmt.Sprintf("%d", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		result[prefix] = fmt.Sprintf("%d", v.Uint())
	case reflect.Float32, reflect.Float64:
		result[prefix] = fmt.Sprintf("%f", v.Float())
	case reflect.Map:
		flattenMap(result, prefix, v)
	case reflect.Slice, reflect.Array:
		flattenSlice(result, prefix, v)
	case reflect.String:
		result[prefix] = v.String()
	case reflect.Chan, reflect.Complex128, reflect.Complex64, reflect.Func, reflect.Interface:
		// ignore
	case reflect.Invalid, reflect.Ptr, reflect.Struct, reflect.Uintptr, reflect.UnsafePointer:
		// ignore
	default:
		panic(fmt.Sprintf("Unknown: %s", v))
	}
}

func flattenMap(result map[string]string, prefix string, v reflect.Value) {
	for _, k := range v.MapKeys() {
		if k.Kind() == reflect.Interface {
			k = k.Elem()
		}

		if k.Kind() != reflect.String {
			panic(fmt.Sprintf("%s: map key is not string: %s", prefix, k))
		}

		flatten(result, fmt.Sprintf("%s/%s", prefix, k.String()), v.MapIndex(k))
	}
}

func flattenSlice(result map[string]string, prefix string, v reflect.Value) {
	prefix += "/"

	result[prefix+"#"] = fmt.Sprintf("%d", v.Len())
	for i := 0; i < v.Len(); i++ {
		flatten(result, fmt.Sprintf("%s%d", prefix, i), v.Index(i))
	}
}
