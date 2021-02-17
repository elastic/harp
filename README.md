# Harp

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/elastic/harp)](https://goreportcard.com/report/github.com/elastic/harp)
[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](http://golang.org)
[![GitHub release](https://img.shields.io/github/release/elastic/harp.svg)](https://github.com/elastic/harp/releases/)
[![Maintenance](https://img.shields.io/badge/Maintained%3F-yes-green.svg)](https://github.com/elastic/harp/graphs/commit-activity)

Harp is for Harpocrates (Ancient Greek: Ἁρποκράτης) the god of silence, secrets
and confidentiality in the Hellenistic religion. - [Wikipedia](https://en.wikipedia.org/wiki/Harpocrates)

> New to harp, let's start with [onboarding tutorial](samples/onboarding/README.md) !
> TL;DR - [Features overview](FEATURES.md)

Harp provides :

* An automation compatible secret management implementation
  * CI/CD integration;
  * Human readable definitions (YAML);
  * Auditable and reproductible pipelines.
* A secret container specification to store and manipulate data;
* A container manipulation library exposed as `github.com/elastic/harp/pkg/bundle`;
* An extensible CLI tool.

And allows :

* bundle level operations
  * Create a container from scratch / template / json (more via plugins);
  * Read value stored in the K/V virtual filesystem;
  * Update the K/V virtual filesystem;
  * Delete keys.
* immutable container level operations
  * Seal / Unseal a container for integrity and confidentiality property conservation;
  * Generate a complete set of data using a YAML Descriptor (`BundleTemplate`);
  * Reproductible patch applied on immutable container (copy-on-write);
  * Import / Export to Vault.

## License

`harp` artifacts and source code is released under [Apache 2.0 Software License](LICENSE).

## Homebrew install

Download a [release](https://github.com/elastic/harp/releases) or build from source.

For stable version

```sh
brew tap elastic/harp
brew install elastic/harp/harp
brew install elastic/harp/harp-server
```

## Build instructions

Download a [release](https://github.com/elastic/harp/releases) or build from source.

### First time

#### Check your go version

> Only last 2 minor versions of a major are supported.

`Harp` is compiled with :

```sh
$ go version
go version go1.16 linux/amd64
```

> Simple go version manager - <https://github.com/stefanmaric/g>

#### Install mage

[Mage](https://magefile.org/) is an alternative to Make where language used is Go.
You can install it using 2 different methods.

##### From source

```sh
# Install mage
git clone https://github.com/magefile/mage
cd mage
go run bootstrap.go
```

##### From brew formula

```sh
brew install mage
```

#### Clone repository

```sh
git clone git@github.com:elastic/harp.git
# Go to tools submodule
cd harp/tools
# Pull tools sources, compile them and install executable in tools/bin
mage
```

### Daily

```sh
export PATH=$HARP_REPO/tools/bin:$PATH
# Build harp in bin folder
mage
```

### Docker

For Tools

You have to build this image before executing artifact pipelines.

```sh
mage docker:tools
```

For CLI

```sh
# or docker image [distroless:static, rootless, noshell]
mage docker:harp
# To execute in the container
docker run --rm -ti --read-only elastic/harp:<version>
```

For container server

```sh
# or docker image [distroless:static, rootless, noshell]
mage docker:harpServer
# To execute in the container
docker run --rm -ti --read-only elastic/harp-server:<version>
```

## Community

Here is the list of external projects used as inspiration :

* [Kubernetes](https://github.com/kubernetes/)
* [Helm](https://github.com/helm/)
* [Open Policy Agent ConfTest](https://github.com/open-policy-agent/conftest)
* [SaltPack](https://github.com/keybase/saltpack)
* [Hashicorp Vault](https://github.com/hashicorp/vault)
* [AWS SDK Go](https://github.com/aws/aws-sdk-go)
