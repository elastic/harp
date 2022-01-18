## 0.2.5

## Not released yet

FEATURES:

* template/engine [#95](https://github.com/elastic/harp/pull/95)
  * `parseJwt` to parse JWT without signature validation
  * `verifyJwt` to parse a JWT with signature validation

DIST:

* sdk/tools:
  * Replace `go-header` dependency by `github.com/denis-tingaikin/go-header` to prevent a possible identity spoofing. [#96](https://github.com/elastic/harp/pull/96)

## 0.2.4

### 2022-01-14

DIST:

* Github actions release automation
* go: Build with Golang 1.17.6.

## 0.2.3

### 2021-12-10

FEATURES:

* container/seal: introduce a naming convention for identity and container keys. [#89](https://github.com/elastic/harp/pull/89)
* cmd/transform [#90](https://github.com/elastic/harp/pull/90)
  * `encrypt` / `decrypt` apply symmetric encryption transformer
  * `encode` / `decode` apply encoding/decoding to given input
  * `sign` / `verify` apply signature algorithm or verify a signature from the given input
* cmd/keygen: JWK Key pair generation [#90](https://github.com/elastic/harp/pull/90)

CHANGES:

* cso/v1: Meta ring only require one path component. [#90](https://github.com/elastic/harp/pull/90)
* container/seal: Modern FIPS compatible container sealing process (ECDH+AES256-CTR+HMAC-SHA384 / ECDSA P-384 / HMAC-SHA512). [#89](https://github.com/elastic/harp/pull/89)
* crypto/paseto: move PASETO v4 primitives to `sdk/security/paseto/v4`. [#87](https://github.com/elastic/harp/pull/87)
* sdk/deps [#91](https://github.com/elastic/harp/pull/91)
  * GHSA - Security freeze
    * github.com/opencontainers/image-spec v1.0.2
    * github.com/opencontainers/runc v1.0.3
  * github.com/hashicorp/hcl/v2 v2.11.1
  *	github.com/ory/dockertest/v3 v3.8.1
  * golang.org/x/crypto v0.0.0-20210915214749-c084706c2272
  * golang.org/x/sys v0.0.0-20210915083310-ed5796bab164
  * golang.org/x/term v0.0.0-20201126162022-7de9c90e9dd1
  * google.golang.org/genproto v0.0.0-20211207154714-918901c715cf
* cmd/transform: Deprecate `encryption` sub command in favor of `encrypt` and `decrypt`. [#90](https://github.com/elastic/harp/pull/90)

DIST:

* go: Build with Golang 1.17.5.
* nix/shell: Expose `shell.nix` to get a consistent development environment. [#87](https://github.com/elastic/harp/pull/87)

## 0.2.2

### 2021-11-24

CHANGES:

* cso/v1: Support new Azure and IBM regions. [#84](https://github.com/elastic/harp/pull/84)

## 0.2.1

### 2021-11-17

BREAKING-CHANGES:

* cmd/ruleset: Ruleset generation from a Bundle has been relocated to `to ruleset` command. [#77](https://github.com/elastic/harp/pull/77)
* bundle/filter: parameter `--jmespath` as been renamed to `--query`. [#77](https://github.com/elastic/harp/pull/77)
* bundle/dump: parameter `--jmespath` as been renamed to `--query`. [#77](https://github.com/elastic/harp/pull/77)
* deprecation: package `github.com/elastic/harp/pkg/bundle/vfs` has been removed. The Golang 1.16 `fs.FS` implementation must be used and located at `github.com/elastic/harp/pkg/bundle/fs`. [#77](https://github.com/elastic/harp/pull/77)
* container/identity: identities are using `ed25519` key pairs vs `x25519` keys in previous versions. For conversion, you can still unseal a container using old `x25519` key based identities, but you can't seal with them. To be future-proof, you have to regenerate new identities. [#79](https://github.com/elastic/harp/pull/80)
* sdk/transformer: Encryption transformers must be imported to be registered in the encryption transformer registry. [#80](https://github.com/elastic/harp/pull/80)

FEATURES:

* bundle/encryption: Partial bundle encryption based on annotations. [#77](https://github.com/elastic/harp/pull/77)
* task/bundle: Fully unit tested. [#77](https://github.com/elastic/harp/pull/77)
* core/kv: Support KV Store publication for Etcd3/Zookeeper/Consul. [#77](https://github.com/elastic/harp/pull/77)
* value/transformer: Transformer mock is available for testing. [#77](https://github.com/elastic/harp/pull/77)
* value/encryption: Expose `encryption.Must(value.Transformer, error)` to build a transformer instance with a panic raised on error. [#77](https://github.com/elastic/harp/pull/77)
* sdk/cmdutil: `DiscardWriter()` is a `io.Writer` provider used to discard all output. [#77](https://github.com/elastic/harp/pull/77)
* sdk/cmdutil: `DirectWriter(io.Writer)` is a `io.Writer` provider used to delegate to input writer. [#77](https://github.com/elastic/harp/pull/77)
* sdk/cmdutil: `NewClosedWriter()` is a `io.Writer` implementation who always return on `Write()` calls. [#77](https://github.com/elastic/harp/pull/77)
* pkg/kv: integration tests and behavior validation test suite. [#78](https://github.com/elastic/harp/pull/78)
* value/transformers: expose new JWE based encryption transformers [#80](https://github.com/elastic/harp/pull/80)
  * `jwe:a128kw:<base64>` to initialize a AES128 Key Wrapper with AES128 GCM Encryption transformer
  * `jwe:a192kw:<base64>` to initialize a AES192 Key Wrapper with AES192 GCM Encryption transformer
  * `jwe:a256kw:<base64>` to initialize a AES256 Key Wrapper with AES256 GCM Encryption transformer
  * `jwe:pbes2-hs256-a128kw:<ascii>` to initialize a PBES2 key derivation function for AES128 key wrapping with AES128 GCM Encryption transformer
  * `jwe:pbes2-hs384-a192kw:<ascii>` to initialize a PBES2 key derivation function for AES192 key wrapping with AES192 GCM Encryption transformer
  * `jwe:pbes2-hs512-a256kw:<ascii>` to initialize a PBES2 key derivation function for AES256 key wrapping with AES256 GCM Encryption transformer
* sdk/transformer: Encryption transformer dynamic factory. [#80](https://github.com/elastic/harp/pull/80)
  * Use `pkg/value/encryption.Register(prefix, factory)` to register a transformer factory matching the given prefix.
* bundle/prefixer: parameter `--remove` added to support prefix removal operation. [#81](https://github.com/elastic/harp/pull/81)
* to/object: support `toml` format as output. [#81](https://github.com/elastic/harp/pull/81)
* value/transformer: Support PASETO `v4.local` transformer. [#82](https://github.com/elastic/harp/pull/82)

CHANGES:

* container/identity: converge to `value.Transformer` usage for identity protection. [#81](https://github.com/elastic/harp/pull/81)
* container/recover: converge to `value.Transformer` usage for container key recovery from an identity. [#81](https://github.com/elastic/harp/pull/81)
* sdk/types: `IsNil()` now recognize nil function pointer. [#77](https://github.com/elastic/harp/pull/77)
* sdk/dep: [#79](https://github.com/elastic/harp/pull/79)
  * github.com/google/gops v0.3.22
  * github.com/gosimple/slug v1.11.2
  * github.com/hashicorp/consul/api v1.11.0
  * github.com/hashicorp/vault/api v1.3.0
  * github.com/zclconf/go-cty v1.10.0
  * go.step.sm/crypto v0.13.0
  * golang.org/x/crypto v0.0.0-20211108221036-ceb1ce70b4fa
  * golang.org/x/sys v0.0.0-20211113001501-0c823b97ae02
  * google.golang.org/genproto v0.0.0-20211112145013-271947fe86fd
  * google.golang.org/grpc v1.42.0

DIST:

* go: Build with Golang 1.17.3.
* tools: Update `golangci-lint` to `v1.43.0`. [#76](https://github.com/elastic/harp/pull/76)
* docs: General review for typo / grammar.

## 0.2.0

### 2021-10-26

BREAKING-CHANGES:

* Metadata storage has been modified to support a JSON level complexity. All plugins must align their metadata management to the new format.

DIST:

* go: Build with Golang 1.17.2.
* homebrew: Approriate harp version can be installed according to your platform architecture and OS [#71](https://github.com/elastic/harp/pull/71)

CHANGES:

* core/vault: Replace JSON encoded metadata in secret data by a JSON object. [#68](https://github.com/elastic/harp/pull/68)
* crypto/pem: Delegate PEM encoding/decoding to `go.step.sm/crypto` [#73](https://github.com/elastic/harp/pull/73)

FEATURES:

* to/vault: Support Vault >1.9 custom metadata for bundle metadata publication. [#68](https://github.com/elastic/harp/pull/68)
* from/vault: Support Vault >1.9 custom metadata for bundle metadata retrieval. [#68](https://github.com/elastic/harp/pull/68)
* from/vault: Support legacy bundle metadata format. [#69](https://github.com/elastic/harp/pull/69)
* template/engine: `jsonEscape` / `jsonUnescape` is added to handle string escaping using JSON character escaping strategy [#70](https://github.com/elastic/harp/pull/70)
* template/engine: `unquote` is added to unquote a `quote` escaped string. [#70](https://github.com/elastic/harp/pull/70)
* bundle/prefixer: Globally add a prefix to all secret packages. [#74](https://github.com/elastic/harp/pull/74)
* plugin/kv: Promote harp-kv as builtin. [#75](https://github.com/elastic/harp/pull/75)

## 0.1.24

### 2021-09-20

CHANGES:

* go: Build with Golang 1.17.1.
