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

package vfs

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/awnumar/memguard"
	"github.com/spf13/afero"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle/secret"
)

// FromBundle initialize an afero filesystem abstraction from a secret bundle
func FromBundle(b *bundlev1.Bundle) (afero.Fs, error) {
	fsMap := map[string]*memguard.Enclave{}

	// Prepare filesystem
	for _, p := range b.Packages {
		var box *memguard.Enclave

		// Skip when no secret chain is defined
		if p.Secrets == nil {
			continue
		}

		// Check if package has locked secret
		if p.Secrets.Locked != nil {
			box = memguard.NewEnclave(p.Secrets.Locked.Value)
			// Clear buffer
			memguard.WipeBytes(p.Secrets.Locked.Value)
		} else {
			// Convert secret as a map
			secrets := map[string]interface{}{}

			for _, s := range p.Secrets.Data {
				var out interface{}
				if err := secret.Unpack(s.Value, &out); err != nil {
					return nil, fmt.Errorf("unable to load secret value, corrupted bundle")
				}

				// Assign to secret map
				secrets[s.Key] = out
			}

			// Check if secret is a file
			if value, ok := secrets["@content"]; ok {
				content := value.([]byte)

				// Assign file content as value
				box = memguard.NewEnclave(content)

				// Clear buffer
				memguard.WipeBytes(content)
			} else {
				// Convert as json
				content, err := json.Marshal(secrets)
				if err != nil {
					return nil, fmt.Errorf("unable to extract secret map as json")
				}

				// Lock buffer
				box = memguard.NewEnclave(content)

				// Clear buffer
				memguard.WipeBytes(content)
			}
		}

		// Add to map
		fsMap[fmt.Sprintf("/%s", p.Name)] = box
	}

	return &bundleFs{
		bundle: b,
		files:  fsMap,
	}, nil
}

// -----------------------------------------------------------------------------

var (
	// ErrReadOnly is raised when calling fs modification operation.
	ErrReadOnly = fmt.Errorf("filesystem is readonly")
	// ErrDirectoryListingIsForbidden is raise dwhen trying to list files from a directory
	ErrDirectoryListingIsForbidden = fmt.Errorf("directory listing is forbidden")
	// ErrNotSeekable is raised when trying to do partial read
	ErrNotSeekable = fmt.Errorf("filesystem is not seekable")
)

// -----------------------------------------------------------------------------

type bundleFs struct {
	bundle *bundlev1.Bundle
	files  map[string]*memguard.Enclave
}

// The name of this FileSystem
func (fs *bundleFs) Name() string {
	return "SecretBundleFS"
}

// Open opens a file, returning it or an error, if any happens.
func (fs *bundleFs) Open(name string) (afero.File, error) {
	payload, ok := fs.files[name]
	if !ok {
		return nil, afero.ErrFileNotFound
	}

	// Open the enclave
	lockedBuffer, err := payload.Open()
	if err != nil {
		return nil, fmt.Errorf("unable to open enclave for path '%s': %w", name, err)
	}

	// Wrap with a file
	return &secretFile{
		name:    name,
		content: lockedBuffer.Reader(),
		size:    int64(lockedBuffer.Size()),
	}, nil
}

// Stat returns a FileInfo describing the named file, or an error, if any
// happens.
func (fs *bundleFs) Stat(name string) (os.FileInfo, error) {
	payload, ok := fs.files[name]
	if !ok {
		return nil, afero.ErrFileNotFound
	}

	return &secretFileInfo{
		name: name,
		size: int64(payload.Size()),
	}, nil
}

// -----------------------------------------------------------------------------

// Create creates a file in the filesystem, returning the file and an
// error, if any happens.
func (fs *bundleFs) Create(name string) (afero.File, error) {
	return nil, ErrReadOnly
}

// Mkdir creates a directory in the filesystem, return an error if any
// happens.
func (fs *bundleFs) Mkdir(name string, perm os.FileMode) error {
	return ErrReadOnly
}

// MkdirAll creates a directory path and all parents that does not exist
// yet.
func (fs *bundleFs) MkdirAll(path string, perm os.FileMode) error {
	return ErrReadOnly
}

// OpenFile opens a file using the given flags and the given mode.
func (fs *bundleFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	return nil, ErrReadOnly
}

// Remove removes a file identified by name, returning an error, if any
// happens.
func (fs *bundleFs) Remove(name string) error {
	return ErrReadOnly
}

// RemoveAll removes a directory path and any children it contains. It
// does not fail if the path does not exist (return nil).
func (fs *bundleFs) RemoveAll(path string) error {
	return ErrReadOnly
}

// Rename renames a file.
func (fs *bundleFs) Rename(oldname, newname string) error {
	return ErrReadOnly
}

// Chmod changes the mode of the named file to mode.
func (fs *bundleFs) Chmod(name string, mode os.FileMode) error {
	return ErrReadOnly
}

// Chown changes the uid and gid of the named file.
func (fs *bundleFs) Chown(name string, uid, gid int) error {
	return ErrReadOnly
}

// Chtimes changes the access and modification times of the named file
func (fs *bundleFs) Chtimes(name string, atime, mtime time.Time) error {
	return ErrReadOnly
}
