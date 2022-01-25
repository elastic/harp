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

package to

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"

	"github.com/google/go-github/v42/github"
	"go.uber.org/zap"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/oauth2"

	"github.com/elastic/harp/pkg/bundle"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/tasks"
)

type GithubActionTask struct {
	_               struct{}
	ContainerReader tasks.ReaderProvider
	Owner           string
	Repository      string
	SecretFilter    string
}

func (t *GithubActionTask) Run(ctx context.Context) error {
	// Create the reader
	reader, err := t.ContainerReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to open input bundle reader: %w", err)
	}

	// Extract bundle from container
	b, err := bundle.FromContainerReader(reader)
	if err != nil {
		return fmt.Errorf("unable to load bundle: %w", err)
	}

	// Retrieve github token
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		return errors.New("GITHUB_TOKEN environment variable must be set")
	}

	// Create an authenticated transport
	tc := oauth2.NewClient(
		ctx,
		oauth2.StaticTokenSource(
			&oauth2.Token{
				AccessToken: githubToken,
			},
		),
	)

	// Create github API client
	client := github.NewClient(tc)

	// Retrieve repository public key
	pub, _, err := client.Actions.GetRepoPublicKey(ctx, t.Owner, t.Repository)
	if err != nil {
		return fmt.Errorf("unable to retireve repository public key for secret encryption: %w", err)
	}

	// Decode public key.
	decodedPublicKey, err := base64.StdEncoding.DecodeString(pub.GetKey())
	if err != nil {
		return fmt.Errorf("unable to decode public key from github: %w", err)
	}

	// The decode key is converted into a fixed size byte array.
	var boxKey [32]byte

	// The secret value is converted into a slice of bytes.
	copy(boxKey[:], decodedPublicKey)

	log.For(ctx).Info("Repository public key feteched", zap.String("key_id", pub.GetKeyID()))

	// Requests to send to github
	githubSecrets := []*github.EncryptedSecret{}

	// Iterate over packages
	for _, p := range b.Packages {
		// Ignore nil secret chain
		if p.Secrets == nil {
			continue
		}

		// Get secrets
		secretMap, err := bundle.AsSecretMap(p)
		if err != nil {
			return fmt.Errorf("unable to retrieve secrets from '%s' package: %w", p.Name, err)
		}

		// Filter secrets map using given filter glob.
		filteredSecrets := secretMap.Glob(t.SecretFilter)

		// Iterate over secrets
		for key, value := range filteredSecrets {
			var secretBytes []byte

			// Pack the secret value
			switch v := value.(type) {
			case string:
				secretBytes = []byte(v)
			case []byte:
				secretBytes = v
			default:
				return fmt.Errorf("can't process secret type %T", value)
			}

			// The secret is encrypted with box.SealAnonymous using the repo's decoded public key.
			encryptedBytes, err := box.SealAnonymous([]byte{}, secretBytes, &boxKey, rand.Reader)
			if err != nil {
				return fmt.Errorf("unable to encrypt the secret payload: %w", err)
			}

			// Prepare the request
			githubSecrets = append(githubSecrets, &github.EncryptedSecret{
				Name:           key,
				KeyID:          pub.GetKeyID(),
				EncryptedValue: base64.StdEncoding.EncodeToString(encryptedBytes),
			})
		}
	}

	// Publish all secrets
	for _, gs := range githubSecrets {
		// Create or update the secret value.
		if _, err := client.Actions.CreateOrUpdateRepoSecret(ctx, t.Owner, t.Repository, gs); err != nil {
			return fmt.Errorf("unable to publish secret to github: %v", err)
		}
	}

	// No error
	return nil
}
