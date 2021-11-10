# Cluster users management

## Scenario

Cluster Admin team handles `cluster service accounts` to authenticate tools that
consumer or produce data stored in Elasticsearch.

### Secret provisioning

#### New cluster

We are going to establish the list of `serviceAccount` need for all clusters
and maintain it as a static value file, used by `harp` template engine to
generate a secret values.

[embedmd]:# (service-accounts.yaml)
```yaml
serviceAccounts:
  - "sa-account-1"
  - "sa-account-2"
  - "sa-account-3"
  - "sa-account-4"

```

Let's now describe the `new cluster secret set` as a `BundleTemplate` to be
used on new cluster instance bootstrap and will be used to provision Vault.

[embedmd]:# (bootstrap-cluster.yaml)
```yaml
apiVersion: harp.elastic.co/v1
kind: BundleTemplate

meta:
  name: "cluster-service-accounts"
  owner: cluster-manager@elastic.co
  description: "Cluster Service Account provisioner"

spec:
  selector:
    quality: "{{ .Values.quality }}"
    platform: "observability"
    product: "deployer"
    version: "1.0.0"

  namespaces:
    application:
    - name: "clusters"
      description: "Managed clusters secrets"
      secrets:
      - suffix: "{{ .Values.installation }}/{{ .Values.region }}/{{ .Values.clusterid }}/users"
        template: |-
          [[- $userMap := dict ]][[ range $user := .Values.serviceAccounts -]]
          [[- $_ := set $userMap $user "{{ noSymbolPassword }}" ]][[ end ]]
          [[ $userMap | toJson ]]
```

This template defines external `template values` :

* `quality` to define target environment quality (`production`, `staging`, `qa`, `dev`)
* `installation` defines a cluster instance group
* `region` defines the logical region used to deploy the cluster
* `clusterid` defines the cluster ID

To generate a `secret container` from this specification :

```sh
harp template \
  --alt-delims \                      # Use alternative delimiters '[[ ]]'
  --in bootstrap-cluster.yaml \       # Input bundle specification
  --values service-accounts.yaml      # service account list
```

So execution of this command, will produce a YAML file with a value generation
template function `{{ noSymbolPassword }}` for each `service account` :

```sh
harp template --alt-delims --in bootstrap-cluster.yaml --values service-accounts.yaml > bootstrap-compiled.yaml
```

[embedmd]:# (bootstrap-compiled.yaml)
```yaml
apiVersion: harp.elastic.co/v1
kind: BundleTemplate

meta:
  name: "cluster-service-accounts"
  owner: cluster-manager@elastic.co
  description: "Cluster Service Account provisioner"

spec:
  selector:
    quality: "{{ .Values.quality }}"
    platform: "observability"
    product: "deployer"
    version: "1.0.0"

  namespaces:
    application:
    - name: "clusters"
      description: "Managed clusters secrets"
      secrets:
      - suffix: "{{ .Values.installation }}/{{ .Values.region }}/{{ .Values.clusterid }}/users"
        template: |-
          {"sa-account-1":"{{ noSymbolPassword }}","sa-account-2":"{{ noSymbolPassword }}","sa-account-3":"{{ noSymbolPassword }}","sa-account-4":"{{ noSymbolPassword }}"}
```

Now if we compile the `secret container` using the generated `BundleTemplate` :

```sh
$ harp from template --in bootstrap-compiled.yaml
  --set quality=production \                         # set quality value
  --set installation=customer1 \                     # set installation value
  --set region=us-east-1 \                           # set region value
  --set clusterid=1234567894561234567898978978900  \ # set clusterid value--out gcm.bundle
  --out cluster.bundle                               # Secret container name
```

It will produce a `secret container` with content :

```sh
$ harp bundle dump --in gcm.bundle --content-only | jq
{
  "app/production/observability/deployer/1.0.0/clusters/customer1/us-east-1/1234567894561234567898978978900/users": {
    "sa-account-1": "8ueVz6uOH6s8e0pY90OhNN66I3NYNTwI",
    "sa-account-2": "0csYsH5o7YJ89sk8HgUwyXT2j9w6I0Sh",
    "sa-account-3": "5jUmm81Uz11K0HnJHa3zSqDU68ev8UYj",
    "sa-account-4": "3kvD2JnB633M9H4W4oMSwuBeFq55IAXm"
  }
}
```

You can now publish this new bundle in Vault by doing the following command :

```sh
export VAULT_ADDR=<vault-url>
export VAULT_TOKEN="$(vault login -token-only -method='oidc')"
harp to vault --in cluster.bundle
```

#### New service account

You have to deploy a new `service account` on all clusters.

First you need to generate a `secret container` from Vault.

```sh
export VAULT_ADDR=<vault-url>
export VAULT_TOKEN="$(vault login -token-only -method='oidc')"
harp from vault
  --path app/production/observability/deployer/1.0.0/clusters
  --path app/staging/observability/deployer/1.0.0/clusters
  --out cluster.bundle
```

Now you have a local copy of all cluster service accounts.

In order to batch patch the `secret container` you have to describe your patch
using a `BundlePatch` specification file.

[embedmd]:# (service-account-deployer.yaml)
```yaml
apiVersion: harp.elastic.co/v1
kind: BundlePatch
meta:
  name: "service-account-deployer"
  owner: cluster-admin@elastic.co
  description: "Add a new cluster service account"
spec:
  rules:
  # Rule targets production and staging path
  - selector:
      matchPath:
        regex: "app/(production|staging)/observability/deployer/1.0.0/clusters/.*/.*/[0-9a-z]{32}/users"
    package:
      # Patch concerns secret data
      data:
        # We want to update a K/V couple
        kv:
          # Add an entry
          add:
            "{{ .Values.serviceAccountName }}": "{{ noSymbolPassword }}"
```

This patch will add a new `service account` in the `users` JSON map for all secrets
that match the given path.

To apply this patch :

```sh
harp bundle patch
  --in cluster.bundle \ # To specify the source secret container where the patch will be applied
  --spec service-account-deployer.yaml \ # Path to patch specification
  --set serviceAccountName=sa-account-5 \ # Set the serviceAccountName value
  --out cluster-patched.bundle
```

It will add a `sa-account-5` key to `users` :

```diff
$ harp bundle dump --in gcm.bundle --content-only | jq
{
  "app/production/observability/deployer/1.0.0/clusters/customer1/us-east-1/1234567894561234567898978978900/users": {
    "sa-account-1": "8ueVz6uOH6s8e0pY90OhNN66I3NYNTwI",
    "sa-account-2": "0csYsH5o7YJ89sk8HgUwyXT2j9w6I0Sh",
    "sa-account-3": "5jUmm81Uz11K0HnJHa3zSqDU68ev8UYj",
    "sa-account-4": "3kvD2JnB633M9H4W4oMSwuBeFq55IAXm",
+    "sa-account-5": "C54Qw2Qe74S7u7R2jpz4ZKpSI7progDR"
  }
}
```

In order to publish changes in Vault :

```sh
harp to vault
  --in cluster-patched.bundle
```

#### Service account rotation

If you want to rotate a `service account` credentials, it's exactly the same
approach using a `BundlePatch`, but this time you update the matching `service account` name value.

[embedmd]:# (service-account-rotator.yaml)
```yaml
apiVersion: harp.elastic.co/v1
kind: BundlePatch
meta:
  name: "service-account-rotator"
  owner: cluster-admin@elastic.co
  description: "Rotate cluster service account password"
spec:
  rules:
  # Rule targets production and staging path
  - selector:
      matchPath:
        regex: "app/(production|staging)/observability/deployer/1.0.0/clusters/.*/.*/[0-9a-z]{32}/users"
    package:
      # Patch concerns secret data
      data:
        # We want to update a K/V couple
        kv:
          # Update entry if exists
          update:
            "{{ .Values.serviceAccountName }}": "{{ noSymbolPassword }}"
```

Execute the following command :

```sh
harp from vault
  --path app/production/observability/deployer/1.0.0/clusters/customer1/us-east-1/1234567894561234567898978978900/users
  | harp bundle patch --spec service-account-rotator.yaml --set serviceAccountName=sa-account-2
  | harp to vault
```
