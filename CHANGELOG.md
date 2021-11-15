## 0.2.1

### Not released yet

BREAKING-CHANGES:

* cmd/ruleset: Ruleset generation from a Bundle has been relocated to `to ruleset` command. [#77](https://github.com/elastic/harp/pull/77)
* bundle/filter: Parameter `--jmespath` as been renamed to `--query`. [#77](https://github.com/elastic/harp/pull/77)
* bundle/dump: Parameter `--jmespath` as been renamed to `--query`. [#77](https://github.com/elastic/harp/pull/77)
* deprecation: Package `github.com/elastic/harp/pkg/bundle/vfs` has been removed. The Golang 1.16 `fs.FS` implementation must be used and located at `github.com/elastic/harp/pkg/bundle/fs`. [#77](https://github.com/elastic/harp/pull/77)

CHANGES:

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
