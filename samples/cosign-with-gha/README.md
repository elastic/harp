# Cosign key materials with GitHub Actions

## Provisioner

### Generate the private key password

```sh
$ echo -n '{{ paranoidPassword }}' | harp t > password
$ COSIGN_PASSWORD=$(cat password) cosign generate-key-pair
```

### Prepare a Harp bundle using the YAML template

With template

```yaml
product/harp/v1.0.0/artifacts/attestations/cosign/private_key/key: {{ .Files.Get "cosign.key" | toYaml }}
product/harp/v1.0.0/artifacts/attestations/cosign/private_key/password: {{ .Files.Get "password" | toYaml }}
product/harp/v1.0.0/artifacts/attestations/cosign/public_key/key: {{ .Files.Get "cosign.pub" | toYaml }}
```

Transform the YAML object as a Harp Bundle

```sh
$ harp template --in template.yaml --root . \
  | harp from object --out harp_cosign.bundle
```

### Check bundle content

```sh
$ harp bundle dump --in harp_cosign.bundle --data-only | jq
{
  "product/harp/v1.0.0/artifacts/attestations/cosign/private_key": {
    "key": "-----BEGIN ENCRYPTED COSIGN PRIVATE KEY-----\neyJrZGYiOnsibmFtZSI6InNjcnlwdCIsInBhcmFtcyI6eyJOIjozMjc2OCwiciI6\nOCwicCI6MX0sInNhbHQiOiJORWY0SzZ3VWZRWW1Ya3FySXB0WStLd3VHcnBJSkZ4\nUUNJQzA3cGVndkZrPSJ9LCJjaXBoZXIiOnsibmFtZSI6Im5hY2wvc2VjcmV0Ym94\nIiwibm9uY2UiOiJxOXlEbUNUa2t1bi9jWjV4Mzc5Vlh4WDBSTmgrc2djeSJ9LCJj\naXBoZXJ0ZXh0IjoiN1NIWjZMbkdpS0NiUUhuWG1ZamFxV0YrUzI5NXhhc3ZWOVJ4\ndzVFNXJmM0FRa2lCZVNOSVE2RlpMSExsaDNkdGhFdi82NzJ3V2VBM3dHNHBRZEx4\nUU54aWQ4b0MzRURZNkNtM3laaDRyZGhqUUNsTlVxVmhCWTlCSUxud1JsOE5nVm9s\na2RiazVHZVFyQUF1SG5uSks1RDQ1dUlUS0k0YWIzRzN3amF3WUx3azVjNXYydDhJ\najFMWjVyRlNZamNkMXBYc3Bveml0eU56RkE9PSJ9\n-----END ENCRYPTED COSIGN PRIVATE KEY-----\n",
    "password": "7cqF]\\FP0ss9uO![olGf32L+UobBCn:BqVWRZbV6=FEhv3jt>I240Ew)SYwe`ncD"
  },
  "product/harp/v1.0.0/artifacts/attestations/cosign/public_key": {
    "key": "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAESUVHp/bUkwMfoM3swWUbBgMt80o2\nx93ZquFX+D/8DBhcR+DWQSebyqsl/YEenL1GeLWmdCKSNROmvjnBV6N7uA==\n-----END PUBLIC KEY-----\n"
  }
}
```

### Publish to Vault (as cold secret storage)

```sh
$ export VAULT_ADDR=https://vault.server.local:8200
$ export VAULT_TOKEN=$(vault login -method=oidc -token-only)
$ harp to vault --in harp_cosign.bundle
```

## Publish to consumers

### Pull private key from Vault

```sh
$ export VAULT_ADDR=https://vault.server.local:8200
$ export VAULT_TOKEN=$(vault login -method=oidc -token-only)
$ harp from vault --path product/harp/v1.0.0/artifacts/attestations/cosign/private_key --out harp_cosign.bundle
```

### Inspect bundle format

```sh
$ harp bundle dump --in harp_cosign.bundle --data-only | jq
{
  "product/harp/v1.0.0/artifacts/attestations/cosign/private_key": {
    "key": "-----BEGIN ENCRYPTED COSIGN PRIVATE KEY-----\neyJrZGYiOnsibmFtZSI6InNjcnlwdCIsInBhcmFtcyI6eyJOIjozMjc2OCwiciI6\nOCwicCI6MX0sInNhbHQiOiJORWY0SzZ3VWZRWW1Ya3FySXB0WStLd3VHcnBJSkZ4\nUUNJQzA3cGVndkZrPSJ9LCJjaXBoZXIiOnsibmFtZSI6Im5hY2wvc2VjcmV0Ym94\nIiwibm9uY2UiOiJxOXlEbUNUa2t1bi9jWjV4Mzc5Vlh4WDBSTmgrc2djeSJ9LCJj\naXBoZXJ0ZXh0IjoiN1NIWjZMbkdpS0NiUUhuWG1ZamFxV0YrUzI5NXhhc3ZWOVJ4\ndzVFNXJmM0FRa2lCZVNOSVE2RlpMSExsaDNkdGhFdi82NzJ3V2VBM3dHNHBRZEx4\nUU54aWQ4b0MzRURZNkNtM3laaDRyZGhqUUNsTlVxVmhCWTlCSUxud1JsOE5nVm9s\na2RiazVHZVFyQUF1SG5uSks1RDQ1dUlUS0k0YWIzRzN3amF3WUx3azVjNXYydDhJ\najFMWjVyRlNZamNkMXBYc3Bveml0eU56RkE9PSJ9\n-----END ENCRYPTED COSIGN PRIVATE KEY-----\n",
    "password": "7cqF]\\FP0ss9uO![olGf32L+UobBCn:BqVWRZbV6=FEhv3jt>I240Ew)SYwe`ncD"
  }
}
```

We want to rename the secret keys to get more clear names on GHA side.
Let's use a `BundlePatch` for that.

```yaml
apiVersion: harp.elastic.co/v1
kind: BundlePatch
meta:
  name: "gha-secret-remapping"
  description: "Prepare secrets to be pushed to GHA"
spec:
  rules:
  - selector:
      matchPath:
        strict: product/harp/v1.0.0/artifacts/attestations/cosign/private_key
    package:
      data:
        kv:
          replaceKeys:
            "key": "COSIGN_PRIVATE_KEY"
            "password": "COSIGN_PRIVATE_KEY_PASSWORD"
```

This patch is only renaming secrets from `product/harp/v1.0.0/artifacts/attestations/cosign/private_key` path :

* `key` ⇒ `COSIN_PRIVATE_KEY`
* `password` ⇒ `COSIGN_PRIVATE_KEY_PASSWORD`

> Target names are arbitrary chosen according to our needs in GitHub pipelines

### Ensure patch application

```sh
$ harp bundle patch --spec gha-renamer.yaml --in harp_cosign.bundle
  | harp bundle dump --data-only
{
  "product/harp/v1.0.0/artifacts/attestations/cosign/private_key": {
    "COSIGN_PRIVATE_KEY_PASSWORD": "7cqF]\\FP0ss9uO![olGf32L+UobBCn:BqVWRZbV6=FEhv3jt>I240Ew)SYwe`ncD",
    "COSIGN_PRIVATE_KEY": "-----BEGIN ENCRYPTED COSIGN PRIVATE KEY-----\neyJrZGYiOnsibmFtZSI6InNjcnlwdCIsInBhcmFtcyI6eyJOIjozMjc2OCwiciI6\nOCwicCI6MX0sInNhbHQiOiJORWY0SzZ3VWZRWW1Ya3FySXB0WStLd3VHcnBJSkZ4\nUUNJQzA3cGVndkZrPSJ9LCJjaXBoZXIiOnsibmFtZSI6Im5hY2wvc2VjcmV0Ym94\nIiwibm9uY2UiOiJxOXlEbUNUa2t1bi9jWjV4Mzc5Vlh4WDBSTmgrc2djeSJ9LCJj\naXBoZXJ0ZXh0IjoiN1NIWjZMbkdpS0NiUUhuWG1ZamFxV0YrUzI5NXhhc3ZWOVJ4\ndzVFNXJmM0FRa2lCZVNOSVE2RlpMSExsaDNkdGhFdi82NzJ3V2VBM3dHNHBRZEx4\nUU54aWQ4b0MzRURZNkNtM3laaDRyZGhqUUNsTlVxVmhCWTlCSUxud1JsOE5nVm9s\na2RiazVHZVFyQUF1SG5uSks1RDQ1dUlUS0k0YWIzRzN3amF3WUx3azVjNXYydDhJ\najFMWjVyRlNZamNkMXBYc3Bveml0eU56RkE9PSJ9\n-----END ENCRYPTED COSIGN PRIVATE KEY-----\n"
  }
}
```

### Publish to GitHub Actions

```sh
$ export GITHUB_TOKEN=ghp_#############
$ harp bundle patch --spec gha-renamer.yaml --in harp_cosign.bundle
  | harp to gha \
      --owner elastic \
      --repository harp \
      --secret-filter "COSIGN_*"
```

This command will register `COSIGN_PRIVATE_KEY_PASSWORD` and `COSIGN_PRIVATE_KEY`
encrypted secrets for `elastic/harp` repository.
