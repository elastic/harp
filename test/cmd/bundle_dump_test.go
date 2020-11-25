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
	"encoding/json"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Harp CLI", func() {
	Describe("bundle dump", func() {
		var (
			cmdParams []string
			pkgPath   string
			session   *gexec.Session
		)

		BeforeEach(func() {
			pkgPath = tmpPath("bundle-dump")
			os.Mkdir(pkgPath, 0o777)
		})

		JustBeforeEach(func() {
			session = startHarp(pkgPath, cmdParams...)
		})

		Context("when dumping a bundle", func() {
			Context("when no arguments specified", func() {
				BeforeEach(func() {
					cmdParams = []string{"bundle", "encrypt"}
				})

				It("exits with status code 0", func() {
					Eventually(session).Should(gexec.Exit(0))
				})

				It("should emit required content", func() {
					Eventually(session.Err).Should(gbytes.Say(`Error: required flag\(s\) "key" not set`))
				})
			})

			Context("with input bundle specified", func() {
				BeforeEach(func() {
					copyIn(fixturePath("bundles"), pkgPath, false)
				})

				Context("when input not exists", func() {
					BeforeEach(func() {
						cmdParams = []string{"bundle", "dump", "--in", "not-existent-bundle.bundle"}
					})

					It("exits with status code 0", func() {
						Eventually(session).Should(gexec.Exit(1))
					})

					It("should emit required content", func() {
						Eventually(session.Err).Should(gbytes.Say("fatal"))
						Eventually(session.Err).Should(gbytes.Say(`unable to open 'not-existent-bundle.bundle' for read`))
					})
				})

				Context("use a file as input", func() {
					BeforeEach(func() {
						cmdParams = []string{"bundle", "dump", "--in", "legacy.bundle"}
					})

					It("should exit with status code 0", func() {
						Eventually(session).Should(gexec.Exit(0))
					})

					It("should emit required content", func() {
						Eventually(func() bool {
							return json.Valid(session.Out.Contents())
						}).ShouldNot(BeTrue())
					})
				})
			})

			Context("with content-only flag specified", func() {
				BeforeEach(func() {
					copyIn(fixturePath("bundles"), pkgPath, false)
				})

				Context("when input not exists", func() {
					BeforeEach(func() {
						cmdParams = []string{"bundle", "dump", "--in", "not-existent-bundle.bundle", "--content-only"}
					})

					It("exits with status code 0", func() {
						Eventually(session).Should(gexec.Exit(1))
					})

					It("should emit required content", func() {
						Eventually(session.Err).Should(gbytes.Say("fatal"))
						Eventually(session.Err).Should(gbytes.Say(`unable to open 'not-existent-bundle.bundle' for read`))
					})
				})

				Context("use a file as input", func() {
					BeforeEach(func() {
						cmdParams = []string{"bundle", "dump", "--in", "legacy.bundle", "--content-only"}
					})

					It("should exit with status code 0", func() {
						Eventually(session).Should(gexec.Exit(0))
					})

					It("should emit required content", func() {
						Eventually(func() bool {
							return json.Valid(session.Out.Contents())
						}).ShouldNot(BeTrue())
					})
				})
			})

			Context("with path-only flag specified", func() {
				BeforeEach(func() {
					copyIn(fixturePath("bundles"), pkgPath, false)
				})

				Context("when input not exists", func() {
					BeforeEach(func() {
						cmdParams = []string{"bundle", "dump", "--in", "not-existent-bundle.bundle", "--path-only"}
					})

					It("exits with status code 0", func() {
						Eventually(session).Should(gexec.Exit(1))
					})

					It("should emit required content", func() {
						Eventually(session.Err).Should(gbytes.Say("fatal"))
						Eventually(session.Err).Should(gbytes.Say(`unable to open 'not-existent-bundle.bundle' for read`))
					})
				})

				Context("use a file as input", func() {
					BeforeEach(func() {
						cmdParams = []string{"bundle", "dump", "--in", "legacy.bundle", "--path-only"}
					})

					It("should exit with status code 0", func() {
						Eventually(session).Should(gexec.Exit(0))
					})

					It("should emit required content", func() {
						Eventually(func() bool {
							return json.Valid(session.Out.Contents())
						}).ShouldNot(BeTrue())
					})
				})
			})
		})
	})
})
