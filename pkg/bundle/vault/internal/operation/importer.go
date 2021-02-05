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
	"fmt"
	"strings"
	"sync"

	"github.com/hashicorp/vault/api"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle/secret"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/vault/kv"
	vpath "github.com/elastic/harp/pkg/vault/path"
)

// Importer initialize a secret importer operation
func Importer(client *api.Client, bundleFile *bundlev1.Bundle, prefix string, withMetadata bool, maxWorkerCount int64) Operation {
	return &importer{
		client:         client,
		bundle:         bundleFile,
		prefix:         prefix,
		withMetadata:   withMetadata,
		backends:       map[string]kv.Service{},
		maxWorkerCount: maxWorkerCount,
	}
}

// -----------------------------------------------------------------------------

type importer struct {
	client         *api.Client
	bundle         *bundlev1.Bundle
	prefix         string
	withMetadata   bool
	backends       map[string]kv.Service
	backendsMutex  sync.RWMutex
	maxWorkerCount int64
}

// Run the implemented operation
// nolint:gocognit,funlen,gocyclo // To refactor
func (op *importer) Run(ctx context.Context) error {
	// Initialize sub context
	g, gctx := errgroup.WithContext(ctx)

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

		// Listen for message
		for secretPackage := range packageChan {
			// Assign local reference
			secretPackage := secretPackage

			if err := gWriterCtx.Err(); err != nil {
				// Stop processing
				break
			}

			// Acquire a token
			if err := sem.Acquire(gWriterCtx, 1); err != nil {
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
				if op.withMetadata {
					// Has annotations
					if len(secretPackage.Annotations) > 0 {
						out, err := json.Marshal(secretPackage.Annotations)
						if err != nil {
							return fmt.Errorf("unable to encode annotations as JSON for path '%v': %w", secretPackage.Name, err)
						}

						// Assign json
						data["harp.elastic.io/v1/bundle#annotations"] = string(out)
					}

					// Has labels
					if len(secretPackage.Labels) > 0 {
						out, err := json.Marshal(secretPackage.Labels)
						if err != nil {
							return fmt.Errorf("unable to encode labels as JSON for path '%v': %w", secretPackage.Name, err)
						}

						// Assign json
						data["harp.elastic.io/v1/bundle#labels"] = string(out)
					}
				}

				// Assemble secret path
				secretPath := secretPackage.Name
				if op.prefix != "" {
					secretPath = fmt.Sprintf("%s/%s", op.prefix, secretPath)
				}

				// Extract root backend path
				rootPath := strings.Split(vpath.SanitizePath(secretPath), "/")[0]

				// Check backend initialization
				if _, ok := op.backends[rootPath]; !ok {
					// Initialize new service for backend
					service, err := kv.New(op.client, rootPath)
					if err != nil {
						return fmt.Errorf("unable to initialize Vault service for '%s' KV backend: %w", op.prefix, err)
					}

					// All queries will be handled by same backend service
					op.backendsMutex.Lock()
					op.backends[rootPath] = service
					op.backendsMutex.Unlock()
				}

				// Write secret to Vault
				if err := op.backends[rootPath].Write(gWriterCtx, secretPath, data); err != nil {
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

		for _, p := range op.bundle.Packages {
			select {
			case <-gctx.Done():
				return gctx.Err()
			case packageChan <- p:
			}
		}

		// No error
		return nil
	})

	// Wait for all goroutime to complete
	if err := g.Wait(); err != nil {
		return fmt.Errorf("vault operation error: %w", err)
	}

	// No error
	return nil
}
