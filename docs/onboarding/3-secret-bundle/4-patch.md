# BundlePatch

A `BundlePatch` allows secret operators to describe `Bundle` modification and
make them reproducible. Each `BundlePatch`generate a new `Bundle` form the
bundle source without altering the source bundle.

- [BundlePatch](#bundlepatch)
  - [Specification](#specification)
    - [PatchOperation](#patchoperation)
      - [Create a key ⇒ value](#create-a-key--value)
      - [Update values](#update-values)
      - [Remove a key](#remove-a-key)
      - [Remove keys by pattern](#remove-keys-by-pattern)
      - [Replace keys](#replace-keys)
    - [PatchSpec](#patchspec)
      - [Sample](#sample)
    - [PatchRule](#patchrule)
      - [Sample](#sample-1)
    - [PatchSelector](#patchselector)
      - [Match by strict path](#match-by-strict-path)
      - [Match by regex path](#match-by-regex-path)
      - [Match by JMES filter](#match-by-jmes-filter)
      - [PatchSelectorMatchPath](#patchselectormatchpath)
    - [PatchPackage](#patchpackage)
      - [PatchPackagePath](#patchpackagepath)
      - [Rename a package](#rename-a-package)
      - [Remove a package](#remove-a-package)
      - [Create if not exists](#create-if-not-exists)
    - [PatchSecret](#patchsecret)
      - [Alter annotations](#alter-annotations)
      - [Alter labels](#alter-labels)
      - [Alter secret data](#alter-secret-data)
  - [Usage](#usage)
    - [Apply a patch](#apply-a-patch)
    - [Generate a patch from bundle difference](#generate-a-patch-from-bundle-difference)

## Specification

### PatchOperation

`PatchOperation` holds information used to alter a key/value formatted object.

* `add` is used to add a new (key => value) association in the map
* `remove` is used to remove a key from the map
* `update` is used to update a value from an existing `key` only
* `replaceKeys` is used to rename a key to another key in the map.

> All keys and values can contain template instructions.

```cpp
// PatchOperation represents atomic patch operations executable on a k/v map.
message PatchOperation {
  // Add a new case-sentitive key and value to related data map.
  // Key and Value can be templatized.
  map<string,string> add = 1;
  // Remove a case-sensitive key from related data map.
  // Key and Value can be templatized.
  repeated string remove = 2;
  // Update case-sensitive existing key from related data map.
  // Key and Value can be templatized.
  map<string,string> update = 3;
  // Replace case-sensitive existing key using the associated value.
  // Value can be templatized.
  map<string,string> replaceKeys = 4;
  // Remove all keys matching these given regexp.
  repeated string removeKeys = 5;
}
```

#### Create a key ⇒ value

```yaml
kv:
  create:
    new-key-1: value1
    # With template
    new-key-{{.Values.number}}: {{ strongPassword | toYaml }}
```

#### Update values

```yaml
kv:
  update:
    key-1: new-value
```

#### Remove a key

```yaml
kv:
  remove:
    - key-1
    - key-2
```

#### Remove keys by pattern

```yaml
kv:
  removeKeys:
    - "key-[0-9]+"
```

#### Replace keys

```yaml
kv:
  replaceKeys:
    "old-key": "new-key"
```

### PatchSpec

`PatchSpec` defines the ordered `PatchRule` collection to apply during the `Bundle`
transformation.

```cpp
// PatchSpec repesetns bundle patch specification holder.
message PatchSpec {
  // Patch selector rules. Applied in the declaration order.
  repeated PatchRule rules = 1;
}
```

#### Sample

```yaml
apiVersion: harp.elastic.co/v1
kind: BundlePatch
meta:
  name: patch-name
  description: Patch description to help
spec:
  rules:
  # Rule 1
  - selector: ...
  # Rule 2
  - selector: ...
```

### PatchRule

```cpp
// PatchRule represents an operation to apply to a given bundle.
message PatchRule {
  // Used to determine is patch strategy is applicable to the package.
  PatchSelector selector = 1;
  // Package patch operations.
  PatchPackage package = 2;
}
```

#### Sample

```yaml
rules:
  # Define who is eligible to the PatchOperation
  - selector:
      matchPath:
        strict: product/harp/v1.0.0/artifacts/attestations/cosign/private_key
    package:
      data:
        # Alter package secrets
        kv:
          replaceKeys:
            "key": "COSIGN_PRIVATE_KEY"
            "password": "COSIGN_PRIVATE_KEY_PASSWORD"
```


### PatchSelector

`PathSelector` defines the strategy used to mark the package as eligible to
`PatchOperation` execution.

```cpp
// PatchSelector represents selecting strategies used to match a bundle resource.
message PatchSelector {
  // Match a package by using its path (secret path).
  PatchSelectorMatchPath matchPath = 1;
  // Match a package using a JMESPath query.
  string jmesPath = 2;
}
```

#### Match by strict path

```yaml
selector:
  matchPath:
    strict: product/harp/v1.0.0/artifacts/attestations/cosign/private_key
```

#### Match by regex path

```yaml
selector:
  matchPath:
    regex: ^app/production/.*$
```

#### Match by JMES filter

```yaml
selector:
  jmesPath: labels.database == "postgres"
```

#### PatchSelectorMatchPath

`PatchSelectorMatchPath` is a package path matcher.

* `strict` is used to filter package path when strictly equal to the given value
* `regex` is used to match the package path with the given regular expression

```cpp
// PatchSelectorMatchPath represents package path matching strategies.
message PatchSelectorMatchPath {
  // Strict case-sensitive path matching.
  // Value can be templatized.
  string strict = 1;
  // Regex path matching.
  // Value can be templatized.
  string regex = 2;
}
```

### PatchPackage

`PatchPackage` represents operation applicable to the selected package.

```cpp
// PatchPackage represents package operations.
message PatchPackage {
  // Path operations.
  PatchPackagePath path = 1;
  // Annotation operations.
  PatchOperation annotations = 2;
  // Label operations.
  PatchOperation labels = 3;
  // Secret data operations.
  PatchSecret data = 4;
  // Flag as remove.
  bool remove = 5;
  // Flag to create if not exist.
  bool create = 6;
}
```


#### PatchPackagePath

This is used to rename the current package to another.

* `template` is used to define the new package path

The template exposes `.Path` as context value to retrieve the current value of
the package.

```cpp
// PatchPackagePath represents package path operations.
message PatchPackagePath {
  // Template used to completely rewrite the package path.
  string template = 1;
}
```

Sample

```yaml
template: |-
    app/production/global/clusters/1.0.0/bootstrap/{{ trimPrefix "services/production/global/observability/" .Path }}
```

#### Rename a package

```yaml
selector:
  matchPath:
    regex: "^services/production/global/clusters/*"
package:
  path:
    template: |-
        app/production/global/clusters/1.0.0/bootstrap/{{ trimPrefix "services/production/global/observability/" .Path }}
```


#### Remove a package

```yaml
selector:
  jmesFilter: labels.deprecated == true
package:
  remove: true
```

#### Create if not exists

> Package creation flag only works with `strict` path matcher.

```yaml
selector:
  path:
    strict: infra/aws/security/eu-central-1/ec2/keys/ssh/ed25519_keys
  package:
    create: true
    data:
      template: |-
        {
          "private": {{ $sshKey := cryptoPair "ssh" }}{{ $sshKey.Private | toSSH | toJson }},
          "public": "{{ $sshKey.Public | toSSH | trim }} cloud-security@elastic.co"
        }
```

### PatchSecret

```cpp
// PatchSecret represents secret data operations.
message PatchSecret {
  // Secret data annotation operations.
  PatchOperation annotations = 1;
  // Secret data label operations.
  PatchOperation labels = 2;
  // Template to override secret data.
  string template = 3;
  // Used to target specific keys inside the secret data.
  PatchOperation kv = 4;
}
```

#### Alter annotations

```yaml
selector:
  matchPath:
    regex: "^/database/postgres/(*.)/admin_account$"
package:
  annotations:
    add:
      "infosec.elastic.co/v1/RiskManagement#leakSeverity": "critical"
    update:
      "secrets.elastic.co/v1/SecretManagement#lastRotation": {{ now | date "2006-01-02T03:04:05Z" }}
    remove:
      - "secrets.elastic.co/{{.Values.schemaVersion}}/legacyAnnotation"
    replaceKeys:
      "old-annotation-key": "new-annotation-key"
```

#### Alter labels

```yaml
selector:
  matchPath:
    regex: "^/database/postgres/(*.)/admin_account$"
package:
  labels:
    add:
      "database": "postgres"
    update:
      "database": "postgres-{{ .Values.postgresVersion }}"
    remove:
      - "database"
    replaceKeys:
      "old-label-key": "new-label-key"
```

#### Alter secret data

```yaml
selector:
  matchPath:
    regex: "^/database/postgres/(*.)/admin_account$"
package:
  data:
    kv:
      add:
        "USER": "admin-{{ randAlpha 8 }}"
      update:
        "USER": "admin-{{ randAlpha 8 }}"
      remove:
        - "PASSWORD"
      replaceKeys:
        "old-secret-key": "new-secret-key"
```

## Usage

### Apply a patch

Given this patch `postgresql-admin-rotator` :

```yaml
apiVersion: harp.elastic.co/v1
kind: BundlePatch
meta:
  name: "service-postgres-rotator"
  owner: cloud-security@elastic.co
  description: "Rotate all postgres service account password"
spec:
  rules:
  # Rule targets production and staging path
  - selector:
      matchPath:
        regex: "app/(production|staging)/{{.Values.environment}}/databases/posgresql/service_account"
    package:
      # Patch concerns secret data
      data:
        # We want to update a K/V couple
        kv:
          # Remove old keys (here generated username)
          removeKeys:
            - "admin-.*"
          # Update entry if exists
          update:
            "admin-{{ randAlpha 8 }}": "{{ strongPassword | b64enc }}"
```

Generate the input `Bundle` on which the `BundlePatch` will be applied :

```sh
$ harp from vault \
    --with-metadata \
    --paths app/production/security \
    # Apply the patch
    | harp bundle patch --spec service-postgres-rotator.yaml \
        --set environment=security \
    # Keep only target path
    | harp bundle filter \
        --keep app/production/security/databases/postgresql \
    # Create a new secret version in Vault
    | harp to vault
```

### Generate a patch from bundle difference

```sh
$ harp from vault \
    --with-metadata \
    --paths app/production/security \
    --out initial.bundle
$ harp bundle patch --spec service-postgres-rotator.yaml \
    --set environment=security \
    --out patched.bundle
$ harp bundle diff \
    --old initial.bundle \
    --new patched.bundle \
    --patch
```

---

* [Previous topic](3-template.md)
* [Index](../)
* [Next topic](5-ruleset.md)
