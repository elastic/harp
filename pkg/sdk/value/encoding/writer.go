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

package encoding

import (
	"encoding/ascii85"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/elastic/harp/pkg/sdk/ioutil"
)

// -----------------------------------------------------------------------------

// NewWriter returns the appropriate writer implementation according to given encoding.
func NewWriter(w io.Writer, encoding string) (io.WriteCloser, error) {
	// Normalize input
	encoding = strings.TrimSpace(strings.ToLower(encoding))

	var encoderWriter io.WriteCloser

	// Apply transformation
	switch encoding {
	case "identity":
		encoderWriter = ioutil.NopCloserWriter(w)
	case "hex", "base16":
		encoderWriter = ioutil.NopCloserWriter(hex.NewEncoder(w))
	case "base32":
		encoderWriter = base32.NewEncoder(base32.StdEncoding, w)
	case "base32hex":
		encoderWriter = base32.NewEncoder(base32.HexEncoding, w)
	case "base64":
		encoderWriter = base64.NewEncoder(base64.StdEncoding, w)
	case "base64raw":
		encoderWriter = base64.NewEncoder(base64.RawStdEncoding, w)
	case "base64url":
		encoderWriter = base64.NewEncoder(base64.URLEncoding, w)
	case "base64urlraw":
		encoderWriter = base64.NewEncoder(base64.RawURLEncoding, w)
	case "base85":
		encoderWriter = ascii85.NewEncoder(w)
	default:
		return nil, fmt.Errorf("unhandled encoding strategy '%s'", encoding)
	}

	// No error
	return encoderWriter, nil
}
