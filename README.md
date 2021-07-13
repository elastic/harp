# Harp

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/elastic/harp)](https://goreportcard.com/report/github.com/elastic/harp)
[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](http://golang.org)
[![GitHub release](https://img.shields.io/github/release/elastic/harp.svg)](https://github.com/elastic/harp/releases/)
[![Maintenance](https://img.shields.io/badge/Maintained%3F-yes-green.svg)](https://github.com/elastic/harp/graphs/commit-activity)

Harp is for Harpocrates (Ancient Greek: Ἁρποκράτης) the god of silence, secrets
and confidentiality in the Hellenistic religion. - [Wikipedia](https://en.wikipedia.org/wiki/Harpocrates)

## Why harp?

* Secret management is in essence a collection of processes that must be
  auditable, executable and reproducible for infosec and operation requirements;
* Secret provisioning must be designed with secret rotation as a day one task,
  due to the fact that secret data must be rotated periodically to keep its
  secret property;
* `Developers` should negotiate secret value for the secret consumer they are
  currently developing, by the contract based on a path (reference to the secret)
  and a value specification (for code contract) without the knowledge of the
  final deployed value;
* `Secret Operators` use different set of tools to achieve secret
  management operation which increases the error/secret exposure probability due to
  tool count involved in the process (incompatibility, changes, etc.);
* Without a defined secret naming convention, the secret storage becomes difficult to
  handle in time (naming is hard) and secret naming could not be helped to
  get a consistent, reliable and flexible secret tree;
* Secret storage backend can use various implementations in different environments
  and should be provisioned consistently.

## How does it work?

![Secret management Pipeline](docs/harp/img/SM-HARP-PIPELINE.png)

### Like a Data pipeline but for secret

`harp` allows you to handle secrets using deterministic pipelines expressed
using an atomic series of CLI operations applied to a commonly shared container
immutable and standalone filesystem used to store secret collection (Bundle)
generated from a template engine via user specification, or external secret
value coming from files or external secret storage.

![Pipelines](docs/harp/img/SM-HARP.png)

These pipelines use the immutable container filesystem as a data exchange
protocol and could be extended for new input, intermediary operation or output
via plugins created with the `harp` SDK.

### Immutable transformation

Each applied transformation create a container with transformed data inside.
This will enforce container reproducibility by eliminating cumulative
side-effects applied to the same container.

The container handles for you the confidentiality and integrity protection applied
to the secret collection stored inside and manipulated by copy during the
pipeline execution.

## What can I do?

> New to harp, let's start with [onboarding tutorial](samples/onboarding/README.md) !
> TL;DR - [Features overview](FEATURES.md)

Harp provides :

* A methodology to design your secret management;
  * Secret naming convention (CSO);
  * A defined common language and complete processes to achieve secret management
    operations;
* A SDK to create your own tools to orchestrate your secret management pipelines;
  * A container manipulation library exposed as `github.com/elastic/harp/pkg/container`;
  * A secret bundle specification to store and manipulate secrets exposed as `github.com/elastic/harp/pkg/bundle`;
  * An `on-steroid` template engine exposed as `github.com/elastic/harp/pkg/template`
  * A path name validation library exposed as `github.com/elastic/harp/pkg/cso`
* A CLI for secret management implementation
  * CI/CD integration;
  * Based on human readable definitions (YAML);
  * In order to create auditable and reproducible pipelines.
  * An extensible tool which can be enhanced via [plugins](https://github.com/elastic/harp-plugins).

And allows :

* Bundle level operations
  * Create a bundle from scratch / template / json (more via plugins);
  * Generate a complete bundle using a YAML Descriptor (`BundleTemplate`) to describe secret and their usages;
  * Read value stored in the K/V virtual filesystem;
  * Update the K/V virtual filesystem;
  * Reproductible patch applied on immutable container (copy-on-write);
  * Import / Export to Vault.
* Immutable container level operations
  * Seal / Unseal a container for integrity and confidentiality property conservation
    to enforce at-rest encryption (aes256-gcm96 or chacha20-poly1305);
  * Multiple identities sealing algorithm;

## License

`harp` artifacts and source code is released under [Apache 2.0 Software License](LICENSE).

## Homebrew install

Download a [release](https://github.com/elastic/harp/releases) or build from source.

For stable version

```sh
brew tap elastic/harp
brew install elastic/harp/harp
```

## Build instructions

Download a [release](https://github.com/elastic/harp/releases) or build from source.

### First time

#### Check your go version

> Only last 2 minor versions of a major are supported.

`Harp` is compiled with :

```sh
$ go version
go version go1.16.6 linux/amd64
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

You have to build this image once before executing artifact pipelines.

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

## Plugins

You can find more Harp feature extensions - <https://github.com/elastic/harp-plugins>

## Community

Here is the list of external projects used as inspiration :

* [Kubernetes](https://github.com/kubernetes/)
* [Helm](https://github.com/helm/)
* [Open Policy Agent ConfTest](https://github.com/open-policy-agent/conftest)
* [SaltPack](https://github.com/keybase/saltpack)
* [Hashicorp Vault](https://github.com/hashicorp/vault)
* [AWS SDK Go](https://github.com/aws/aws-sdk-go)
