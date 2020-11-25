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

package bundle

import (
	"io"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
)

// Reader exposes bundle reader contract
type Reader interface {
	Read(reader io.Reader) (*bundlev1.Bundle, error)
}

// Writer exposes bundle writer contract
type Writer interface {
	Write(file *bundlev1.Bundle) error
}

// Visitor delares the bundle vistor contract.
type Visitor interface {
	Error() error
	VisitForFile(obj *bundlev1.Bundle)
	VisitForPackage(obj *bundlev1.Package)
	VisitForChain(obj *bundlev1.SecretChain)
	VisitForKV(obj *bundlev1.KV)
}
