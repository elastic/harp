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

package age

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"filippo.io/age"
	"filippo.io/age/armor"

	"github.com/elastic/harp/pkg/sdk/value"
	"github.com/elastic/harp/pkg/sdk/value/encryption"
)

const (
	agePublicPrefix  = "age-recipients"
	agePrivatePrefix = "age-identity"
)

func init() {
	encryption.Register(agePublicPrefix, Transformer)
	encryption.Register(agePrivatePrefix, Transformer)
}

// Transformer returns a fernet encryption transformer
func Transformer(key string) (value.Transformer, error) {
	switch {
	case strings.HasPrefix(key, "age-recipients:"):
		// Remove the prefix
		key = strings.TrimPrefix(key, "age-recipients:")

		// Split recipients
		recipientRaw := strings.Split(key, ":")

		recipients := []age.Recipient{}
		for _, r := range recipientRaw {
			// Check given keys
			k, err := age.ParseX25519Recipient(r)
			if err != nil {
				return nil, fmt.Errorf("age: unable to initialize age transformer, '%s' is an invalid recipient: %w", r, err)
			}

			// Add to recipients
			recipients = append(recipients, k)
		}

		// Return decorator constructor
		return &ageTransformer{
			recipients: recipients,
		}, nil
	case strings.HasPrefix(key, "age-identity:"):
		// Remove the prefix
		key = strings.TrimPrefix(key, "age-identity:")

		// Split identities
		identityRaw := strings.Split(key, ":")

		identities := []age.Identity{}
		for _, r := range identityRaw {
			// Check given keys
			k, err := age.ParseX25519Identity(r)
			if err != nil {
				return nil, fmt.Errorf("age: unable to initialize age transformer, '%s' is an invalid identity: %w", r, err)
			}

			// Add to identities
			identities = append(identities, k)
		}

		// Return decorator constructor
		return &ageTransformer{
			identities: identities,
		}, nil
	}

	// Default to error
	return nil, errors.New("age: prefix not supported")
}

// -----------------------------------------------------------------------------

type ageTransformer struct {
	recipients []age.Recipient
	identities []age.Identity
}

func (d *ageTransformer) To(_ context.Context, input []byte) ([]byte, error) {
	var (
		in  = bytes.NewReader(input)
		buf = &bytes.Buffer{}
	)

	// Check recipients count
	if len(d.recipients) == 0 {
		return nil, errors.New("no recipients specified")
	}

	// Amrmor writer
	a := armor.NewWriter(buf)

	// Encrypt with given recipients
	w, err := age.Encrypt(a, d.recipients...)
	if err != nil {
		return nil, err
	}

	// Copy stream
	if _, err := io.Copy(w, in); err != nil {
		return nil, err
	}

	// Close the writer
	if err := w.Close(); err != nil {
		return nil, err
	}

	// Close armor writer
	if err := a.Close(); err != nil {
		return nil, err
	}

	// No error
	return buf.Bytes(), nil
}

func (d *ageTransformer) From(_ context.Context, input []byte) ([]byte, error) {
	var (
		in  io.Reader = bytes.NewReader(input)
		out bytes.Buffer
	)

	// Check identities count
	if len(d.identities) == 0 {
		return nil, errors.New("no identities specified")
	}

	// Check armor usage
	rr := bufio.NewReader(in)
	if start, _ := rr.Peek(len(armor.Header)); string(start) == armor.Header {
		in = armor.NewReader(rr)
	} else {
		in = rr
	}

	// Decrypt with given identities
	w, err := age.Decrypt(in, d.identities...)
	if err != nil {
		return nil, err
	}

	// Copy stream
	if _, err := io.Copy(&out, w); err != nil {
		return nil, err
	}

	// No error
	return out.Bytes(), nil
}
