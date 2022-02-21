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

package transit

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/hashicorp/vault/api"

	"github.com/elastic/harp/pkg/vault/logical"
	vpath "github.com/elastic/harp/pkg/vault/path"
)

type service struct {
	logical   logical.Logical
	mountPath string
	keyName   string
}

// New instantiates a Vault transit backend encryption service.
func New(client *api.Client, mountPath, keyName string) (Service, error) {
	return &service{
		logical:   client.Logical(),
		mountPath: strings.TrimSuffix(path.Clean(mountPath), "/"),
		keyName:   keyName,
	}, nil
}

// -----------------------------------------------------------------------------

func (s *service) Encrypt(ctx context.Context, cleartext []byte) ([]byte, error) {
	// Prepare query
	encryptPath := vpath.SanitizePath(path.Join(url.PathEscape(s.mountPath), "encrypt", url.PathEscape(s.keyName)))
	data := map[string]interface{}{
		"plaintext": base64.StdEncoding.EncodeToString(cleartext),
	}

	// Send to Vault.
	secret, err := s.logical.Write(encryptPath, data)
	if err != nil {
		return nil, fmt.Errorf("unable to encrypt with '%s' key: %w", s.keyName, err)
	}

	// Check response wrapping
	if secret.WrapInfo != nil {
		// Unwrap with response token
		secret, err = s.logical.Unwrap(secret.WrapInfo.Token)
		if err != nil {
			return nil, fmt.Errorf("unable to unwrap the response: %w", err)
		}
	}

	// Parse server response.
	if cipherText, ok := secret.Data["ciphertext"].(string); ok && cipherText != "" {
		return []byte(cipherText), nil
	}

	// Return error.
	return nil, errors.New("could not encrypt given data")
}

func (s *service) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	// Prepare query
	decryptPath := vpath.SanitizePath(path.Join(url.PathEscape(s.mountPath), "decrypt", url.PathEscape(s.keyName)))
	data := map[string]interface{}{
		"ciphertext": string(ciphertext),
	}

	// Send to Vault.
	secret, err := s.logical.Write(decryptPath, data)
	if err != nil {
		return nil, fmt.Errorf("unable to decrypt with '%s' key: %w", s.keyName, err)
	}

	// Check response wrapping
	if secret.WrapInfo != nil {
		// Unwrap with response token
		secret, err = s.logical.Unwrap(secret.WrapInfo.Token)
		if err != nil {
			return nil, fmt.Errorf("unable to unwrap the response: %w", err)
		}
	}

	// Parse server response.
	if plainText64, ok := secret.Data["plaintext"].(string); ok && plainText64 != "" {
		plainText, err := base64.StdEncoding.DecodeString(plainText64)
		if err != nil {
			return nil, fmt.Errorf("unable to decode secret: %w", err)
		}

		// Return no error
		return plainText, nil
	}

	// Return error.
	return nil, errors.New("could not decrypt given data")
}
