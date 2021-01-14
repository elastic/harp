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

package container

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/awnumar/memguard"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/afero"
	"go.uber.org/zap"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	containerv1 "github.com/elastic/harp/api/gen/go/harp/container/v1"
	"github.com/elastic/harp/pkg/bundle"
	"github.com/elastic/harp/pkg/bundle/vfs"
	"github.com/elastic/harp/pkg/cloud/aws/session"
	"github.com/elastic/harp/pkg/container"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/sdk/value/encryption"
	"github.com/elastic/harp/pkg/server/storage"
)

// -----------------------------------------------------------------------------

// SetKeyring assigns the container keyring for bundle loader.
func SetKeyring(keys []string) {
	containerKeyring = keys
}

// -----------------------------------------------------------------------------

var (
	once             sync.Once
	containerKeyring []string
)

const (
	schemeBundleDefault    = "bundle"
	schemeBundleStdin      = "bundle+stdin"
	schemeBundleFromFile   = "bundle+file"
	schemeBundleFromHTTP   = "bundle+http"
	schemeBundleFromHTTPS  = "bundle+https"
	schemeBundleFromS3     = "bundle+s3"
	schemeBundleFromGCS    = "bundle+gcs"
	schemeBundleFromAzBlob = "bundle+azblob"
)

func init() {
	once.Do(func() {
		// Register to storage factory
		storage.MustRegister(schemeBundleDefault, build)
		storage.MustRegister(schemeBundleFromFile, build)
		storage.MustRegister(schemeBundleFromHTTP, build)
		storage.MustRegister(schemeBundleFromS3, build)
		storage.MustRegister(schemeBundleFromGCS, build)
		storage.MustRegister(schemeBundleFromAzBlob, build)
		storage.MustRegister(schemeBundleStdin, build)
	})
}

// -----------------------------------------------------------------------------

func withDefault(q url.Values, key, defaultValue string) string {
	v := q.Get(key)
	if v == "" {
		return defaultValue
	}

	return v
}

// To refactor implements strategy pattern and probably plugins extension
// via named pipe or gRPC servers like TF providers.
func build(u *url.URL) (storage.Engine, error) {
	q := u.Query()

	switch u.Scheme {
	case schemeBundleFromS3:
		// Build session from url
		opts, err := session.FromURL(u.String())
		if err != nil {
			return nil, fmt.Errorf("unable to parse session URL: %w", err)
		}
		// Build AWS session
		sess, err := session.NewSession(opts)
		if err != nil {
			return nil, fmt.Errorf("unable to initialize session: %w", err)
		}
		// Delegate to loader
		return buildWithLoader(u, &s3Loader{
			s3api:      s3.New(sess),
			bucketName: opts.BucketName,
		})
	case schemeBundleFromAzBlob:
		azureConnString := os.Getenv("AZURE_CONNECTION_STRING")
		if azureConnString == "" {
			return nil, errors.New("AZURE_CONNECTION_STRING env. variable must be set for azblob backend")
		}
		return buildWithLoader(u, &azureBlobLoader{
			bucketName: u.Hostname(),
			prefix:     withDefault(q, "prefix", ""),
			connString: azureConnString,
		})
	case schemeBundleFromGCS:
		return buildWithLoader(u, &gcsLoader{
			bucketName: u.Hostname(),
			prefix:     withDefault(q, "prefix", ""),
		})
	case schemeBundleFromHTTP:
		return buildWithLoader(u, &httpLoader{
			scheme: "http",
			host:   u.Host,
		})
	case schemeBundleFromHTTPS:
		return buildWithLoader(u, &httpLoader{
			scheme: "https",
			host:   u.Host,
		})
	case schemeBundleDefault, schemeBundleFromFile:
		fs := afero.NewOsFs()
		fs = afero.NewReadOnlyFs(fs)
		return buildWithLoader(u, &fileLoader{
			fs: fs,
		})
	case schemeBundleStdin:
		return buildWithLoader(u, &stdinLoader{})

	default:
	}

	return nil, fmt.Errorf("unsupported builder scheme (%s)", u.Scheme)
}

func buildWithLoader(u *url.URL, loader Loader) (storage.Engine, error) {
	// Initialize context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Fetch bundle using loader
	br, errDriver := loader.Reader(ctx, u.Path)
	if errDriver != nil {
		return nil, fmt.Errorf("unable to load container content: %w", errDriver)
	}

	// Extract bundle container key form url
	var (
		q              = u.Query()
		containerIDRaw = q.Get("cid")
		unlockKeyRaw   = q.Get("unlock")
	)

	// Initialize bundle
	b, err := getBundle(ctx, br, containerIDRaw, unlockKeyRaw)
	if err != nil {
		return nil, fmt.Errorf("unable to extract bundle: %w", err)
	}

	// Initialize virtual filesystem
	fs, err := vfs.FromBundle(b)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize bundle filesystem: %w", err)
	}

	// Build engine instance
	return &engine{
		u:  u,
		fs: fs,
	}, nil
}

func getBundle(ctx context.Context, br io.Reader, containerID, psk string) (*bundlev1.Bundle, error) {
	// Check container key usage
	var (
		b   *bundlev1.Bundle
		err error
	)

	if containerID != "" {
		// Load container
		sealed, errLoad := container.Load(br)
		if errLoad != nil {
			return nil, fmt.Errorf("unable to load secret container: %w", errLoad)
		}

		// Append given key, and keyring
		containerKeys := append([]string{}, containerID)
		containerKeys = append(containerKeys, containerKeyring...)

		var (
			unsealed  *containerv1.Container
			errUnseal error
		)
		for _, containerKeyRaw := range containerKeys {
			// Decode private key
			containerKey, errDecode := base64.RawURLEncoding.DecodeString(containerKeyRaw)
			if errDecode != nil {
				log.For(ctx).Warn("Invalid key, ignored for encoding error", zap.Error(err))
				continue
			}

			// Unseal container
			unsealed, errUnseal = container.Unseal(sealed, memguard.NewBufferFromBytes(containerKey))
			if errUnseal != nil {
				log.For(ctx).Warn("Unable to unseal container with given key, key is ignored", zap.Error(errUnseal))
				continue
			}

			// Break if container is unsealed
			if unsealed != nil {
				break
			}
		}
		if errUnseal != nil {
			return nil, fmt.Errorf("unable to unseal container: %w", errUnseal)
		}
		if unsealed == nil {
			return nil, fmt.Errorf("unable to unseal container: no key match")
		}

		// Extract bundle
		b, err = bundle.FromContainer(unsealed)
		if err != nil {
			return nil, fmt.Errorf("unable to extract Bundle from sealed container: %w", err)
		}
	} else {
		// No container key assume unsealed container.
		b, err = bundle.FromContainerReader(br)
		if err != nil {
			return nil, fmt.Errorf("unable to extract Bundle from unsealed container: %w", err)
		}
	}

	// Decrypt encrypted bundle using PSK
	if psk != "" {
		// Initialize a transformer from key
		t, err := encryption.FromKey(psk)
		if err != nil {
			return nil, fmt.Errorf("unbale to initialize encryption transformer to unlock bundle: %w", err)
		}

		// Unlock the bundle
		err = bundle.UnLock(ctx, b, t)
		if err != nil {
			return nil, fmt.Errorf("unable to unlock bundle: %w", err)
		}
	}

	// Return result bundle
	return b, nil
}
