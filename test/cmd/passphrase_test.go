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
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Harp CLI", func() {
	Describe("passphrase", func() {
		var harpCmd *exec.Cmd

		Context("when no arguments specified", func() {
			BeforeEach(func() {
				harpCmd = exec.Command(harpPath, "passphrase")
			})

			It("exits with status code 0", func() {
				sess, err := gexec.Start(harpCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(gexec.Exit(0))
			})
		})

		Context("when specifying word count", func() {
			BeforeEach(func() {
				harpCmd = exec.Command(harpPath, "passphrase", "--word-count", "8")
			})

			It("exits with status code 0", func() {
				sess, err := gexec.Start(harpCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(gexec.Exit(0))
			})
		})

		Context("when specifying word count using alias", func() {
			BeforeEach(func() {
				harpCmd = exec.Command(harpPath, "passphrase", "-w", "8")
			})

			It("exits with status code 0", func() {
				sess, err := gexec.Start(harpCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(gexec.Exit(0))
			})
		})
	})
})
