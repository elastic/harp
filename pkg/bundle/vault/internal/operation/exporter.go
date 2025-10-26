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
	"sync"
	"sync/atomic"
	"time"

	"github.com/imdario/mergo"
	"go.uber.org/zap"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle/secret"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/sdk/types"
	"github.com/elastic/harp/pkg/vault/kv"
	vaultPath "github.com/elastic/harp/pkg/vault/path"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

const (
	contextErrorNone = "none"
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
	pathCount      int
	pathCountMutex sync.RWMutex
}

// Run the implemented operation
//
//nolint:funlen,gocognit,gocyclo // refactor
func (op *exporter) Run(ctx context.Context) error {
	startTime := time.Now()

	log.For(ctx).Info("Starting vault export operation",
		zap.String("path", op.path),
		zap.Int64("max_workers", op.maxWorkerCount),
		zap.Bool("with_metadata", op.withMetadata),
	)

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

		var processedCount atomic.Int64
		// Listen for message
		for secretPath := range pathChan {
			secPath := secretPath

			if err := gReaderCtx.Err(); err != nil {
				log.For(gReaderCtx).Error("Context error detected, stopping processing",
					zap.Error(err),
					zap.Int64("processed_count", processedCount.Load()),
				)
				break
			}

			log.For(gReaderCtx).Debug("Attempting semaphore acquisition",
				zap.String("secret_path", secretPath),
				zap.Int64("max_workers", op.maxWorkerCount),
				zap.Bool("context_done", gReaderCtx.Done() != nil),
				zap.String("context_error", func() string {
					if err := gReaderCtx.Err(); err != nil {
						return err.Error()
					}
					return contextErrorNone
				}()),
			)

			// Acquire a token
			if err := sem.Acquire(gReaderCtx, 1); err != nil {
				log.For(gReaderCtx).Error("Semaphore acquisition failed",
					zap.Error(err),
					zap.String("secret_path", secretPath),
					zap.String("path", op.path),
					zap.Int64("max_workers", op.maxWorkerCount),
					zap.String("context_error", func() string {
						if ctxErr := gReaderCtx.Err(); ctxErr != nil {
							return ctxErr.Error()
						}
						return contextErrorNone
					}()),
					zap.Bool("is_context_canceled", errors.Is(gReaderCtx.Err(), context.Canceled)),
					zap.Bool("is_deadline_exceeded", errors.Is(gReaderCtx.Err(), context.DeadlineExceeded)),
				)

				return fmt.Errorf("unable to acquire a semaphore token: %w", err)
			}

			log.For(gReaderCtx).Debug("Exporting secret ...", zap.String("path", secretPath))

			// Build function reader
			gReader.Go(func() error {
				// Release token on finish
				defer sem.Release(1)
				defer processedCount.Add(1)

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
				secretData, secretMeta, errRead := op.service.ReadVersion(gReaderCtx, vaultPackagePath, vaultVersion)
				if errRead != nil {
					// Mask path not found or empty secret value
					if errors.Is(errRead, kv.ErrNoData) || errors.Is(errRead, kv.ErrPathNotFound) {
						log.For(gReaderCtx).Debug("No data / path found for given path", zap.String("path", secPath))
						return nil
					}

					// Classify error types for better debugging
					errorType := ClassifyVaultError(errRead)

					log.For(gReaderCtx).Error("Failed to read secret from Vault",
						zap.String("secret_path", vaultPackagePath),
						zap.Uint32("version", vaultVersion),
						zap.String("error_type", errorType),
						zap.Error(errRead),
					)

					return fmt.Errorf("unexpected vault error: %w", errRead)
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
					// Check for old metadata prefix
					if strings.HasPrefix(strings.ToLower(k), legacyBundleMetadataPrefix) {
						metadata[strings.ToLower(k)] = fmt.Sprintf("%s", v)
						// Ignore secret unpacking for this value
						continue
					}

					// Check for new metadata prefix
					if strings.EqualFold(k, kv.VaultMetadataDataKey) {
						if rawMetadata, ok := v.(map[string]interface{}); ok {
							for k, v := range rawMetadata {
								metadata[k] = fmt.Sprintf("%s", v)
							}
						} else {
							log.For(gReaderCtx).Error("Vault metadata type has unexpected type, processing skipped.", zap.String("path", secPath))
						}

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

				// Extract useful metadata
				for k, v := range secretMeta {
					switch k {
					case "version":
						// Convert version
						rawVersion := json.Number(fmt.Sprintf("%s", v))
						version, err := rawVersion.Int64()
						if err != nil {
							log.For(gReaderCtx).Warn("unable to unpack secret version as int64.", zap.Error(err), zap.Any("value", v))
						} else {
							pack.Secrets.Version = uint32(version)
						}
					case "custom_metadata":
						// Check nil
						if types.IsNil(v) {
							continue
						}

						// Copy as metadata
						customMap, ok := v.(map[string]interface{})
						if ok {
							for metaKey, metaValue := range customMap {
								metadata[metaKey] = metaValue.(string)
							}
						} else {
							log.For(gReaderCtx).Warn("unable to unpack secret custom metadata, invalid type.", zap.Any("value", v))
						}
					}
				}

				// Process package metadata distribution
				if op.withMetadata {
					for key, value := range metadata {
						// Merge with package
						switch {
						case strings.HasPrefix(key, "label#"):
							pack.Labels[strings.TrimPrefix(key, "label#")] = value
						case strings.HasPrefix(key, legacyBundleMetadataPrefix):
							// Legacy metadata

							// Clean key
							key = strings.TrimPrefix(key, legacyBundleMetadataPrefix)

							// Unpack value
							var data map[string]string
							if errDecode := json.Unmarshal([]byte(value), &data); errDecode != nil {
								log.For(gReaderCtx).Error("unable to decode package legacy metadata object as JSON", zap.Error(errDecode), zap.String("key", key), zap.String("path", secPath))
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
								log.For(gReaderCtx).Warn("unhandled legacy metadata", zap.String("key", key), zap.String("path", secPath))
								continue
							}

							// Merge with Vault metadata
							if errMergo := mergo.MergeWithOverwrite(meta, data, mergo.WithOverride); errMergo != nil {
								log.For(gReaderCtx).Warn("unable to merge package legacy metadata object", zap.Error(errMergo), zap.String("key", key), zap.String("path", secPath))
								continue
							}
						default:
							pack.Annotations[key] = value
						}
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

		// Wait for all readers to complete
		if err := gReader.Wait(); err != nil {
			return err
		}

		log.For(gctx).Info("Secret reader consumer completed",
			zap.Int64("total_processed", processedCount.Load()),
		)

		return nil
	})

	// Producers ---------------------------------------------------------------

	// Vault crawler
	g.Go(func() error {
		defer close(pathChan)

		log.For(gctx).Info("Starting path producer (vault crawler)",
			zap.String("base_path", op.path),
		)

		producerStartTime := time.Now()

		// Walk the vault path tree
		if err := op.walk(gctx, op.path, op.path, pathChan); err != nil {
			return err
		}

		op.pathCountMutex.RLock()
		totalPaths := op.pathCount
		op.pathCountMutex.RUnlock()

		log.For(gctx).Info("Path producer completed",
			zap.Int("total_paths_published", totalPaths),
			zap.Duration("total_time", time.Since(producerStartTime)),
			zap.Float64("avg_rate_per_sec", float64(totalPaths)/time.Since(producerStartTime).Seconds()),
		)

		return nil
	})

	// Wait for all goroutime to complete
	if err := g.Wait(); err != nil {
		return fmt.Errorf("vault operation error: %w", err)
	}

	log.For(ctx).Info("Vault export operation completed",
		zap.Duration("total_duration", time.Since(startTime)),
		zap.String("path", op.path),
	)

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
			log.For(ctx).Warn("Path producer context canceled",
				zap.Error(ctx.Err()),
				zap.Int("published_count", op.pathCount),
			)
			return ctx.Err()
		case keys <- currPath:
			op.pathCountMutex.Lock()
			op.pathCount++
			currentCount := op.pathCount
			op.pathCountMutex.Unlock()

			// Log progress periodically
			if currentCount%50 == 0 {
				log.For(ctx).Debug("Path producer progress",
					zap.Int("paths_published", currentCount),
					zap.String("latest_path", currPath),
				)
			}
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

func extractVersion(packagePath string) (mountPath string, backendVersion uint32, err error) {
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
