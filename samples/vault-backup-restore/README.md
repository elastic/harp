# Vault Secret Backup / Restore

Use harp to export all KV backend secret values and save it as a bundle.

> This method is used for Vault secret backend migration.

## Scenario

### Prepare backup plan

`harp` can crawl and export secret path for KV secret backends (`v1` and
`v2`).

You don't have to specify the version used, `harp` detect and set up an
internal service adapted to secret backend.

You don't need to adapt your path according to KV backend neither, just specify
the secret path, and it will be translated according to the target secret backend
 version.

### Backup

#### Export secrets as a bundle

```sh
harp from vault --out backup.bundle \
        --path meta \
        --path infra \
        --path platform \
        --path product \
        --path app \
        --path artifact
```

#### List exported secrets

```sh
$ harp bundle dump --in backup.bundle --path-only
meta/cso
...
```

#### Seal the bundle

In order to save the bundle in a dedicated storage, you have to seal it first to
protect secret values.

```sh
harp container seal \
  --identity-file backup.json \
  --in backup.bundle \
  --out backup-$(date +%Y-%m-%d).bundle
```

#### Save in your backup storage

```sh
aws s3 cp backup-$(date +%Y-%m-%d).bundle* s3://mybackupbucket/
```

### Restore secrets from bundle

#### Retrieve bundle

Download the bundle

```sh
aws s3 cp s3://mybackupbucket/backup-YYYY-MM-DD.bundle .
```

#### Unseal the bundle

Recover the container key from passphrase

```sh
$ harp container recover --in backup.json
...
Container key : B6c8EIkjs5nr_8sKZ3iX1gMrt-Mudg8tEIQQ8LDGBG4
```

Unseal the bundle

```sh
harp container unseal --in backup-YYYY-MM-DD.bundle --out restore.bundle
```

#### Import bundle to Vault

Prepare Vault authentication context

```sh
export VAULT_ADDR=https://...
export VAULT_TOKEN=$(vault login -method=oidc -token-only)
```

Restore the bundle

```sh
harp to vault --in restore.bundle
```
