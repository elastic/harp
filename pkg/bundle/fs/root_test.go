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

// +build go1.16

package fs

import (
	"io/fs"
	"testing"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	fuzz "github.com/google/gofuzz"
)

func TestFromBundle(t *testing.T) {
	type args struct {
		paths []string
		b     *bundlev1.Bundle
	}
	tests := []struct {
		name    string
		args    args
		want    BundleFS
		wantErr bool
	}{
		{
			name:    "nil",
			wantErr: true,
		},
		{
			name: "empty",
			args: args{
				b: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{},
				},
			},
			wantErr: false,
		},
		{
			name: "valid",
			args: args{
				b: &bundlev1.Bundle{
					Packages: []*bundlev1.Package{
						{
							Name: "application/test",
						},
						{
							Name: "application/production/test",
						},
						{
							Name: "application/staging/test",
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FromBundle(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromBundle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestFromBundle_Fuzz(t *testing.T) {
	// Making sure the descrption never panics
	for i := 0; i < 50; i++ {
		f := fuzz.New()

		var (
			src bundlev1.Bundle
		)

		// Prepare arguments
		f.Fuzz(&src)

		// Execute
		FromBundle(&src)
	}
}

func mustFromBundle(b *bundlev1.Bundle) BundleFS {
	fs, err := FromBundle(b)
	if err != nil {
		panic(err)
	}
	return fs
}

var testBundle = &bundlev1.Bundle{
	Packages: []*bundlev1.Package{
		{
			Name: "application/test",
		},
		{
			Name: "application/production/test",
		},
		{
			Name: "application/staging/test",
		},
	},
}

func Test_bundleFs_Open(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fs      BundleFS
		args    args
		want    fs.File
		wantErr bool
	}{
		{
			name: "empty",
			fs:   mustFromBundle(testBundle),
			args: args{
				name: "",
			},
		},
		{
			name: "directory",
			fs:   mustFromBundle(testBundle),
			args: args{
				name: "application/production",
			},
		},
		{
			name: "directory not exists",
			fs:   mustFromBundle(testBundle),
			args: args{
				name: "application/whatever",
			},
			wantErr: true,
		},
		{
			name: "file",
			fs:   mustFromBundle(testBundle),
			args: args{
				name: "application/production/test",
			},
		},
		{
			name: "file not exists",
			fs:   mustFromBundle(testBundle),
			args: args{
				name: "application/production/whatever",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.fs.Open(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("bundleFs.Open() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_bundleFs_Open_Fuzz(t *testing.T) {
	bfs, _ := FromBundle(testBundle)

	// Making sure the descrption never panics
	for i := 0; i < 50; i++ {
		f := fuzz.New()

		var (
			name string
		)

		// Prepare arguments
		f.Fuzz(&name)

		// Execute
		bfs.Open(name)
	}
}

func Test_bundleFs_ReadDir(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fs      BundleFS
		args    args
		want    []fs.DirEntry
		wantErr bool
	}{
		{
			name: "empty",
			fs:   mustFromBundle(testBundle),
			args: args{
				name: "",
			},
		},
		{
			name: "directory",
			fs:   mustFromBundle(testBundle),
			args: args{
				name: "application/production",
			},
		},
		{
			name: "directory not exists",
			fs:   mustFromBundle(testBundle),
			args: args{
				name: "application/whatever",
			},
			wantErr: true,
		},
		{
			name: "file",
			fs:   mustFromBundle(testBundle),
			args: args{
				name: "application/production/test",
			},
			wantErr: true,
		},
		{
			name: "file not exists",
			fs:   mustFromBundle(testBundle),
			args: args{
				name: "application/production/whatever",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.fs.ReadDir(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("bundleFs.ReadDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_bundleFs_ReadDir_Fuzz(t *testing.T) {
	bfs, _ := FromBundle(testBundle)

	// Making sure the descrption never panics
	for i := 0; i < 50; i++ {
		f := fuzz.New()

		var (
			name string
		)

		// Prepare arguments
		f.Fuzz(&name)

		// Execute
		bfs.ReadDir(name)
	}
}

func Test_bundleFs_ReadFile(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fs      BundleFS
		args    args
		want    []fs.DirEntry
		wantErr bool
	}{
		{
			name: "empty",
			fs:   mustFromBundle(testBundle),
			args: args{
				name: "",
			},
			wantErr: true,
		},
		{
			name: "directory",
			fs:   mustFromBundle(testBundle),
			args: args{
				name: "application/production",
			},
			wantErr: true,
		},
		{
			name: "directory not exists",
			fs:   mustFromBundle(testBundle),
			args: args{
				name: "application/whatever",
			},
			wantErr: true,
		},
		{
			name: "file",
			fs:   mustFromBundle(testBundle),
			args: args{
				name: "application/production/test",
			},
		},
		{
			name: "file not exists",
			fs:   mustFromBundle(testBundle),
			args: args{
				name: "application/production/whatever",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.fs.ReadFile(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("bundleFs.ReadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_bundleFs_ReadFile_Fuzz(t *testing.T) {
	bfs, _ := FromBundle(testBundle)

	// Making sure the descrption never panics
	for i := 0; i < 50; i++ {
		f := fuzz.New()

		var (
			name string
		)

		// Prepare arguments
		f.Fuzz(&name)

		// Execute
		bfs.ReadFile(name)
	}
}

func Test_bundleFs_Stat(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fs      BundleFS
		args    args
		want    []fs.DirEntry
		wantErr bool
	}{
		{
			name: "empty",
			fs:   mustFromBundle(testBundle),
			args: args{
				name: "",
			},
		},
		{
			name: "directory",
			fs:   mustFromBundle(testBundle),
			args: args{
				name: "application/production",
			},
		},
		{
			name: "directory not exists",
			fs:   mustFromBundle(testBundle),
			args: args{
				name: "application/whatever",
			},
			wantErr: true,
		},
		{
			name: "file",
			fs:   mustFromBundle(testBundle),
			args: args{
				name: "application/production/test",
			},
		},
		{
			name: "file not exists",
			fs:   mustFromBundle(testBundle),
			args: args{
				name: "application/production/whatever",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.fs.Stat(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("bundleFs.Stat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_bundleFs_Stat_Fuzz(t *testing.T) {
	bfs, _ := FromBundle(testBundle)

	// Making sure the descrption never panics
	for i := 0; i < 50; i++ {
		f := fuzz.New()

		var (
			name string
		)

		// Prepare arguments
		f.Fuzz(&name)

		// Execute
		bfs.Stat(name)
	}
}
