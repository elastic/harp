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
)

// -----------------------------------------------------------------------------

// NewReader returns a reader implementation matching the given encoding strategy.
func NewReader(r io.Reader, encoding string) (io.Reader, error) {
	// Normalize input
	encoding = strings.TrimSpace(strings.ToLower(encoding))

	var (
		decoderReader io.Reader
	)

	// Apply transformation
	switch encoding {
	case "identity":
		decoderReader = r
	case "hex", "base16":
		decoderReader = hex.NewDecoder(r)
	case "base32":
		decoderReader = base32.NewDecoder(base32.StdEncoding, r)
	case "base32hex":
		decoderReader = base32.NewDecoder(base32.HexEncoding, r)
	case "base64":
		decoderReader = base64.NewDecoder(base64.StdEncoding, r)
	case "base64raw":
		decoderReader = base64.NewDecoder(base64.RawStdEncoding, r)
	case "base64url":
		decoderReader = base64.NewDecoder(base64.URLEncoding, r)
	case "base64urlraw":
		decoderReader = base64.NewDecoder(base64.RawURLEncoding, r)
	case "base85":
		decoderReader = ascii85.NewDecoder(r)
	default:
		return nil, fmt.Errorf("unhandled decoding strategy '%s'", encoding)
	}

	// No error
	return decoderReader, nil
}
