# Features overview

- [Glossary](#glossary)
- [Bundle management](#bundle-management)
- [Features](#features)
- [Pipelines](#pipelines)
- [Usages](#usages)
  - [Secret Container](#secret-container)
    - [Seal a secret container](#seal-a-secret-container)
      - [Create an identity](#create-an-identity)
        - [Use a passphrase as private key protection](#use-a-passphrase-as-private-key-protection)
        - [Use Vault in\-transit key to encrypt private key](#use-vault-in-transit-key-to-encrypt-private-key)
      - [Ephemeral Container Key](#ephemeral-container-key)
      - [Deterministic Container Key](#deterministic-container-key)
      - [Recover a container key from indentity](#recover-a-container-key-from-indentity)
    - [Unseal a secret container](#unseal-a-secret-container)
  - [Secret Bundle](#secret-bundle)
    - [Create a bundle from template](#create-a-bundle-from-template)
    - [Create a bundle from a JSON map](#create-a-bundle-from-a-json-map)
    - [Read a secret value](#read-a-secret-value)
      - [Example](#example)
    - [Calculate a bundle difference](#calculate-a-bundle-difference)
    - [Patch a bundle](#patch-a-bundle)
    - [Dump a secret bundle](#dump-a-secret-bundle)
    - [Import a JSON bundle](#import-a-json-bundle)
    - [Encrypt secret values](#encrypt-secret-values)
    - [Decrypt secret values](#decrypt-secret-values)
  - [Vault specific commands](#vault-specific-commands)
    - [Export a complete secret backend from Vault](#export-a-complete-secret-backend-from-vault)
    - [Import a bundle in a target secret backend in Vault](#import-a-bundle-in-a-target-secret-backend-in-vault)

## Glossary

- `secret` is a tuple of information identified by a `key` holding a `value`
    characterized by an array of bytes.
- `package` is a collection of `secret` kv pair, it is identified by a name
    usually called the secret path;
- `bundle` is a collection of `package` items, and form an atomic group of
    information representing a secret value state.
- `container/crate` is the file used to securely store the `bundle`.
- `secret operator` is a role assigned to an identity allowed to manage secrets;
- `secret consumer`is a role assigned to an identity allowed to read secrets with
    authorized policy;
- `secret backend` is a technical component used to store and organize secrets;

## Bundle management

harp is used by `secret-operators` to manage and produce secret
bundles. It implements secret data management pipelines, to make it auditable
and reproductible.

![Secret management Pipeline](docs/harp/img/SM-HARP-PIPELINE.png)

### Features

- Input(s)

  - Read
    - `Hashicorp Vault`
    - JSON map
    - Secret container dump
  - Generate
    - `BundleTemplate` for secret bootstrap

- Ouput(s)
  - `Hashicorp Vault`

### Pipelines

`harp` allows you to handle secret using deterministic pipelines expressed
using a serie of atomic cli operations.

![Pipelines](docs/harp/img/SM-HARP.png)

> The main objective is to reach as soon as possible the harp native
> container to be used by the harp core cli.
> If you need to pull or push secret from / to external secret storage engine,
> just use the SDK du generate a harp plugin to pull secret and store
> them as a harp container.

## Usages

### Secret Container

#### Seal a secret container

##### Create an identity

Identities are cryptographic keypairs (Curve25519) used for sealing process.

The secret container allows sealing to use multiple identities (public keys) during
the process so that these identities matching private keys could be used to unseal
the secret container.

###### Use a passphrase as private key protection

Generate a passphrase first. This passphrase will be used to encrypt the
private key of the identity.

```sh
harp passphrase > passphrase.txt
```

> Passphrase must be stored and permissionned effeciently in your secret storage.

Create a recovery identity :

```sh
harp container identity \
    --passphrase $(cat passphrase.txt) \
    --description "Recovery" \
    --out recovery.json
```

Sample identity

```json
{
    "@apiVersion": "harp.elastic.co/v1",
    "@kind": "ContainerIdentity",
    "@timestamp": "2020-10-27T19:56:47.47957Z",
    "@description": "Recovery",
    "public": "KWgAQdzWeKo8xvClMM_rY_NgNq9wCsQsODxSNtqUJV4",
    "private": {
        "encoding": "jwe",
        "content": "eyJhbGciOiJQQkVTMi1IUzUxMitBMjU2S1ciLCJjdHkiOiJqd2sranNvbiIsImVuYyI6IkEyNTZHQ00iLCJwMmMiOjUwMDAwMSwicDJzIjoiYmtFNE9GaFljV3MwTTFobGMxcERSQSJ9.gzyUTwV9QjP6bic22QkQlqmwAeGhnubDFEFiruRP2x1xL84K90teSw.n4_O17H720MKw972.CipjKnrp4Mb5Hgdy1xcUcEHh3ioyg0UbGXd_RneFiopTq0qUm1u4aymq9fSabdrRiKlW3LjGy71MpF8kd_0SD1-5Qzg13NId8oTRI1hnfZDKBrN_nZMPYIOnACqSTB7kjywHdZS0rhHDCVQUtYUy5JAkiw-rRge6ShOlidwNshI.xCc_XAaEPjJIGc9Ucb1asg"
    }
}
```

> Recovery identity can be "publicly" stored.

###### Use Vault in-transit key to encrypt private key

> This will remove the passphrase usage, and transform the permission to unseal
> a container by adding/removing permission to use the transit backend key.

Create a Vault in-transit backend key first.

```sh
$ vault secrets enable transit
Success! Enabled the transit secrets engine at: transit/
```

Create a transit encryption key identity :

```sh
$ vault write -f transit/keys/harp
Success! Data written to: transit/keys/harp
```

Create a recovery identity :

```sh
$ export VAULT_ADDR=...
$ export VAULT_TOKEN=...
$ harp container identity \
    --vault-transit-key harp \
    --description "Recovery" \
    --out recovery.json
```

Sample identity

```json
{
    "@apiVersion": "harp.elastic.co/v1",
    "@kind": "ContainerIdentity",
    "@timestamp": "2020-10-27T19:58:51.398987Z",
    "@description": "Recovery",
    "public": "4hHPpiJMnVhQFlnveRKeCdaPoqHzW74Rro0S1X33QS4",
    "private": {
        "encoding": "kms:vault:q6pcgHWM6wSJWG6OYmHM97DMMeerqTXExYAolfhn4N8",
        "content": "vault:v1:CNMnI9sIRYYD6pRl1TQ3KHHO+JCmfZiAtD+XBnnIxHt4F6CeFYmuUtY6k7+XAMxtAWG5NtLgS0uyPken2ef1ihJ6Pf6DOtlgDUCnDKyVEvGeZcdOdaZHTgc3YIX/wDY9odmtUvjJGaPNMtIADtMPcjkOLgZH3FnF701dJcsKPxr1fqTQd8mCiFFWqWF9kOQYMqf/1yBybcY6XOGI"
    }
}
```

##### Ephemeral Container Key

For immutability principle, the sealing process generates a new Container Key
at each execution. It means that all the container consumers must know the new
container to be able to unseal it.

In order to `seal` a secret container, you can use the following commands :

Seal the container using the generated passphrase for recovery :

```sh
$ harp container seal \
    --identity $(cat recovery.json | jq -r ".public") \
    --in unsealed.container \
    --out sealed.container
Container Key: .....
```

##### Deterministic Container Key

Seal the container using a deterministic container key derived from a master key.
This will prevent modification of container consumers after each container seals.

Generate a master key :

> Keep this key as an high sensitive secret.

```sh
harp keygen master-key > master.key
```

Seal the secret container using deterministic container key derivation (DCKD) :

```sh
$ harp container seal \
    --identity $(cat recovery.json | jq -r ".public") \
    --dckd-master-key $(cat master.key) \
    --dckd-target "customer-1:release-XXX:2020-10-31" \
    --in unsealed.container \
    --out sealed.container
Container key : ....
```

- The `dckd-master-key` flag defines the root key used for derivation.
- The `dckd-target` flag defines an arbitry string acting as a salt for Key
    Derivation Function.

#### Recover a container key from indentity

When the container key is lost, you can use attached one of identity private keys
to unseal the container.

For passphrase recovery :

```sh
$ harp container recover --identity recovery.json --passphrase $(cat passphrase.txt)
Container key : mPjzX1A5PcGtZ0nacxkhjl0pZE8XYw84KYF5NO6jhVA
```

Fo Vault recovery :

```sh
harp container recover --vault-transit-key harp --identity recovery.json
Container key : VyEJ6lMy7CPOjJnPYMjH-M7uWUym5utYo4JDVNPPMc8
```

#### Unseal a secret container

In order to modify a bundle, this bundle need to be unsealed.

```sh
$ harp container seal --in sealed.bundle --out secret.bundle
Enter container key:
```

### Secret Bundle

#### Create a bundle from template

You have to create a `BundleTemplate` that will contains all secret generation
specification.
This specification is embedded in the bundle so that it will be used for secret
rotation based on the specification.

Given this YAML specificcation :

<details><summary>infra.yaml</summary>
<p>

```yaml
apiVersion: harp.elastic.co/v1
kind: BundleTemplate
meta:
    name: "Ecebootstrap"
    owner: syseng@elstc.co
    description: "ECE Secret Bootstrap"
spec:
    selector:
        product: "ece"
        version: "v1.0.0"
    namespaces:
        infrastructure:
            - provider: "aws"
              description: "ESSP AWS Account"
              regions:
                  - name: "us-east-1"
                    services:
                        - type: "rds"
                          name: "adminconsole"
                          description: "PostgreSQL Database used for AdminConsole storage backend"
                          secrets:
                              - suffix: "accounts/root_credentials"
                                description: "Root Admin Account"
                                template: |
                                    {
                                      "user": "{{.Provider}}-{{.Account}}-{{.Region}}-dbroot-{{ randAlphaNum 8 }}",
                                      "password": "{{ paranoidPassword | b64enc }}"
                                    }

                        - type: "mail"
                          name: "mailgun"
                          description: "Mailgun encryption keys"
                          secrets:
                              - suffix: "security/signature_keys"
                                description: "Signature keys"
                                template: |
                                    {
                                        "privateKey": "{{ $sigKey := cryptoPair "rsa" }}{{ $sigKey.Private | jwk | b64enc }}",
                                        "publicKey": "{{ $sigKey.Public | jwk | b64enc }}"
                                    }
                              - suffix: "security/encryption_keys"
                                description: "Encryption keys"
                                template: |
                                    {
                                        "encryptionKey": "{{ cryptoKey "aes:256" }}"
                                    }
```

</p>
</details>

```sh
harp from template --in spec.yaml --out infra.bundle
```

#### Create a bundle from a JSON map

You can create a `Bundle` using a json map.

The JSON input must match the following format.

```json
{
    "path-1": {
        "key1": "value1"
    },
    "path-2": {
        "key1": "value1"
    }
}
```

You can generate a `Bundle` using the following command :

```sh
cat << EOF | harp from jsonmap --out json.bundle
{
  "platform/production/customer-1/us-east-1/billing/recurly/vendor_api_key": {
    "API_KEY": "recurly-foo-api-123456789"
  },
  "platform/production/customer-1/us-east-1/postgresql/admiconsole/admin_credentials": {
    "password": "KmZkZFBXaH5IYzNlR1FFYnxXZVJCUTBvIzBOM0JAITFtYk1malhCRVZoSixmNSxNZXZwVWVnUjpMMDg3QGdFOQ==",
    "username": "dbadmin-uKj9BJGO"
  },
  "platform/production/customer-1/us-east-1/zookeeper/accounts/admin_credentials": {
    "password": "Sm55Vnthb3QxMXpESTVtKHVvOEE5aEVVczhEb2gqWnhTP2VuQzc1bFZ7eVlGL1A2YHJPfX5pMUZoS2lsLnJzQg==",
    "username": "zkadmin-DDrEQA8i"
  }
}
EOF
```

#### Read a secret value

Read a secret value at a given `path`, and optionally extract only the given `field`.

```sh
harp bundle read --in unsealed.bundle \
    --path <path> \
    --field <field>
```

> With dump and jq

```sh
harp bundle dump --in unsealed.bundle --content-only | jq -r '.["<path>"].<field>' | jq
```

##### Example

```sh
harp bundle dump --in unsealed.bundle --content-only | jq -r '.["infra/aws/security/global/ec2/default/ssh/ed25519_keys"].privateKey'
```

Is equivalent to

```sh
harp bundle read --in unsealed.bundle \
    --path "infra/aws/security/global/ec2/default/ssh/ed25519_keys" \
    --field privateKey
```

#### Calculate a bundle difference

This is used to generate a diff report from 2 bundles.

```sh
$ harp bundle diff --src input.bundle --dest other.bundle
...
    "infra/aws/customer-1/us-east-1/rds/adminconsole/accounts/root_credentials": bundle.KV{
-       "password": string("YnFkXldoR3E9Z2lJVy5Hc0FeMDhHIUs5eDpHR0E0VTVuaG9JZkRLUW1hN08rNFoyQltKRnwwTDlXQV1lRXRiUA=="),
+       "password": string("e09+Sjg5UUhWWWFTWC4zWElWaXljQXtlV1ZtNG1PQC9ZZDJWelt5fGFLInNPWXZGMFU1TW45NUhPazQ+TkZYcg=="),
        "user":     string("dbroot-w9NinCPl"),
    },
...
```

#### Patch a bundle

It uses a specification to apply tranformations to the given bundle.

> Apply transformation according to a strict path matching selector

<details><summary>postgresql-rotator.yaml</summary>
<p>

```yaml
apiVersion: harp.elastic.co/v1
kind: BundlePatch
meta:
    name: "postgresql-rotator"
    owner: cloud-security@elastic.co
    description: "Rotate postgresql password"
spec:
    rules:
        # Target a precise secret
        - selector:
              matchPath:
                  # Strict match
                  strict: "platform/{{.Values.quality}}/{{.Values.account}}/{{.Values.region}}/postgresql/{{.Values.component}}/admin_credentials"

          # Apply this operation on selector matches
          package:
              # Access data
              data:
                  # Target an explicit keys only
                  kv:
                      remove: ["port"]
                      add:
                          "listener": "5432"
                      update:
                          "username": "dbuser-{{.Values.component}}-{{ randAlphaNum 8 }}"
                          "password": "{{ paranoidPassword | b64enc }}"
```

</p>
</details>

> Apply transformation according to a regex path matching selector

<details><summary>secret-service-fernet-rotator.yaml</summary>
<p>

```yaml
apiVersion: harp.elastic.co/v1
kind: BundlePatch
meta:
    name: "fernet-key-rotator"
    owner: cloud-security@elastic.co
    description: "Rotate or create all fernet key of given bundle"
spec:
    rules:
        # Object selector
        - selector:
              # Package path match this regexp
              matchPath:
                  # Regex match
                  regex: ".*"

          # Apply this operation
          package:
              # On package annotation
              annotations:
                  # Update annotation value with new secret
                  update:
                      secret-service.elstc.co/encryptionKey: |-
                          {{ cryptoKey "fernet" }}

              # On package data
              data:
                  # Update annotations
                  annotations:
                      # Update annotation value with new secret
                      update:
                          secret-service.elstc.co/encryptionKey: |-
                              {{ cryptoKey "fernet" }}
```

</p>
</details>

```sh
harp from vault \
    --path platform/production/customer-1/us-east-1/postgresql/admiconsole/admin_credentials \
    | harp bundle patch --spec postgresql-rotator.yaml \
        --set quality=production \
        --set account=customer-1 \
        --set region=us-east-1 \
        --set component=adminconsole | harp to vault
```

This will pull the given value from vault as a bundle, rotate the targeted
secret according to values and secret path built with them, and then publish
the bundle back to vault.

#### Dump a secret bundle

If you need to inspect internal representation of the bundle, you could use
`dump` action. This will read the bundle and produce a JSON as output.

> This output is compatible with `from dump` command

```sh
$ harp bundle dump --in input.bundle | jq
{
    "annotations": {...},
    "labels": {...},
    "version": 1,
    "packages": [
        {
            "annotations": {...},
            "labels": {...},
            "name": "secrets/database/postgres.yml",
            "secrets": {
                "annotations": {...},
                "labels": {...},
                "data": [
                    {
                        "key": "URL",
                        "type": "string",
                        "value": "<base64>"
                    }
                ],
                "versions": {...}
            }
        }
    ]
    "template": {...}
}
```

If you want to export the bundle content, for additionnal operations (jq, diff, etc.)
you can add the `--content-only` flag, it will export the build as a JSON map.

> This output is **not** compatible with `load` command

```sh
$ harp bundle dump --content-only | jq
{
  "platform/production/customer-1/us-east-1/billing/recurly/vendor_api_key": {
    "API_KEY": "recurly-foo-api-123456789"
  },
  "platform/production/customer-1/us-east-1/postgresql/admiconsole/admin_credentials": {
    "password": "KmZkZFBXaH5IYzNlR1FFYnxXZVJCUTBvIzBOM0JAITFtYk1malhCRVZoSixmNSxNZXZwVWVnUjpMMDg3QGdFOQ==",
    "username": "dbadmin-uKj9BJGO"
  },
  "platform/production/customer-1/us-east-1/zookeeper/accounts/admin_credentials": {
    "password": "Sm55Vnthb3QxMXpESTVtKHVvOEE5aEVVczhEb2gqWnhTP2VuQzc1bFZ7eVlGL1A2YHJPfX5pMUZoS2lsLnJzQg==",
    "username": "zkadmin-DDrEQA8i"
  }
}
```

If you want to export the bundle secret paths.

```sh
$ harp bundle dump --path-only
platform/production/customer-1/us-east-1/billing/recurly/vendor_api_key
platform/production/customer-1/us-east-1/postgresql/admiconsole/admin_credentials
platform/production/customer-1/us-east-1/zookeeper/accounts/admin_credentials
```

#### Import a JSON bundle

Sometimes, you need to process secret bundle before using it for example :

- Secret rotation
- Bundle modifications
- Namespace or Package remapping

For that, you need to dump the secret bundle, process the JSON using your tool
or language, and then reimport the bundle to generate the binary one.

```sh
harp bundle dump --in input.bundle | ./remap.py | harp from dump --out remapped.bundle
```

#### Encrypt secret values

In order to protect you unsealed bundle for confidentiality requirements, you
can encrypt secret values.

Supported encryption:

- `aes256-gcm96`
- `secretbox`
- `fernet`

For this purpose, you have to generate a key using `keygen` subcommands.

```sh
$ harp keygen secretbox
secretbox:Vm1xW_Tp6coVww2SRCWBIR3fh77-oZefXsJiuG02LNw=
```

Then apply tranformation to secret bundle

```sh
harp bundle encrypt --in unsealed.bundle --out encrypted.bundle \
    --key secretbox:Vm1xW_Tp6coVww2SRCWBIR3fh77-oZefXsJiuG02LNw=
```

You can still read all keys of the bundle, but secret values attached are now
`locked`.

```sh
$ harp bundle dump --in encrypted.bundle | jq
{
    "packages": [
        {
            "namespace":"foo",
            "name": "secrets/database/postgres.yml",
            "locked": "<base64>"
        }
    ]
    "template": {...}
}
```

#### Decrypt secret values

```sh
harp bundle decrypt --in encrypted.bundle --out decrypted.bundle \
    --key secretbox:Vm1xW_Tp6coVww2SRCWBIR3fh77-oZefXsJiuG02LNw=
```

#### Vault specific commands

##### Export a complete secret backend from Vault

> Only secrets visible to you will be exported. Also this a CPU/Network intensive
> operation, be aware of it.

This will be used to export all secrets from a Vault secret engine K/V backend in
a unsealed bundle.

```sh
harp from vault \
    --path infra \
    --path platform \
    --path product \
    --path app \
    --path artifact \
    --out vault-backup.bundle
```

Or you can pass path via `stdin` or `file` with `--pass-from` flag:

```sh
harp bundle dump --path-only --in vault.bundle \
    | harp from vault \
        --out imported.bundle \
        --paths-from -
```

##### Import a bundle in a target secret backend in Vault

This will be used to import an unsealed bundle into a given Vault K/V backend path.

```sh
harp to vault \
    --in infra.bundle \
    --prefix legacy
```
