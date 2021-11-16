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

package vault

import (
	"fmt"
	"path"
	"strings"

	"github.com/elastic/harp/pkg/sdk/value"
	"github.com/elastic/harp/pkg/sdk/value/encryption"
	"github.com/elastic/harp/pkg/sdk/value/encryption/aead"
	"github.com/elastic/harp/pkg/sdk/value/encryption/envelope"
	"github.com/elastic/harp/pkg/sdk/value/encryption/secretbox"
)

type DataEncryption string

var (
	AESGCM           DataEncryption = "aesgcm"
	Chacha20Poly1305 DataEncryption = "chacha20poly1305"
	Secretbox        DataEncryption = "secretbox"
)

func init() {
	encryption.Register("vault", Transformer)
}

// Vault returns an envelope encryption using a remote transit backend for key
// encryption.
// vault:<path>:<data encryption>
func Transformer(key string) (value.Transformer, error) {
	// Remove the prefix
	key = strings.TrimPrefix(key, "vault:")

	// Split path / encryption
	parts := strings.SplitN(key, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("key format error, invalid part count")
	}

	// Create default vault client
	client, err := DefaultClient()
	if err != nil {
		return nil, fmt.Errorf("unable to initialize vault client: %w", err)
	}

	// Split transit backend path
	mountPath, keyName := path.Split(parts[0])

	// Create transit backend service
	backend, err := client.Transit(path.Clean(mountPath), keyName)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize vault transit backend service: %w", err)
	}

	// Prepare data encryption
	var dataEncryptionFunc encryption.TransformerFactoryFunc
	dataEncryptionMethod := strings.TrimSpace(strings.ToLower(parts[1]))
	switch dataEncryptionMethod {
	case string(AESGCM):
		dataEncryptionFunc = aead.AESGCM
	case string(Chacha20Poly1305):
		dataEncryptionFunc = aead.Chacha20Poly1305
	case string(Secretbox):
		dataEncryptionFunc = secretbox.Transformer
	default:
		return nil, fmt.Errorf("unsupported data encryption '%s' for envelope transformer", dataEncryptionMethod)
	}

	// Wrap the transformer with envelope
	return envelope.Transformer(backend, dataEncryptionFunc)
}

func TransformerKey(mountPath, keyName string, dataEncryption DataEncryption) string {
	return fmt.Sprintf("vault:%s/%s:%s", strings.TrimSuffix(path.Clean(mountPath), "/"), strings.TrimPrefix(keyName, "/"), dataEncryption)
}
