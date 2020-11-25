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

package cmd_test

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var (
	harpPath string
	tmpDir   string
)

// -----------------------------------------------------------------------------

var _ = SynchronizedBeforeSuite(func() []byte {
	binPath, err := gexec.Build("github.com/elastic/harp/cmd/harp")
	Expect(err).NotTo(HaveOccurred())

	return []byte(binPath)
}, func(data []byte) {
	harpPath = string(data)

	SetDefaultEventuallyTimeout(10 * time.Second)
})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	gexec.CleanupBuildArtifacts()
})

var _ = BeforeEach(func() {
	var err error

	tmpDir, err = ioutil.TempDir("", "harp-test")
	Expect(err).NotTo(HaveOccurred())

	os.Setenv("HOME", tmpDir)
})

var _ = AfterEach(func() {
	os.RemoveAll(tmpDir)
})

// -----------------------------------------------------------------------------

func harpCommand(dir string, args ...string) *exec.Cmd {
	cmd := exec.Command(harpPath, args...)
	cmd.Dir = dir

	return cmd
}

func startHarp(dir string, args ...string) *gexec.Session {
	cmd := harpCommand(dir, args...)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	return session
}

func tmpPath(destination string) string {
	return filepath.Join(tmpDir, destination)
}

func fixturePath(name string) string {
	return filepath.Join("../fixtures", name)
}

func fromFile(basePath, filename string) *os.File {
	f, err := os.Open(filepath.Join(basePath, filename))
	Expect(err).ShouldNot(HaveOccurred())
	return f
}

func copyIn(sourcePath, destinationPath string, recursive bool) {
	err := os.MkdirAll(destinationPath, 0o777)
	Expect(err).NotTo(HaveOccurred())

	files, err := ioutil.ReadDir(sourcePath)
	Expect(err).NotTo(HaveOccurred())
	for _, f := range files {
		srcPath := filepath.Join(sourcePath, f.Name())
		dstPath := filepath.Join(destinationPath, f.Name())
		if f.IsDir() {
			if recursive {
				copyIn(srcPath, dstPath, recursive)
			}
			continue
		}

		src, err := os.Open(srcPath)

		Expect(err).NotTo(HaveOccurred())
		defer src.Close()

		dst, err := os.Create(dstPath)
		Expect(err).NotTo(HaveOccurred())
		defer dst.Close()

		_, err = io.Copy(dst, src)
		Expect(err).NotTo(HaveOccurred())
	}
}

func sameFile(filePath, otherFilePath string) bool {
	content, readErr := ioutil.ReadFile(filePath)
	Expect(readErr).NotTo(HaveOccurred())
	otherContent, readErr := ioutil.ReadFile(otherFilePath)
	Expect(readErr).NotTo(HaveOccurred())
	Expect(string(content)).To(Equal(string(otherContent)))
	return true
}

// -----------------------------------------------------------------------------

func TestIntegration(t *testing.T) {
	SetDefaultEventuallyTimeout(30 * time.Second)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}
