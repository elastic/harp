## 0.2.1

### Not released yet

BREAKING-CHANGES:

* cmd/ruleset: Ruleset generation from a Bundle has been relocated to `to ruleset` command. [#77](https://github.com/elastic/harp/pull/77)
* bundle/filter: parameter `--jmespath` as been renamed to `--query`. [#77](https://github.com/elastic/harp/pull/77)
* bundle/dump: parameter `--jmespath` as been renamed to `--query`. [#77](https://github.com/elastic/harp/pull/77)
* deprecation: package `github.com/elastic/harp/pkg/bundle/vfs` has been removed. The Golang 1.16 `fs.FS` implementation must be used and located at `github.com/elastic/harp/pkg/bundle/fs`. [#77](https://github.com/elastic/harp/pull/77)
* container/identity: identities are using `ed25519` key pairs vs `x25519` keys in previous versions. Allows identities to be used for signing and encryption purpose. [#79](https://github.com/elastic/harp/pull/80)
* sdk/transformer: Encryption transformers must be imported to be registered in the encryption transformer registry. [#80](https://github.com/elastic/harp/pull/80)

FEATURES:

* bundle/encryption: Partial bundle encryption based on annotations. [#77](https://github.com/elastic/harp/pull/77)
* task/bundle: Fully unit tested. [#77](https://github.com/elastic/harp/pull/77)
* core/kv: Support KV Store publication for Etcd3/Zookeeper/Consul. [#77](https://github.com/elastic/harp/pull/77)
* value/transformer: Transformer mock is available for testing. [#77](https://github.com/elastic/harp/pull/77)
* value/encryption: Expose `encryption.Must(value.Transformer, error)` to build a transformer instance with a panic raised on error. [#77](https://github.com/elastic/harp/pull/77)
* sdk/types: `IsNill()` now recognize nil function pointer. [#77](https://github.com/elastic/harp/pull/77)
* sdk/cmdutil: `DiscardWriter()` is a `io.Writer` provider used to discard all output. [#77](https://github.com/elastic/harp/pull/77)
* sdk/cmdutil: `DirectWriter(io.Writer)` is a `io.Writer` provider used to delegate to input writer. [#77](https://github.com/elastic/harp/pull/77)
* sdk/cmdutil: `NewClosedWriter()` is a `io.Writer` implementation who always return on `Write()` calls. [#77](https://github.com/elastic/harp/pull/77)
* pkg/kv: integration tests and behavior validation test suite. [#78](https://github.com/elastic/harp/pull/78)
* value/transformers: expose new JWE based encryption transformers [#80](https://github.com/elastic/harp/pull/80)
  * `jwe:a128kw:<base64>` to initialize a AES128 Key Wrapper with AES128 GCM Encryption transformer
  * `jwe:a192kw:<base64>` to initialize a AES192 Key Wrapper with AES192 GCM Encryption transformer
  * `jwe:a256kw:<base64>` to initialize a AES256 Key Wrapper with AES256 GCM Encryption transformer
  * `jwe:pbes2-a128kw:<ascii>` to initialize a PBES2 key derivation function for AES128 key wrapping with AES128 GCM Encryption transformer
  * `jwe:pbes2-a192kw:<ascii>` to initialize a PBES2 key derivation function for AES192 key wrapping with AES192 GCM Encryption transformer
  * `jwe:pbes2-a256kw:<ascii>` to initialize a PBES2 key derivation function for AES256 key wrapping with AES256 GCM Encryption transformer
* sdk/transformer: Encryption transformer dynamic factory. [#80](https://github.com/elastic/harp/pull/80)
  * Use `github.com/elastic/harp/pkg/value/encryption.Register(prefix, factory)` to register a transformer factory matching the given prefix.
* bundle/prefixer: parameter `--remove` added to support prefix removal operation [#81](https://github.com/elastic/harp/pull/81)

CHANGES:

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
