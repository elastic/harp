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
	"bytes"
	"os"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"github.com/elastic/harp/pkg/bundle"
)

const (
	fernetKey    = "C0_4dWyx6McFlLaYchjRI_jzEItNOxr9TitOqRu5b04="
	aes256Key    = "aes-gcm:Bh8GQlNwH2xX7qviJXYXdRTtsYc1vF7GVNjxFm6kDyA="
	secretBoxKey = "secretbox:ckQSo9PZ-Yrgy7Q5JiH2d7li716xnUjjTKcNyd8Tomc="
)

var _ = Describe("Harp CLI", func() {
	Describe("bundle encrypt/decrypt", func() {
		var (
			cmdParams []string
			pkgPath   string
			session   *gexec.Session
		)

		BeforeEach(func() {
			pkgPath = tmpPath("bundle-encrypt-decrypt")
			os.Mkdir(pkgPath, 0o777)
		})

		JustBeforeEach(func() {
			session = startHarp(pkgPath, cmdParams...)
		})

		Context("when encrypting a bundle", func() {
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
						cmdParams = []string{"bundle", "encrypt", "--in", "not-existent-bundle.bundle", "--key", fernetKey}
					})

					It("exits with status code 0", func() {
						Eventually(session).Should(gexec.Exit(1))
					})

					It("should emit required content", func() {
						Eventually(session.Err).Should(gbytes.Say("fatal"))
						Eventually(session.Err).Should(gbytes.Say(`unable to open 'not-existent-bundle.bundle' for read`))
					})
				})

				Context("using fernet encryption", func() {
					Context("use stdout as output", func() {
						BeforeEach(func() {
							cmdParams = []string{"bundle", "encrypt", "--in", "legacy.bundle", "--key", fernetKey}
						})

						It("should exit with status code 0", func() {
							Eventually(session).Should(gexec.Exit(0))
							Eventually(func() error {
								_, err := bundle.FromContainerReader(bytes.NewReader(session.Out.Contents()))
								return err
							}).ShouldNot(HaveOccurred())
						})
					})

					Context("use a file as output", func() {
						BeforeEach(func() {
							cmdParams = []string{"bundle", "encrypt", "--in", "legacy.bundle", "--key", fernetKey, "--out", "fernet.bundle"}
						})

						It("should exit with status code 0", func() {
							Eventually(session).Should(gexec.Exit(0))
						})

						It("should emit required content", func() {
							Eventually(path.Join(pkgPath, "fernet.bundle")).Should(BeARegularFile())
							Eventually(func() error {
								_, err := bundle.FromContainerReader(fromFile(pkgPath, "fernet.bundle"))
								return err
							}).ShouldNot(HaveOccurred())
						})
					})
				})

				Context("using aes-256 encryption", func() {
					Context("use stdout as output", func() {
						BeforeEach(func() {
							cmdParams = []string{"bundle", "encrypt", "--in", "legacy.bundle", "--key", aes256Key}
						})

						It("should exit with status code 0", func() {
							Eventually(session).Should(gexec.Exit(0))
							Eventually(func() error {
								_, err := bundle.FromContainerReader(bytes.NewReader(session.Out.Contents()))
								return err
							}).ShouldNot(HaveOccurred())
						})
					})

					Context("use a file as output", func() {
						BeforeEach(func() {
							cmdParams = []string{"bundle", "encrypt", "--in", "legacy.bundle", "--key", aes256Key, "--out", "aes-gcm.bundle"}
						})

						It("should exit with status code 0", func() {
							Eventually(session).Should(gexec.Exit(0))
						})

						It("should emit required content", func() {
							Eventually(path.Join(pkgPath, "aes-gcm.bundle")).Should(BeARegularFile())
							Eventually(func() error {
								_, err := bundle.FromContainerReader(fromFile(pkgPath, "aes-gcm.bundle"))
								return err
							}).ShouldNot(HaveOccurred())
						})
					})
				})

				Context("using secretbox encryption", func() {
					Context("use stdout as output", func() {
						BeforeEach(func() {
							cmdParams = []string{"bundle", "encrypt", "--in", "legacy.bundle", "--key", secretBoxKey}
						})

						It("should exit with status code 0", func() {
							Eventually(session).Should(gexec.Exit(0))
							Eventually(func() error {
								_, err := bundle.FromContainerReader(bytes.NewReader(session.Out.Contents()))
								return err
							}).ShouldNot(HaveOccurred())
						})
					})

					Context("use a file as output", func() {
						BeforeEach(func() {
							cmdParams = []string{"bundle", "encrypt", "--in", "legacy.bundle", "--key", secretBoxKey, "--out", "secretbox.bundle"}
						})

						It("should exit with status code 0", func() {
							Eventually(session).Should(gexec.Exit(0))
						})

						It("should emit required content", func() {
							Eventually(path.Join(pkgPath, "secretbox.bundle")).Should(BeARegularFile())
							Eventually(func() error {
								_, err := bundle.FromContainerReader(fromFile(pkgPath, "secretbox.bundle"))
								return err
							}).ShouldNot(HaveOccurred())
						})
					})
				})
			})
		})

		Context("when decrypting a bundle", func() {
			Context("when no arguments specified", func() {
				BeforeEach(func() {
					cmdParams = []string{"bundle", "decrypt"}
				})

				It("exits with status code 0", func() {
					Eventually(session).Should(gexec.Exit(0))
				})

				It("should emit required content", func() {
					Eventually(session.Err).Should(gbytes.Say(`Error: required flag\(s\) "key" not set`))
				})
			})

			Context("using fernet encryption", func() {
				BeforeEach(func() {
					copyIn(fixturePath("bundles"), pkgPath, false)
				})

				Context("use stdout as output", func() {
					BeforeEach(func() {
						cmdParams = []string{"bundle", "decrypt", "--in", "legacy-fernet.bundle", "--key", fernetKey}
					})

					It("should exit with status code 0", func() {
						Eventually(session).Should(gexec.Exit(0))
						Eventually(func() error {
							_, err := bundle.FromContainerReader(bytes.NewReader(session.Out.Contents()))
							return err
						}).ShouldNot(HaveOccurred())
					})
				})

				Context("use a file as output", func() {
					BeforeEach(func() {
						cmdParams = []string{"bundle", "decrypt", "--in", "legacy-fernet.bundle", "--key", fernetKey, "--out", "legacy-fernet-decrypted.bundle"}
					})

					It("should exit with status code 0", func() {
						Eventually(session).Should(gexec.Exit(0))
					})

					It("should emit required content", func() {
						Eventually(path.Join(pkgPath, "legacy-fernet-decrypted.bundle")).Should(BeARegularFile())
						Eventually(func() error {
							_, err := bundle.FromContainerReader(fromFile(pkgPath, "legacy-fernet-decrypted.bundle"))
							return err
						}).ShouldNot(HaveOccurred())
					})
				})
			})

			Context("using aes256 encryption", func() {
				BeforeEach(func() {
					copyIn(fixturePath("bundles"), pkgPath, false)
				})

				Context("use stdout as output", func() {
					BeforeEach(func() {
						cmdParams = []string{"bundle", "decrypt", "--in", "legacy-aes256.bundle", "--key", aes256Key}
					})

					It("should exit with status code 0", func() {
						Eventually(session).Should(gexec.Exit(0))
						Eventually(func() error {
							_, err := bundle.FromContainerReader(bytes.NewReader(session.Out.Contents()))
							return err
						}).ShouldNot(HaveOccurred())
					})
				})

				Context("use a file as output", func() {
					BeforeEach(func() {
						cmdParams = []string{"bundle", "decrypt", "--in", "legacy-aes256.bundle", "--key", aes256Key, "--out", "legacy-aes256-decrypted.bundle"}
					})

					It("should exit with status code 0", func() {
						Eventually(session).Should(gexec.Exit(0))
					})

					It("should emit required content", func() {
						Eventually(path.Join(pkgPath, "legacy-aes256-decrypted.bundle")).Should(BeARegularFile())
						Eventually(func() error {
							_, err := bundle.FromContainerReader(fromFile(pkgPath, "legacy-aes256-decrypted.bundle"))
							return err
						}).ShouldNot(HaveOccurred())
					})
				})
			})

			Context("using secretbox encryption", func() {
				BeforeEach(func() {
					copyIn(fixturePath("bundles"), pkgPath, false)
				})

				Context("use stdout as output", func() {
					BeforeEach(func() {
						cmdParams = []string{"bundle", "decrypt", "--in", "legacy-secretbox.bundle", "--key", secretBoxKey}
					})

					It("should exit with status code 0", func() {
						Eventually(session).Should(gexec.Exit(0))
						Eventually(func() error {
							_, err := bundle.FromContainerReader(bytes.NewReader(session.Out.Contents()))
							return err
						}).ShouldNot(HaveOccurred())
					})
				})

				Context("use a file as output", func() {
					BeforeEach(func() {
						cmdParams = []string{"bundle", "decrypt", "--in", "legacy-secretbox.bundle", "--key", secretBoxKey, "--out", "legacy-secretbox-decrypted.bundle"}
					})

					It("should exit with status code 0", func() {
						Eventually(session).Should(gexec.Exit(0))
					})

					It("should emit required content", func() {
						Eventually(path.Join(pkgPath, "legacy-secretbox-decrypted.bundle")).Should(BeARegularFile())
						Eventually(func() error {
							_, err := bundle.FromContainerReader(fromFile(pkgPath, "legacy-secretbox-decrypted.bundle"))
							return err
						}).ShouldNot(HaveOccurred())
					})
				})
			})
		})
	})
})
