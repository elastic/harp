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

package operation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/imdario/mergo"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle/secret"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/vault/kv"
	vaultPath "github.com/elastic/harp/pkg/vault/path"
)

// Exporter initialize a secret exporter operation
func Exporter(service kv.Service, backendPath string, output chan *bundlev1.Package, withMetadata bool, maxWorkerCount int64) Operation {
	return &exporter{
		service:        service,
		path:           backendPath,
		withMetadata:   withMetadata,
		output:         output,
		maxWorkerCount: maxWorkerCount,
	}
}

// -----------------------------------------------------------------------------

type exporter struct {
	service        kv.Service
	path           string
	withMetadata   bool
	output         chan *bundlev1.Package
	maxWorkerCount int64
}

// Run the implemented operation
//nolint:funlen,gocognit,gocyclo // refactor
func (op *exporter) Run(ctx context.Context) error {
	// Initialize sub context
	g, gctx := errgroup.WithContext(ctx)

	// Prepare channels
	pathChan := make(chan string)

	// Validate worker count
	if op.maxWorkerCount < 1 {
		op.maxWorkerCount = 1
	}

	// Consumers ---------------------------------------------------------------

	// Secret reader
	g.Go(func() error {
		// Initialize a semaphore with maxReaderWorker tokens
		sem := semaphore.NewWeighted(op.maxWorkerCount)

		// Reader errGroup
		gReader, gReaderCtx := errgroup.WithContext(gctx)

		// Listen for message
		for secretPath := range pathChan {
			secPath := secretPath

			if err := gReaderCtx.Err(); err != nil {
				// Stop processing
				break
			}

			// Acquire a token
			if err := sem.Acquire(gReaderCtx, 1); err != nil {
				return fmt.Errorf("unable to acquire a semaphore token: %w", err)
			}

			log.For(gReaderCtx).Debug("Exporting secret ...", zap.String("path", secretPath))

			// Build function reader
			gReader.Go(func() error {
				// Release token on finish
				defer sem.Release(1)

				if err := gReaderCtx.Err(); err != nil {
					// Context has already an error
					return nil
				}

				// Extract desired version from path
				vaultPackagePath, vaultVersion, errPackagePath := extractVersion(secPath)
				if errPackagePath != nil {
					return fmt.Errorf("unable to parse package path '%s': %w", secPath, errPackagePath)
				}

				// Read from Vault
				secretData, secretMeta, err := op.service.ReadVersion(gReaderCtx, vaultPackagePath, vaultVersion)
				if err != nil {
					// Mask path not found or empty secret value
					if errors.Is(err, kv.ErrNoData) || errors.Is(err, kv.ErrPathNotFound) {
						log.For(gReaderCtx).Debug("No data / path found for given path", zap.String("path", secPath))
						return nil
					}
					return fmt.Errorf("unexpected vault error: %w", err)
				}

				// Prepare secret list
				chain := &bundlev1.SecretChain{
					Version:         uint32(0),
					Data:            make([]*bundlev1.KV, 0),
					NextVersion:     nil,
					PreviousVersion: nil,
				}

				// Prepare metadata holder
				metadata := map[string]string{}

				// Iterate over secret bundle
				for k, v := range secretData {
					// Check for metadata prefix
					if strings.HasPrefix(strings.ToLower(k), bundleMetadataPrefix) {
						metadata[strings.ToLower(k)] = fmt.Sprintf("%s", v)

						// Ignore secret unpacking for this value
						continue
					}

					// Pack secret value
					s, errPack := op.packSecret(k, v)
					if errPack != nil {
						return fmt.Errorf("unable to pack secret value for path '%s' with key '%s' : %w", secPath, k, errPack)
					}

					// Add secret to package
					chain.Data = append(chain.Data, s)
				}

				// Prepare the secret package
				pack := &bundlev1.Package{
					Labels:      map[string]string{},
					Annotations: map[string]string{},
					Name:        vaultPackagePath,
					Secrets:     chain,
				}

				// Process package metadata
				if op.withMetadata {
					for key, value := range metadata {
						// Clean key
						key = strings.TrimPrefix(key, bundleMetadataPrefix)

						// Unpack value
						var data map[string]string
						if errDecode := json.Unmarshal([]byte(value), &data); errDecode != nil {
							log.For(gReaderCtx).Error("unable to decode package metadata object as JSON", zap.Error(errDecode), zap.String("key", key), zap.String("path", secPath))
							continue
						}

						var meta interface{}

						// Merge with package
						switch key {
						case "#annotations":
							meta = &pack.Annotations
						case "#labels":
							meta = &pack.Labels
						default:
							log.For(gReaderCtx).Warn("unhandled metadata", zap.Error(err), zap.String("key", key), zap.String("path", secPath))
							continue
						}

						// Merge with Vault metadata
						if errMergo := mergo.MergeWithOverwrite(meta, data, mergo.WithOverride); errMergo != nil {
							log.For(gReaderCtx).Warn("unable to merge package metadata object", zap.Error(errMergo), zap.String("key", key), zap.String("path", secPath))
							continue
						}
					}
				}

				// Embed secret storage metadata
				if secretMeta != nil {
					// Encode all Vault metadata as json
					metadataJSON, errMetaJSON := json.Marshal(secretMeta)
					if errMetaJSON != nil {
						return fmt.Errorf("unable to encode Vault metadata for path '%s': %w", secPath, errMetaJSON)
					}
					pack.Annotations[vaultKVMetadata] = string(metadataJSON)

					// Extract useful metadata
					for k, v := range secretMeta {
						switch k {
						case "version":
							pack.Annotations[vaultKVv2MetadataVersion] = fmt.Sprintf("%s", v)
						case "created_time":
							pack.Annotations[vaultKVv2MetadataCreatedTime] = fmt.Sprintf("%s", v)
						}
					}
				}

				// Dispatch annotations to package
				if v, ok := pack.Annotations[vaultKVv2MetadataVersion]; ok {
					// Convert version
					secretVersion, errParse := strconv.ParseUint(v, 10, 32)
					if errParse != nil {
						log.For(ctx).Warn("unable to parse secret data version as a valid number: %w", zap.Error(errParse))
					} else {
						pack.Secrets.Version = uint32(secretVersion)
					}
				}

				// Publish secret package
				select {
				case <-gReaderCtx.Done():
					return gReaderCtx.Err()
				case op.output <- pack:
					return nil
				}
			})
		}

		return gReader.Wait()
	})

	// Producers ---------------------------------------------------------------

	// Vault crawler
	g.Go(func() error {
		defer close(pathChan)
		return op.walk(gctx, op.path, op.path, pathChan)
	})

	// Wait for all goroutime to complete
	if err := g.Wait(); err != nil {
		return fmt.Errorf("vault operation error: %w", err)
	}

	// No error
	return nil
}

// -----------------------------------------------------------------------------

func (op *exporter) walk(ctx context.Context, basePath, currPath string, keys chan string) error {
	// List secret of basepath
	res, err := op.service.List(ctx, basePath)
	if err != nil {
		return fmt.Errorf("unable to list secret entries for '%s': %w", basePath, err)
	}

	// Check path is a leaf
	if res == nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case keys <- currPath:
		}
		return nil
	}

	// Iterate on all subpath
	for _, p := range res {
		if err := op.walk(ctx, path.Join(basePath, p), path.Join(currPath, p), keys); err != nil {
			return fmt.Errorf("unable to walk '%s' : %w", path.Join(basePath, p), err)
		}
	}

	// No error
	return nil
}

func (op *exporter) packSecret(key string, value interface{}) (*bundlev1.KV, error) {
	// Pack secret value
	payload, err := secret.Pack(value)
	if err != nil {
		return nil, fmt.Errorf("unable to pack secret '%s': %w", key, err)
	}

	// Build the secret object
	return &bundlev1.KV{
		Key:   key,
		Type:  fmt.Sprintf("%T", value),
		Value: payload,
	}, nil
}

func extractVersion(packagePath string) (string, uint32, error) {
	// Check arguments
	if packagePath == "" {
		return "", 0, fmt.Errorf("unable to extract path and version from an empty string")
	}

	// Looks a little hack-ish for me
	u, err := url.ParseRequestURI(fmt.Sprintf("harp://bundle/%s", packagePath))
	if err != nil {
		return "", 0, fmt.Errorf("unable to parse package path: %w", err)
	}

	// Get version
	versionRaw := u.Query().Get("version")
	if versionRaw == "" {
		// Get latest
		return vaultPath.SanitizePath(u.Path), 0, nil
	}

	// Convert
	versionUnit, errParse := strconv.ParseUint(versionRaw, 10, 32)
	if errParse != nil {
		return "", 0, fmt.Errorf("unable to parse version as a valid integer: %w", err)
	}

	// Return path elements
	return vaultPath.SanitizePath(u.Path), uint32(versionUnit), nil
}
