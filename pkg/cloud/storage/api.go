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

package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"
)

// -----------------------------------------------------------------------------

// Backend is a generic interface for storage backends
type Backend interface {
	GetObject(ctx context.Context, path string) (*Object, error)
}

// -----------------------------------------------------------------------------

// Object is a generic representation of a storage object
type Object struct {
	Path         string
	Content      io.ReadCloser
	LastModified time.Time
}

// HasExtension determines whether or not an object contains a file extension
func (object *Object) HasExtension(extension string) bool {
	return filepath.Ext(object.Path) == fmt.Sprintf(".%s", extension)
}
