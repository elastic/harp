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

package engine

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/alessio/shellescape"

	"github.com/elastic/harp/pkg/sdk/security/crypto"
	"github.com/elastic/harp/pkg/sdk/security/crypto/bech32"
	"github.com/elastic/harp/pkg/sdk/security/diceware"
	"github.com/elastic/harp/pkg/sdk/security/password"
	"github.com/elastic/harp/pkg/template/engine/internal/codec"
)

// FuncMap returns a mapping of all of the functions that Temmplate has.
func FuncMap(secretReaders []SecretReaderFunc) template.FuncMap {
	f := sprig.TxtFuncMap()

	// Add some extra functionality
	extra := template.FuncMap{
		// Password
		"customPassword":   password.Generate,
		"paranoidPassword": password.Paranoid,
		"noSymbolPassword": password.NoSymbol,
		"strongPassword":   password.Strong,
		// Diceware
		"customDiceware":   diceware.Diceware,
		"basicDiceware":    diceware.Basic,
		"strongDiceware":   diceware.Strong,
		"paranoidDiceware": diceware.Paranoid,
		// Encoder
		"toToml":        codec.ToTOML,
		"toYaml":        codec.ToYAML,
		"fromYaml":      codec.FromYAML,
		"fromYamlArray": codec.FromYAMLArray,
		"toJson":        codec.ToJSON,
		"fromJson":      codec.FromJSON,
		"fromJsonArray": codec.FromJSONArray,
		// Crypto
		"toJwk":      crypto.ToJWK,
		"fromJwk":    crypto.FromJWK,
		"toPem":      crypto.ToPEM,
		"encryptPem": crypto.EncryptPEM,
		"toSSH":      crypto.ToSSH,
		"cryptoKey":  crypto.Key,
		"cryptoPair": crypto.Keypair,
		"keyToBytes": crypto.KeyToBytes,
		// Secret
		"secret": SecretReaders(secretReaders),
		// JWT/JWE
		"encryptJwe": crypto.EncryptJWE,
		"decryptJwe": crypto.DecryptJWE,
		"toJws":      crypto.ToJWS,
		"parseJwt":   crypto.ParseJWT,
		"verifyJwt":  crypto.VerifyJWT,
		// Hex
		"hexenc": hex.EncodeToString,
		"hexdec": hex.DecodeString,
		// Bech32
		"bech32enc": bech32.Encode,
		"bech32dec": crypto.Bech32Decode,
		// Base64
		"b64urlenc": func(in string) string {
			return base64.URLEncoding.EncodeToString([]byte(in))
		},
		"b64urldec": func(in string) (string, error) {
			out, err := base64.URLEncoding.DecodeString(in)
			return string(out), err
		},
		// Escaping
		"urlPathEscape":    url.PathEscape,
		"urlPathUnescape":  url.PathUnescape,
		"urlQueryEscape":   url.QueryEscape,
		"urlQueryUnescape": url.QueryUnescape,
		"shellEscape":      shellescape.Quote,
		"jsonEscape": func(in string) (string, error) {
			b, err := json.Marshal(in)
			// Trim the beginning and trailing " character
			return strings.Trim(string(b), `"`), err
		},
		"jsonUnescape": func(in string) (string, error) {
			var out string
			if err := json.Unmarshal([]byte(in), &out); err != nil {
				return "", err
			}
			return out, nil
		},
		"unquote": strconv.Unquote,
	}

	for k, v := range extra {
		f[k] = v
	}

	return f
}
