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
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/vault/api"
	"go.uber.org/zap"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle/secret"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/vault/kv"
	vpath "github.com/elastic/harp/pkg/vault/path"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// Importer initialize a secret importer operation
func Importer(client *api.Client, bundleFile *bundlev1.Bundle, prefix string, withMetadata, withVaultMetadata bool, maxWorkerCount int64) Operation {
	return &importer{
		client:            client,
		bundle:            bundleFile,
		prefix:            prefix,
		withMetadata:      withMetadata || withVaultMetadata,
		withVaultMetadata: withVaultMetadata,
		backends:          map[string]kv.Service{},
		maxWorkerCount:    maxWorkerCount,
	}
}

// -----------------------------------------------------------------------------

type importer struct {
	client            *api.Client
	bundle            *bundlev1.Bundle
	prefix            string
	withMetadata      bool
	withVaultMetadata bool
	backends          map[string]kv.Service
	backendsMutex     sync.RWMutex
	maxWorkerCount    int64
}

func (op *importer) logBundleAnalysis(ctx context.Context) {
	var (
		totalSecrets   int
		totalPackages  = len(op.bundle.Packages)
		emptyPackages  int
		largestPackage string
		maxSecrets     int
		backendPaths   = make(map[string]int)
	)

	for _, pkg := range op.bundle.Packages {
		if pkg.Secrets == nil || len(pkg.Secrets.Data) == 0 {
			emptyPackages++
			continue
		}

		secretCount := len(pkg.Secrets.Data)
		totalSecrets += secretCount

		if secretCount > maxSecrets {
			maxSecrets = secretCount
			largestPackage = pkg.Name
		}

		// Count backend distribution
		secretPath := pkg.Name
		if op.prefix != "" {
			secretPath = path.Join(op.prefix, secretPath)
		}
		rootPath := strings.Split(vpath.SanitizePath(secretPath), "/")[0]
		backendPaths[rootPath]++
	}

	log.For(ctx).Info("Bundle analysis",
		zap.Int("total_packages", totalPackages),
		zap.Int("empty_packages", emptyPackages),
		zap.Int("packages_with_secrets", totalPackages-emptyPackages),
		zap.Int("total_secrets", totalSecrets),
		zap.Float64("avg_secrets_per_package", float64(totalSecrets)/float64(totalPackages-emptyPackages)),
		zap.String("largest_package", largestPackage),
		zap.Int("max_secrets_in_package", maxSecrets),
		zap.Int("unique_backends", len(backendPaths)),
		zap.Any("backend_distribution", backendPaths),
	)
}

// Run the implemented operation
//
//nolint:gocognit,funlen,gocyclo // To refactor
func (op *importer) Run(ctx context.Context) error {
	startTime := time.Now()

	log.For(ctx).Info("Starting vault import operation",
		zap.String("prefix", op.prefix),
		zap.Int("total_packages", len(op.bundle.Packages)),
		zap.Int64("max_workers", op.maxWorkerCount),
		zap.Bool("with_metadata", op.withMetadata),
		zap.Bool("with_vault_metadata", op.withVaultMetadata),
	)
	// Initialize sub context
	g, gctx := errgroup.WithContext(ctx)

	op.logBundleAnalysis(gctx)

	// Prepare channels
	packageChan := make(chan *bundlev1.Package)

	// Validate worker count
	if op.maxWorkerCount < 1 {
		op.maxWorkerCount = 1
	}

	// consumers ---------------------------------------------------------------

	// Secret writer
	g.Go(func() error {
		// Initialize a semaphore with maxReaderWorker tokens
		sem := semaphore.NewWeighted(op.maxWorkerCount)

		// Writer errGroup
		gWriter, gWriterCtx := errgroup.WithContext(gctx)

		processedCount := 0
		// Listen for message
		for secretPackage := range packageChan {
			if err := gWriterCtx.Err(); err != nil {
				log.For(gWriterCtx).Error("Context error detected, stopping processing",
					zap.Error(err),
					zap.Int("processed_count", processedCount),
				)
				break
			}

			log.For(gWriterCtx).Debug("Attempting semaphore acquisition",
				zap.String("package_path", secretPackage.Name),
				zap.Int64("max_workers", op.maxWorkerCount),
				zap.Bool("context_done", gWriterCtx.Done() != nil),
				zap.String("context_error", func() string {
					if err := gWriterCtx.Err(); err != nil {
						return err.Error()
					}
					return "none"
				}()),
			)

			// Acquire a token
			if err := sem.Acquire(gWriterCtx, 1); err != nil {
				log.For(gWriterCtx).Error("Semaphore acquisition failed",
					zap.Error(err),
					zap.String("package_path", secretPackage.Name),
					zap.String("prefix", op.prefix),
					zap.Int64("max_workers", op.maxWorkerCount),
					zap.String("context_error", func() string {
						if ctxErr := gWriterCtx.Err(); ctxErr != nil {
							return ctxErr.Error()
						}
						return "none"
					}()),
					zap.Bool("is_context_canceled", gWriterCtx.Err() == context.Canceled),
					zap.Bool("is_deadline_exceeded", gWriterCtx.Err() == context.DeadlineExceeded),
				)

				return fmt.Errorf("unable to acquire a semaphore token: %w", err)
			}

			log.For(gWriterCtx).Debug("Writing secret ...", zap.String("prefix", op.prefix), zap.String("path", secretPackage.Name))

			// Build function reader
			gWriter.Go(func() error {
				defer sem.Release(1)

				if err := gWriterCtx.Err(); err != nil {
					// Context has already an error
					return nil
				}

				// No data to insert
				if secretPackage.Secrets == nil {
					return nil
				}

				data := map[string]interface{}{}
				// Wrap secret k/v as a map
				for _, s := range secretPackage.Secrets.Data {
					// Unpack secret to original value
					var value interface{}
					if err := secret.Unpack(s.Value, &value); err != nil {
						return fmt.Errorf("unable to unpack secret value for path '%s' with key '%s': %w", secretPackage.Name, s.Key, err)
					}

					// Assign to map for vault storage
					data[s.Key] = value
				}

				// Export metadata
				metadata := map[string]interface{}{}
				if op.withMetadata {
					// Has annotations
					if len(secretPackage.Annotations) > 0 {
						for k, v := range secretPackage.Annotations {
							metadata[k] = v
						}
					}

					// Has labels
					if len(secretPackage.Labels) > 0 {
						for k, v := range secretPackage.Labels {
							metadata[fmt.Sprintf("label#%s", k)] = v
						}
					}
				}

				// Assemble secret path
				secretPath := secretPackage.Name
				if op.prefix != "" {
					secretPath = path.Join(op.prefix, secretPath)
				}

				// Extract root backend path
				rootPath := strings.Split(vpath.SanitizePath(secretPath), "/")[0]

				// Check backend initialization
				if _, ok := op.backends[rootPath]; !ok {
					// Initialize new service for backend
					log.For(gWriterCtx).Info("Initializing new KV backend",
						zap.String("root_path", rootPath),
						zap.String("full_secret_path", secretPath),
						zap.Bool("with_vault_metadata", op.withVaultMetadata),
					)
					service, err := kv.New(op.client, rootPath, kv.WithVaultMetatadata(op.withVaultMetadata))
					if err != nil {
						log.For(gWriterCtx).Error("Failed to initialize KV backend",
							zap.String("root_path", rootPath),
							zap.Error(err),
						)
						return fmt.Errorf("unable to initialize Vault service for '%s' KV backend: %w", op.prefix, err)
					}

					log.For(gWriterCtx).Debug("Successfully initialized KV backend",
						zap.String("root_path", rootPath),
						zap.Int("total_backends", len(op.backends)+1),
					)

					// All queries will be handled by same backend service
					op.backendsMutex.Lock()
					op.backends[rootPath] = service
					op.backendsMutex.Unlock()
				}

				// Write secret to Vault
				if err := op.backends[rootPath].WriteWithMeta(gWriterCtx, secretPath, data, metadata); err != nil {
					// Classify error types for better debugging
					errorType := "unknown"
					if strings.Contains(err.Error(), "connection") {
						errorType = "connection"
					} else if strings.Contains(err.Error(), "permission") {
						errorType = "permission"
					} else if strings.Contains(err.Error(), "timeout") {
						errorType = "timeout"
					}

					log.For(gWriterCtx).Error("Failed to write secret to Vault",
						zap.String("secret_path", secretPath),
						zap.String("root_path", rootPath),
						zap.String("error_type", errorType),
						zap.Int("secret_count", len(data)),
						zap.Int("metadata_count", len(metadata)),
						zap.Error(err),
					)
					return fmt.Errorf("unable to write secret data for path '%s': %w", secretPath, err)
				}

				// No error
				return nil
			})
		}

		// No error
		return gWriter.Wait()
	})

	// producers ---------------------------------------------------------------

	// Bundle package publisher
	g.Go(func() error {
		defer close(packageChan)

		log.For(gctx).Info("Starting package producer",
			zap.Int("total_packages", len(op.bundle.Packages)),
		)

		published := 0
		startTime := time.Now()

		for _, p := range op.bundle.Packages {
			select {
			case <-gctx.Done():
				log.For(gctx).Warn("Producer context canceled",
					zap.Error(gctx.Err()),
					zap.Int("published_count", published),
					zap.Int("remaining_packages", len(op.bundle.Packages)-published),
				)
				return gctx.Err()
			case packageChan <- p:
				published++

				// Log every 50 packages or at specific percentiles
				if published%50 == 0 || published == len(op.bundle.Packages)/4 ||
					published == len(op.bundle.Packages)/2 || published == len(op.bundle.Packages) {
					log.For(gctx).Debug("Producer progress",
						zap.Int("published", published),
						zap.Int("total", len(op.bundle.Packages)),
						zap.Float64("rate_per_sec", float64(published)/time.Since(startTime).Seconds()),
					)
				}
			}
		}

		log.For(gctx).Info("Package producer completed",
			zap.Int("total_published", published),
			zap.Duration("total_time", time.Since(startTime)),
			zap.Float64("avg_rate_per_sec", float64(published)/time.Since(startTime).Seconds()),
		)

		// No error
		return nil
	})

	// Wait for all goroutime to complete
	if err := g.Wait(); err != nil {
		return fmt.Errorf("vault operation error: %w", err)
	}

	log.For(ctx).Info("Vault import operation completed",
		zap.Duration("total_duration", time.Since(startTime)),
		zap.String("prefix", op.prefix),
		zap.Int("total_packages", len(op.bundle.Packages)),
	)
	// No error
	return nil
}
