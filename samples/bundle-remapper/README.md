# Secret Path Remapping

## Scenario

Cluster Admin related secrets are stored in `services/production/clusters`
path prefix in order to be CSO compliant secret must be moved in `app/production/global/clusters/1.0.0/bootstrap`
path prefix.

2 solution(s) :

- Using bundle patch (simple and portable transformation with simple logic)
- Using bundle mapper SDK (simple to complex transformation with complex logic)

### With Bundle Patch

#### Patch specification

```yaml
apiVersion: harp.elastic.co/v1
kind: BundlePatch
meta:
  name: "secret-relocator"
  description: "Move cluster secrets to CSO compliant path"
spec:
  rules:
  - selector:
      matchPath:
        regex: "^services/production/clusters/*"
    package:
      path:
        template: |-
            app/production/global/clusters/1.0.0/bootstrap/{{ trimPrefix "services/production/clusters/" .Path }}
```

#### Apply patch

```sh
$ harp bundle patch --in clusters.bundle --spec patch.yaml | harp bundle dump --path-only
app/production/global/clusters/v1.0.0/bootstrap/ap-northeast-1/<unique-32-chars-id>/users
app/production/global/clusters/v1.0.0/bootstrap/ap-northeast-1/<unique-32-chars-id>/users
app/production/global/clusters/v1.0.0/bootstrap/ap-northeast-1/<unique-32-chars-id>/users
app/production/global/clusters/v1.0.0/bootstrap/ap-southeast-1/<unique-32-chars-id>/users
...
```

#### Publish to Vault

```sh
harp b patch --in clusters.bundle --spec patch.yaml | harp to vault
```

This command does :

- Read content of `clusters.bundle`
- Apply bundle transformation from patch
- `harp` produces a new container as `stdio`
- `harp` read the container as `stdin`, decode the container and then batch
  push all secret from container to Vault.

#### One-liner

```sh
harp from vault --path services/production/global/clusters |
  harp bundle patch --spec patch.yaml |
  harp to vault
```

### Clean old secrets

> There is no `rm -rf` equivalent in Vault nor Harp. You have to use
> Harp to extract secret paths to delete and then use `vault kv metadata delete` command.

```sh
harp bundle dump --in clusters.bundle --path-only | xargs -n 1 vault kv metadata delete
```

### Bundle mapper SDK

Prepare the bundle mapper go module

```sh
go mod init harp.elastic.co/v1/remapper
```

Update dependency references

```sh
export HARP_HOME=<path to harp repository>
go mod edit -replace=github.com/elastic/harp=$HARP_HOME
```

Implement secret path mapper

<details><summary>main.go</summary>
<p>

[embedmd]:# (main.go)
```go
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

package main

import (
 "context"
 "fmt"
 "strings"

 bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
 "github.com/elastic/harp/pkg/bundle/pipeline"
 "github.com/elastic/harp/pkg/sdk/log"

 "go.uber.org/zap"
)

func main() {
 var (
  // Initialize an execution context
  ctx = context.Background()
 )

 // Run the pipeline
 if err := pipeline.Run(ctx,
  "secret-remapper",                          // Job name
  pipeline.PackageProcessor(packageRemapper), // Package processor
 ); err != nil {
  log.For(ctx).Fatal("unable to process bundle", zap.Error(err))
 }
}

// -----------------------------------------------------------------------------

func packageRemapper(ctx pipeline.Context, p *bundlev1.Package) error {

 // Remapping condition
 if !strings.HasPrefix(p.Name, "services/production/global/clusters/") {
  // Skip path remapping
  return nil
 }

 // Remap secret path
 p.Name = fmt.Sprintf("app/production/global/clusters/v1.0.0/bootstrap/%s", strings.TrimPrefix(p.Name, "services/production/global/clusters/"))

 // No error
 return nil
}
```

</p>
</details>

Compile the mapper

```sh
go build
```

#### Prepare harp bundle from Vault secrets

Login to Vault

```sh
export VAULT_ADDR=<vault-address>
export VAULT_TOKEN="$(vault login -token-only -method='oidc')"
# To reuse vault cli token
export VAULT_TOKEN="$(cat ~/.vault-token)"
```

Export `clusters` secrets

```sh
harp from vault --path services/production/global/clusters --out clusters.bundle
```

Check container paths

```sh
$ harp b dump --in clusters.bundle --path-only
services/production/global/clusters/ap-northeast-1/<unique-32-chars-id>/users
services/production/global/clusters/ap-northeast-1/<unique-32-chars-id>/users
services/production/global/clusters/ap-northeast-1/<unique-32-chars-id>/users
services/production/global/clusters/ap-southeast-1/<unique-32-chars-id>/users
...
```

#### Remap secret container

Apply bundle transformation

> Bundle SDK use stdin as input to be used in `pipe` based command pipeline

```sh
$ cat clusters.bundle | ./remapper | harp b dump --path-only
app/production/global/clusters/1.0.0/bootstrap/ap-northeast-1/<unique-32-chars-id>/users
app/production/global/clusters/1.0.0/bootstrap/ap-northeast-1/<unique-32-chars-id>/users
app/production/global/clusters/1.0.0/bootstrap/ap-northeast-1/<unique-32-chars-id>/users
app/production/global/clusters/1.0.0/bootstrap/ap-southeast-1/<unique-32-chars-id>/users
...
```

#### Publish to Vault

```sh
cat clusters.bundle | ./remapper | harp to vault
```

This command does :

- Read content of `clusters.bundle`
- Write as `stdin` to `remapper`
- `remapper` decode input as a harp bundle and apply transformation
- `remapper` produces a new container as `stdio`
- `harp` read the container as `stdin`, decode the container and then batch
  push all secret from container to Vault.

#### One-liner

```sh
harp from vault --path services/production/global/clusters | ./remapper | harp to vault
```
