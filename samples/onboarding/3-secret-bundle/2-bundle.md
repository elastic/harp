# Bundle

## Specification

This [specification](https://github.com/elastic/harp/blob/main/api/proto/harp/bundle/v1/bundle.proto) declares internal secret storage object structure.

This is used to store data and metadata related to objects.

### File

> `File` will be renammed to `Bundle` in future version for vocabulary consistency.

```cpp
// File is a concrete secret bundle file.
message File {
  // Map of string keys and values that can be used to organize and categorize
  // (scope and select) objects.
  map<string,string> labels = 1;
  // Annotations is an unstructured key value map stored with a resource that
  // may be set by external tools to store and retrieve arbitrary metadata.
  map<string,string> annotations = 2;
  // Version of the file
  uint32 version = 3;
  // Secret package collection
  repeated Package packages = 4;
  // Bundle template object
  Template template = 5;
  // Associated values
  google.protobuf.BytesValue values = 6;
  // MerkleTreeRoot value of content
  bytes merkle_tree_root = 7;
}
```

* The `labels` are used for Bundle tagging purpose (query, filtering, etc.).
* The `annotations` are produced by external secret storage reader, and consumed
  by secret bundle writers.
* The `version` is an unsigned 32bit integer of the Bundle version (must be `1`).
* The `packages` are a list of `Package` object.
* The `template`is the BundleTemplate used to generate the `Bundle`.
* The `values` is the encoded values used for Bundle generation.

### Package

A `Package` is addressed by a `name` (aka secret path), and refer to secret
data collection chain.

```cpp
// Package is a secret organizational unit.
message Package {
  // Map of string keys and values that can be used to organize and categorize
  // (scope and select) objects.
  map<string,string> labels = 1;
  // Annotations is an unstructured key value map stored with a resource that
  // may be set by external tools to store and retrieve arbitrary metadata.
  map<string,string> annotations = 2;
  // Package name as a complete secret path (CSO compliance recommended)
  string name = 3;
  // Active secret version
  SecretChain secrets = 4;
  // SecretChain versions
  map<fixed32, SecretChain> versions = 5;
}
```

* The `labels` are used for Bundle tagging purpose (query, filtering, etc.).
* The `annotations` are produced by external secret storage reader, and consumed
  by secret bundle writers.
* The `name` is the package name generally used as a path.
* The `secrets` is the last version of secret data.
* The `version` is a map of SecretChain where key is the version, and value the
  data.

### Secret Chain

A `Secret Chain` is an item of versioned data.

```cpp
// SecretChain describe a secret version chain.
message SecretChain {
  // Map of string keys and values that can be used to organize and categorize
  // (scope and select) objects.
  map<string,string> labels = 1;
  // Annotations is an unstructured key value map stored with a resource that
  // may be set by external tools to store and retrieve arbitrary metadata.
  map<string,string> annotations = 2;
  // Version identifier
  fixed32 version = 3;
  // Secret K/V collection
  repeated KV data = 4;
  // Link to previous version
  google.protobuf.UInt32Value previous_version = 5;
  // Link to next version
  google.protobuf.UInt32Value next_version = 6;
  // Locked buffer when encryption is enabled
  google.protobuf.BytesValue locked = 7;
}
```

* The `labels` are used for Bundle tagging purpose (query, filtering, etc.).
* The `annotations` are produced by external secret storage reader, and consumed
  by secret bundle writers.
* The `version` sets the version number of the current SecretChain.
* The `data` is a list of `KV` used to store information addressed by the package
  name.
* The `previous_version` references the previous version index.
* The `next_version` references the next version index.
* The `locked` attribute refers to KV 2nd encryption layer used for legacy writer.

### Key/Value item

The `KV` defines a list element with a `key`, a `type` of value, and the
serialized `value`.

```cpp
// KV contains the key, the value and the type of the value.
message KV {
  // Key
  string key = 1;
  // Golang type of initial value before packing
  string type = 2;
  // Value must be encoded using secret.Pack method
  bytes value = 3;
}
```

## Usages

The `Bundle` struct can be observed by using the following command :

```sh
$ harp bundle dump --in secrets.container
{
    ... JSON version of the bundle ...
}
```

### Minimal Bundle

Given this command :

```sh
echo '{"secrets/path":{"foo":"bar"}}' | harp from jsonmap | harp bundle dump | jq
```

> Generate a `Bundle` using input JSON, and then display the JSON Bundle
> representation.

It will generate a minimal `Bundle` object like the following :

```json
// File
{
  "packages": [
    {
      "name": "secrets/path",
      "secrets": {
        "data": [
          {
            "key": "foo",
            "type": "string",
            "value": "EwNiYXI="
          }
        ]
      }
    }
  ],
  "merkleTreeRoot": "GaV1ySQ2Do12L/DnFzDNtClPWQitAiSCFTtnJ2GzvAQnGn7mEVQS+scgiQWtQlk/K7Pvd1pNVPcN+d4rtqzF8A=="
}
```

> Value is encoded using `asn.1` serialization format.

### Generated Bundle

Given this command :

```sh
harp from found-secrets --in cmd/harp/test/fixtures/found-secrets/valid/bundle.yaml | harp bundle dump | jq
```

> Generate a `Bundle` using a `found-secrets` YAML bundle, and then display the JSON Bundle
> representation.

It will generate an output like the following :

```json
// File
{
  "annotations": {
    // Annotations produced by the found-secrets reader
    "secret-service.elstc.co/bundleImportDate": "2020-10-06T12:33:02Z",
    "secret-service.elstc.co/vaultBackendPath": "secrets",
    "secret-service.elstc.co/vaultPathPrefix": "local/dev/test-region/testapp"
  },
  "version": 1,
  "packages": [
    // Package
    {
      "annotations": {
        // Annotations produced by the found-secrets reader
        "secret-service.elstc.co/bundleEncryptionKey": "v6bG94y9geWy4dVncgCBAEBg8D0hpqBs4Q1E4kZybfo=",
        "secret-service.elstc.co/bundleName": "dist",
        "secret-service.elstc.co/packageEncryptionKey": "ECH9By8ZGLF3qv86b20UVJRh6wswXxSUO2zMkuahlq0=",
        "secret-service.elstc.io/packageFileName": "secrets/dist/env.yaml"
      },
      // Secret path
      "name": "secrets/dist/env.yaml",
      "secrets": {
        // SecretChain
        "data": [
          // KV
          {
            "key": "DIST_ENV",
            "type": "string",
            "value": "DBtzaW1wbGVfZW52X3ZhbHVlX2F3ZXNvbWUhISE="
          }
        ]
      }
    }
  ],
  "merkleTreeRoot": "pN8gRh1gqZ1w8ApgzqQFP1rjsmwaj14C3SWMoTOehw5oSjEJ5T5hgv02xvK43YT2d5lChMkZVeKcmCdFIMdKgQ==",
}
```

### Bundle content

Secret data is encoded internally to prevent display problems.

You can dump a bundle content as a JSON map by using the following command :

```sh
$ harp bundle dump --content-only --in secrets.container
{
  "secrets/dist/env.yaml": {
    "DIST_ENV": "dist_env_value_awesome!!!"
  },
  "secrets/simple/common_file.yaml": {
    "common_file.conf": "Y29tbW9uIGZpbGUgZm9yIGNvbmZpZ3VyYXRpb24K"
  },
  "secrets/simple/dev/dev.yaml": {
    "dev.sh": "IyEvYmluL2Jhc2gKCnNldCAteGUKCmVjaG8gInRlc3QgbmVzdGVkIGRpcmVjdG9yeSIK"
  },
  "secrets/simple/env.yaml": {
    "SIMPLE_ENV": "simple_env_value_awesome!!!"
  },
  "secrets/simple/file.yaml": {
    "logstashish.conf": "aW5wdXQgewogICAgYmVhdHMgewogICAgICAgIGlzaCA9PiB0cnVlCiAgICB9Cn0K"
  }
}
```

> All metadata are lost by doing this operation, but it's a lot easier to inspect
> data.

---

* [Previous topic](1-introduction.md)
* [Index](../)
* [Next topic](3-template.md)
