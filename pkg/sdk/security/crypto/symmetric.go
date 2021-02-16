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

package crypto

import (
	"encoding/base64"
	"fmt"

	"github.com/awnumar/memguard"
	"github.com/fernet/fernet-go"
)

// -----------------------------------------------------------------------------

// Key generates symmetric encryption keys according to given keyType.
func Key(keyType string) (string, error) {
	switch keyType {
	case "aes:128":
		key := memguard.NewBufferRandom(16).Bytes()
		return base64.StdEncoding.EncodeToString(key), nil
	case "aes:192":
		key := memguard.NewBufferRandom(24).Bytes()
		return base64.StdEncoding.EncodeToString(key), nil
	case "aes:256":
		key := memguard.NewBufferRandom(32).Bytes()
		return base64.StdEncoding.EncodeToString(key), nil
	case "aes:siv":
		key := memguard.NewBufferRandom(64).Bytes()
		return base64.StdEncoding.EncodeToString(key), nil
	case "secretbox":
		key := memguard.NewBufferRandom(32).Bytes()
		return base64.StdEncoding.EncodeToString(key), nil
	case "chacha20":
		key := memguard.NewBufferRandom(32).Bytes()
		return base64.StdEncoding.EncodeToString(key), nil
	case "fernet":
		// Generate a fernet key
		k := &fernet.Key{}
		if err := k.Generate(); err != nil {
			return "", err
		}
		return k.Encode(), nil
	default:
		return "", fmt.Errorf("invalid keytype (%s) [aes:128, aes:192, aes:256, aes:siv, secretbox, chacha20, fernet]", keyType)
	}
}
