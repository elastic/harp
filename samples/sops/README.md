# Create a secret container from a structured file

- [Create a secret container from a structured file](#create-a-secret-container-from-a-structured-file)
  - [Secret template](#secret-template)
    - [Sample output](#sample-output)
  - [Deploy secrets](#deploy-secrets)
    - [Hashicorp Vault](#hashicorp-vault)
      - [Application access profile](#application-access-profile)
    - [Mozilla SOPS](#mozilla-sops)
      - [Create the identity](#create-the-identity)
      - [Seal using sops](#seal-using-sops)
      - [Retrieve a bundle from a sops encrypted file](#retrieve-a-bundle-from-a-sops-encrypted-file)
      - [Generate a YAML output from a bundle](#generate-a-yaml-output-from-a-bundle)
    - [Complete scenario](#complete-scenario)

## Secret template

Create your secret template from scratch. This will define your secret specification
used by the secret consumer.

```yaml
app:
    {{ .Values.env }}:
        database:
            user: app-{{ randAlpha 16 }}
            password: {{ paranoidPassword }}
        server:
            privacy:
                principal: {{ paranoidPassword | b64enc }}
            http:
                session:
                    cookieKeyB64: {{ paranoidPassword | b64enc }}
                token:
                    signingPrivateKeyJWK: |-
                        {{ $key := cryptoPair "ec:p384" }}{{ $key.Private | toJwk }}
                    signingPublicKeyJWK: |-
                        {{ $key.Public | toJwk }}
        vendor:
            mailgun:
                apiKey: {{ .Values.mailgun.apikey | quote }}
```

> When designing a secret tree, please consider the authorizations applicable to
> the secrets. Don't expose too many secrets just to make it easy to handle and
> on the other hand don't split secrets to make it too atomic. Try to bundle
> them in package where the authorization policy will be applied.

| Secret Path | Description | Specification | Cardinality / Rotation Period |
| ----------- | ----------- | ------------- | ----------------------------- |
| `app/{{ env }}/database/user` | Defines the applicative user used by your application to manage service data. | Has the `app-` prefix to identify `app` service accounts, and a random alphanumeric 16 characters value as discriminent to handle multi instances of the same service. | One per instace / 90 days |
| `app/{{ env }}/database/password` | Defines the password used to authenticate to application database identity. | 64 Printable ASCII characters. | One per instance / 7 days |
| `app/{{ env }}/server/privacy/principal` | Defines the seed used by cryptographic function to anonymize the principal. | Standard Base64 encoded 64 Printable ASCII characters. | One per environment / No rotation |
| `app/{{ env }}/server/http/session/cookieKeyB64` | Defines the encryption key used by cookie encryption function. | Standard Base64 encoded 64 Printable ASCII characters. | One per environment / 30 days |
| `app/{{ env }}/server/token/signingPrivateKeyJWK` | Defines the JWT Token signing key encoded using JWK. | EC P384 Private Key encoded using JWK. | One per environment / 24 hours |
| `app/{{ env }}/server/token/signingPublicKeyJWK` | Defines the JWT Token verification public key encoded using JWK. | EC P384 Public Key encoded using JWK. | One per environment / 24 hours |
| `app/{{ env }}/server/vendor/mailgun/apiKey` | Defines the Mailgun API Key used by the service. | Imported value during the rendering. | One per environment / 180 days |

### Sample output

```sh
$ harp template --in template.yaml --set env=production --set mailgun.apikey=1234567890
app:
    production:
        database:
            user: app-tNtKnOgYXuuJyArI
            password: (Pzd~O9JQ"Epm)QN-eWZAscJHxjyfQlXXq_t=HwEc1it3Bu9h+A7um777q"t4n*8
        server:
            privacy:
                principal: R0FLZVROZDZlMGRUPFYsQXJRWFluVGlJYUFCTDBIT3p5a145TTRAZjNsc0B6XnhmKGo0djU4Y2AzY01vaDolcA==
            http:
                session:
                    cookieKeyB64: aV5kU292VTYvTlhJM1BrdDgzSEJUcXgrOSVIfkU2Tlp2azN4YXBhWlEpLmJjZiNkZz96VyVvQ0piN3YwbHRpNQ==
                token:
                    signingPrivateKeyJWK: |-
                        {"kty":"EC","kid":"sAt9gTz5oUmTefVK0Ib1OfRv7uKVC1at9bB1VjR52QE=","crv":"P-384","x":"XeCZRaJDh4i1ywansUMAh2kN6WbEqWNnQc0diC0SkVxmCAcxA69PQKbyYJy49z9Y","y":"FuDWJpFuuYX0JDRSJTQ5uexLCU3-G4tGEGAoRtWQLRCZIpz4tfcd87bFiIhAb4MT","d":"gHmrbNC5X8rp4ORxwQkdeqMctMjP_jLO3ngd_oxp-mclK3O7TIpPM0dm7TEx0AjR"}
                    signingPublicKeyJWK: |-
                        {"kty":"EC","kid":"sAt9gTz5oUmTefVK0Ib1OfRv7uKVC1at9bB1VjR52QE=","crv":"P-384","x":"XeCZRaJDh4i1ywansUMAh2kN6WbEqWNnQc0diC0SkVxmCAcxA69PQKbyYJy49z9Y","y":"FuDWJpFuuYX0JDRSJTQ5uexLCU3-G4tGEGAoRtWQLRCZIpz4tfcd87bFiIhAb4MT"}
        vendor:
            mailgun:
                apiKey: "1234567890"
```

## Deploy secrets

### Hashicorp Vault

```sh
$ harp from object --in production.yaml \
  | harp bundle dump --data-only \
  | jq
{
  "app/production/database": {
    "password": "c7#kaTKgAqjAuC6KKdAz)_tsa9mBetD83lAk2cSvyH.jB0AeW%/m4g~4Vi]|I7x!",
    "user": "app-xhyXgOEzYaPmFtYS"
  },
  "app/production/server/http/session": {
    "cookieKeyB64": "fHJ1N1Z6YWI5RSIuWm1VNHZCTVB4NGdjSCVmJSFseWNwcjVRT2gwWC95TSlmNVVPfTc5X2tmaHp4TVUxSHR0dQ=="
  },
  "app/production/server/http/token": {
    "signingPrivateKeyJWK": "{\"kty\":\"EC\",\"kid\":\"m5xr81U_PIKPLd0uzi6QH_kefkMoiN3fF5OUR34cF5I=\",\"crv\":\"P-384\",\"x\":\"nqrvtZkR8WxH_1URu3rDkd-FtBysDSjfmyUwctEV5xpWH_O71rOCsZDXltB5wBtq\",\"y\":\"0tqE2Atw9OncJsi0gGNw5k2Ay5u3xfCJlVZKPSDjKgqWIOpVYDVFqFN6IFu7qIgX\",\"d\":\"Ln4mPosjcCXRXq7TcKEVoWS8b2kfpCXnT4ifUSFFtAzesLWKo9AUQg1lStt8Oz8P\"}\n",
    "signingPublicKeyJWK": "{\"kty\":\"EC\",\"kid\":\"m5xr81U_PIKPLd0uzi6QH_kefkMoiN3fF5OUR34cF5I=\",\"crv\":\"P-384\",\"x\":\"nqrvtZkR8WxH_1URu3rDkd-FtBysDSjfmyUwctEV5xpWH_O71rOCsZDXltB5wBtq\",\"y\":\"0tqE2Atw9OncJsi0gGNw5k2Ay5u3xfCJlVZKPSDjKgqWIOpVYDVFqFN6IFu7qIgX\"}\n"
  },
  "app/production/server/privacy": {
    "principal": "NUgvbUViQy1hdWEwTG1cRGkkMjxVWD4xaXdZT3R5YTF6X0JoMDlMQUpiZ2kuNnVwbDNWXDhQQmFjUlRkLUdwZg=="
  },
  "app/production/vendor/mailgun": {
    "apiKey": "1234567890"
  }
}
```

> To be compliant with our internal secret organization (CSO), we need to rewrite the
> secret paths to append the product key prefix.

Let's create a `BundlePatch` to change all prefixes from `app/{{.Env}}` to `app/{{.Env}}/security/sops-sample/v1.0.0/microservice-1`.

```yaml
apiVersion: harp.elastic.co/v1
kind: BundlePatch
meta:
  name: "secret-relocator"
  description: "Move sops secrets to CSO compliant path"
spec:
  rules:
  - selector:
      matchPath:
        regex: "^app/production/*"
    package:
      path:
        template: |-
            app/production/security/sops-sample/v1.0.0/microservice-1/{{ trimPrefix "app/production/" .Path }}
  - selector:
      matchPath:
        regex: "^app/staging/*"
    package:
      path:
        template: |-
            app/staging/security/sops-sample/v1.0.0/microservice-1/{{ trimPrefix "app/staging/" .Path }}
```

Run the package relocation :

```sh
$ harp from object --on production.yaml \
  | harp bundle patch --spec cso-relocator.yaml \
  | harp bundle dump --data-only \
  | jq
{
  "app/production/security/sops-sample/v1.0.0/microservice-1/database": {
    "password": "c7#kaTKgAqjAuC6KKdAz)_tsa9mBetD83lAk2cSvyH.jB0AeW%/m4g~4Vi]|I7x!",
    "user": "app-xhyXgOEzYaPmFtYS"
  },
  "app/production/security/sops-sample/v1.0.0/microservice-1/server/http/session": {
    "cookieKeyB64": "fHJ1N1Z6YWI5RSIuWm1VNHZCTVB4NGdjSCVmJSFseWNwcjVRT2gwWC95TSlmNVVPfTc5X2tmaHp4TVUxSHR0dQ=="
  },
  "app/production/security/sops-sample/v1.0.0/microservice-1/server/http/token": {
    "signingPrivateKeyJWK": "{\"kty\":\"EC\",\"kid\":\"m5xr81U_PIKPLd0uzi6QH_kefkMoiN3fF5OUR34cF5I=\",\"crv\":\"P-384\",\"x\":\"nqrvtZkR8WxH_1URu3rDkd-FtBysDSjfmyUwctEV5xpWH_O71rOCsZDXltB5wBtq\",\"y\":\"0tqE2Atw9OncJsi0gGNw5k2Ay5u3xfCJlVZKPSDjKgqWIOpVYDVFqFN6IFu7qIgX\",\"d\":\"Ln4mPosjcCXRXq7TcKEVoWS8b2kfpCXnT4ifUSFFtAzesLWKo9AUQg1lStt8Oz8P\"}\n",
    "signingPublicKeyJWK": "{\"kty\":\"EC\",\"kid\":\"m5xr81U_PIKPLd0uzi6QH_kefkMoiN3fF5OUR34cF5I=\",\"crv\":\"P-384\",\"x\":\"nqrvtZkR8WxH_1URu3rDkd-FtBysDSjfmyUwctEV5xpWH_O71rOCsZDXltB5wBtq\",\"y\":\"0tqE2Atw9OncJsi0gGNw5k2Ay5u3xfCJlVZKPSDjKgqWIOpVYDVFqFN6IFu7qIgX\"}\n"
  },
  "app/production/security/sops-sample/v1.0.0/microservice-1/server/privacy": {
    "principal": "NUgvbUViQy1hdWEwTG1cRGkkMjxVWD4xaXdZT3R5YTF6X0JoMDlMQUpiZ2kuNnVwbDNWXDhQQmFjUlRkLUdwZg=="
  },
  "app/production/security/sops-sample/v1.0.0/microservice-1/vendor/mailgun": {
    "apiKey": "1234567890"
  }
}
```

Once everything is compliant with your expectations, publish to Vault.

```sh
$ harp from object --in production.yaml \
  | harp bundle patch --spec cso-relocator.yaml \
  | harp to vault
```

#### Application access profile

Secrets are now published in Vault by the operator. We need to create a secret
consumer role and a bound policy to retrict operations on secrets.

Start by defining the Application role first.

```yaml
apiVersion: harp.elastic.co/terraformer/v1
kind: AppRoleDefinition
meta:
  name: "sample"
  owner: "cloud-security@elastic.co"
  description: "sample service approle & policy"
  issues:
    - <github issue urls attached to the access request>
spec:
  selector:
    platform: "security"
    product: "sample"
    version: "v1.0.0"
    component: "microservice-1"
    environments:
      - production
      - staging

  namespaces:
    # CSO Compliant paths
    application:
      - suffix: "database"
        description: "Database connnection settings"
        capabilities: ["read"]
      - suffix: "server/privacy"
        description: "Privacy anonymizer"
        capabilities: ["read"]
      - suffix: "server/session"
        description: "HTTP Session related secrets"
        capabilities: ["read"]
      - suffix: "server/token"
        description: "JWT Token provider related secrets"
        capabilities: ["read"]
      - suffix: "vendor/mailgun"
        description: "Mailgun vendor"
        capabilities: ["read"]
```

Generate the `service` Vault AppRole and Policy using a terraform script via
`harp-terraformer` plugin.

```sh
$ harp terraformer service --spec approle.yaml
```

It will generate the following HCL script

```ruby
# Generated with Harp Terraformer, Don't modify.
# https://github.com/elastic/harp-plugins/tree/main/cmd/harp-terraformer
# ---
# SpecificationHash: "8nuVOhyalizDvImmlIAoCqH91OBIPUhL4ab24bpUdNE="
# Owner: "cloud-security@elastic.co"
# Date: "2021-08-09T11:14:34Z"
# Description: "sample service approle & policy"
# Issues:
# ---
#
# ------------------------------------------------------------------------------

# Create the policy
data "vault_policy_document" "service-sample-production" {
  # Application secrets
  rule {
    description  = "Database connnection settings"
    path         = "app/data/production/security/sample/v1.0.0/microservice-1/database"
    capabilities = ["read"]
  }

  rule {
    description  = "Privacy anonymizer"
    path         = "app/data/production/security/sample/v1.0.0/microservice-1/server/privacy"
    capabilities = ["read"]
  }

  rule {
    description  = "HTTP Session related secrets"
    path         = "app/data/production/security/sample/v1.0.0/microservice-1/server/session"
    capabilities = ["read"]
  }

  rule {
    description  = "JWT Token provider related secrets"
    path         = "app/data/production/security/sample/v1.0.0/microservice-1/server/token"
    capabilities = ["read"]
  }

  rule {
    description  = "Mailgun vendor"
    path         = "app/data/production/security/sample/v1.0.0/microservice-1/vendor/mailgun"
    capabilities = ["read"]
  }
}

# Register the policy
resource "vault_policy" "service-sample-production" {
  name   = "service-sample-production"
  policy = data.vault_policy_document.service-sample-production.hcl
}

# ------------------------------------------------------------------------------
#
# Register the backend role
resource "vault_approle_auth_backend_role" "sample-production" {
  backend   = "service"
  role_name = "sample-production"

  token_policies = [
    "service-default",
    "service-sample-production",
  ]
}
```

### Mozilla SOPS

Mozilla SOPS is a file-encryption based solution which allows you to handle
multi-recipient encryption where keys can be handled by Cloud KMS, PGP, Age, etc.

As a sample, I decided to use `age` encryption as an alternative to PGP.

#### Create the identity

We need to generate a secret key pair. Using `age-keygen`.
```sh
$ age-keygen -o sample-production.age
Public key: age18nnjwxyu4v45ma3dp40n42qfzhlxz5fuwuxlmkcf8x625wusze0s0qy4ug
```

This file is sensitive and must be kept secret.

```ruby
# created: 2021-08-10T11:46:29+02:00
# public key: age18nnjwxyu4v45ma3dp40n42qfzhlxz5fuwuxlmkcf8x625wusze0s0qy4ug
AGE-SECRET-KEY-13HA98GJ0NGZY7KCHQVCM5SYZG9EJ8VR0UJW4FW0NWSXY49T280TSWUM7MG
```

#### Seal using sops

Encrypt and replace file content using the `age` public key. So that only the
owner of the private key could decrypt the file.

```sh
$ sops -e -i \
    --age age18nnjwxyu4v45ma3dp40n42qfzhlxz5fuwuxlmkcf8x625wusze0s0qy4ug \
    production.yaml
```

Now you can see that the `production.yaml` file is encrypted, but it keeps its
structure visible.

```yaml
app:
    production:
        database:
            user: ENC[AES256_GCM,data:oQRwVndIF6nEdok1NwjaQLNjF64=,iv:+20uH+6df96a3a9cqIH8WCKjPyQzKkr7TmnCVAey8A0=,tag:0zjvnOldE/wP8SGBOXoDaw==,type:str]
            password: ENC[AES256_GCM,data:QfNWOZajBhqBlCYTf/7+0xRMsA6b8kdF8Z5B6wBDHMI0qOBf6OvAgMlwa2In1jTOs/LZtQA/USVf4VijyyVeug==,iv:Y0UKXVwPglJBrfN9jD0EHHI7VUl4lJwu4kd7DbzsLX4=,tag:MB/tZFpt8iARmDC8SUfl0g==,type:str]
        server:
            privacy:
                principal: ENC[AES256_GCM,data:sU//BpXSuCslRI0QEWn7OD40kvvbuer/E44ahKL1uq833g3rpAlH56VMnX8JIWe2HEjFPbcW73Ln2NznhWInIMeQw2Z8LbchVxLZKw8dvaJfuP2hNN9x4A==,iv:32mMVST6ssYrFw357+vJt86alY8D6u6+EQOP7i0lnp0=,tag:6qdFYygrMB0ZY3VIO6i3Mw==,type:str]
            http:
                session:
                    cookieKeyB64: ENC[AES256_GCM,data:hNQ6CEpwIPH+/7S1QcmVRn9ID+Bl49BzA1kKT7EjhuwD4coWzH0i1gGCE1GlGjZBqY2g0Kr6zuD/n38x5xsAX/CWa8JO2Nakw5lqPK0R00bZtnRpi2m1Gw==,iv:mUqwnfQ1f4EfoC0zLWNZ1rkfgLykWsUfDbdC+IHijls=,tag:GxSl7iDPOGDuRSfG9Hy7OQ==,type:str]
                token:
                    signingPrivateKeyJWK: ENC[AES256_GCM,data:vYw8oXa1PiFyVpQi63VdrcxZLY9fHWp5aImi2MF61j5zwNPzTrjEWfyFTElEieppb6BLiFRhfa2xh8jq1FIxzzhnkcQkjTGfOPRgON6si2Oxy94m8iPWiZFFQticPShSCTyOjion9KAhxCjwSHJ0YLc5DJ7Xab8/LLNhz+UvT3O8fnw0lSk2e+qcn7TcN6a9EM73LTFYnneOjRqKC3p8zMKg7wmlHiPKTabd0Vr1jYWRURetdmb5km2ee8YXJ7qziyPyJGn13QHfUcOoCgrgnmRVuWEgeB/7g17Dg7lLwyQPW2BnUTvBiaowjUSkUGs5NhybLx7wlHBX6kcKWm7aOyH+3Ma80Gyi4jn/Lwap1UjhBMEZkMRV3zSrPt6Kqc956k2vNA==,iv:FMitt6BsTuyQLhOtrvHl51paNvwgFRromyK6HMnasAU=,tag:gFswivMyMMZdB3lx6PEJzw==,type:str]
                    signingPublicKeyJWK: ENC[AES256_GCM,data:G4vXSZ+oEdp3T3y3x5eRfhg4bwqpwaAeZlrB56nAer2it3ptkhwaTONp2SndmZDB5vMzNo8ARMTYhpVfdJfvSCxf+6ovJ2PCW0pcYhVS4ynEUgkZWFDnlqFDv0oQd3t6CbIN4ZeO/XX5k8w0TK+HCV54m4nQaywLHnm1b+qtH1sssO5OXRKeSRqC3VZe7UaiTFLorkkzdGBu/JKTi5iPTAxrNWRJ5le0+FuENNUohMU13Sih2D8SGzmthSLFvr6ntbYgwpbyj4IiObS/35TGcDf1cMNzztvZ6JbDPDM=,iv:13Tn46zhwt4sMt8QpvDiZzCMgJFBgSzB+gIbPZcgZs0=,tag:API7I+Mg6WEbpJIzhaZ/Wg==,type:str]
        vendor:
            mailgun:
                apiKey: ENC[AES256_GCM,data:7RkPURpjiMu8tw==,iv:MOYpKEGiazFrTx7w1tbjUcVrRas4ckIKB6fhoLB4TEQ=,tag:XTZgMu9hriG9mavmo7gdZA==,type:str]
sops:
    kms: []
    gcp_kms: []
    azure_kv: []
    hc_vault: []
    age:
        - recipient: age18nnjwxyu4v45ma3dp40n42qfzhlxz5fuwuxlmkcf8x625wusze0s0qy4ug
          enc: |
            -----BEGIN AGE ENCRYPTED FILE-----
            YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IFgyNTUxOSBnQlFnQ1lDdGZQVzJGTVBW
            TzJycE9zZk53MGU3Uk9sY3gzR0lLZnlidkdBCkRPZjNOcThQT2VVckJLZ3BCay9V
            cVVxSndwUXRYdjU2cFJ0YWdWT2llY28KLS0tIFdZbE1FOW13SHd1UWpVMXpSR2Fj
            eUNzRXNnbTc4dzNXVFhGeGpXcU9KaEEKes529R7EtlY6IMHVpCjeL0Wk9Aj2r0ch
            Wi3NVJSKrsHKsIPdwCMkvFXP9dkWRfvowARZqCTQy18aITUhfwdN6g==
            -----END AGE ENCRYPTED FILE-----
    lastmodified: "2021-08-10T11:43:30Z"
    mac: ENC[AES256_GCM,data:GLVSyoSeWm2obzgl/8HXNx9nKsw7FiIyp6ypQpEXZsmD/quysoMphviv6AoLfwu8u2CUUyi8Eycd9bCSP6bvbkihVT1rsH42wvNatPjm/YwdK/CYF1mrveLskwlWAYEqvFjd4ls+IZh63N7fP5q5waUaDdPY2H21vKcpULujdZQ=,iv:YO7qZKNqR1M8K+vV4DqBKXA69MPj8xhlgr+pQCr12Qk=,tag:drXuAtMSyYm3nEUylO8CVA==,type:str]
    pgp: []
    unencrypted_suffix: _unencrypted
    version: 3.7.1
```

You can now safely publish `production.yaml` file to Git

```sh
$ git add production.yaml
$ git commit -a -m "chore(secrets): update production secrets."
$ git push
```

#### Retrieve a bundle from a sops encrypted file

You can decrypt the `production.yaml` file if you have access to the private
key used to encrypt.

```sh
$ export SOPS_AGE_KEY_FILE=sample-production.age
$ sops -d production.yaml
app:
    production:
        database:
            user: app-tNtKnOgYXuuJyArI
            password: (Pzd~O9JQ"Epm)QN-eWZAscJHxjyfQlXXq_t=HwEc1it3Bu9h+A7um777q"t4n*8
        server:
            privacy:
                principal: R0FLZVROZDZlMGRUPFYsQXJRWFluVGlJYUFCTDBIT3p5a145TTRAZjNsc0B6XnhmKGo0djU4Y2AzY01vaDolcA==
            http:
                session:
                    cookieKeyB64: aV5kU292VTYvTlhJM1BrdDgzSEJUcXgrOSVIfkU2Tlp2azN4YXBhWlEpLmJjZiNkZz96VyVvQ0piN3YwbHRpNQ==
                token:
                    signingPrivateKeyJWK: '{"kty":"EC","kid":"sAt9gTz5oUmTefVK0Ib1OfRv7uKVC1at9bB1VjR52QE=","crv":"P-384","x":"XeCZRaJDh4i1ywansUMAh2kN6WbEqWNnQc0diC0SkVxmCAcxA69PQKbyYJy49z9Y","y":"FuDWJpFuuYX0JDRSJTQ5uexLCU3-G4tGEGAoRtWQLRCZIpz4tfcd87bFiIhAb4MT","d":"gHmrbNC5X8rp4ORxwQkdeqMctMjP_jLO3ngd_oxp-mclK3O7TIpPM0dm7TEx0AjR"}'
                    signingPublicKeyJWK: '{"kty":"EC","kid":"sAt9gTz5oUmTefVK0Ib1OfRv7uKVC1at9bB1VjR52QE=","crv":"P-384","x":"XeCZRaJDh4i1ywansUMAh2kN6WbEqWNnQc0diC0SkVxmCAcxA69PQKbyYJy49z9Y","y":"FuDWJpFuuYX0JDRSJTQ5uexLCU3-G4tGEGAoRtWQLRCZIpz4tfcd87bFiIhAb4MT"}'
        vendor:
            mailgun:
                apiKey: "1234567890"
```

In order to convert a YAML object to an Harp Secret Bundle, you have to use
`harp from object`.

```sh
$ sops -d production.yaml \
  | harp from object \
  | harp bundle dump --data-only \
  | jq
{
  "app/production/database": {
    "password": "(Pzd~O9JQ\"Epm)QN-eWZAscJHxjyfQlXXq_t=HwEc1it3Bu9h+A7um777q\"t4n*8",
    "user": "app-tNtKnOgYXuuJyArI"
  },
  "app/production/server/http/session": {
    "cookieKeyB64": "aV5kU292VTYvTlhJM1BrdDgzSEJUcXgrOSVIfkU2Tlp2azN4YXBhWlEpLmJjZiNkZz96VyVvQ0piN3YwbHRpNQ=="
  },
  "app/production/server/http/token": {
    "signingPrivateKeyJWK": "{\"kty\":\"EC\",\"kid\":\"sAt9gTz5oUmTefVK0Ib1OfRv7uKVC1at9bB1VjR52QE=\",\"crv\":\"P-384\",\"x\":\"XeCZRaJDh4i1ywansUMAh2kN6WbEqWNnQc0diC0SkVxmCAcxA69PQKbyYJy49z9Y\",\"y\":\"FuDWJpFuuYX0JDRSJTQ5uexLCU3-G4tGEGAoRtWQLRCZIpz4tfcd87bFiIhAb4MT\",\"d\":\"gHmrbNC5X8rp4ORxwQkdeqMctMjP_jLO3ngd_oxp-mclK3O7TIpPM0dm7TEx0AjR\"}",
    "signingPublicKeyJWK": "{\"kty\":\"EC\",\"kid\":\"sAt9gTz5oUmTefVK0Ib1OfRv7uKVC1at9bB1VjR52QE=\",\"crv\":\"P-384\",\"x\":\"XeCZRaJDh4i1ywansUMAh2kN6WbEqWNnQc0diC0SkVxmCAcxA69PQKbyYJy49z9Y\",\"y\":\"FuDWJpFuuYX0JDRSJTQ5uexLCU3-G4tGEGAoRtWQLRCZIpz4tfcd87bFiIhAb4MT\"}"
  },
  "app/production/server/privacy": {
    "principal": "R0FLZVROZDZlMGRUPFYsQXJRWFluVGlJYUFCTDBIT3p5a145TTRAZjNsc0B6XnhmKGo0djU4Y2AzY01vaDolcA=="
  },
  "app/production/vendor/mailgun": {
    "apiKey": "1234567890"
  }
}
```

From this point, you can apply all harp available bundle operations.

#### Generate a YAML output from a bundle

For example to have applied a `BundlePatch` to rotate a secret value.

```yaml
apiVersion: harp.elastic.co/v1
kind: BundlePatch
meta:
  name: "token-jwk-rotator"
  description: "Create a new JWK key for JWT signing."
spec:
  rules:
  - selector:
      matchPath:
        strict: "app/production/server/http/token"
    package:
        data:
            template: |-
                {
                    "signingPrivateKeyJWK": {{ $key := cryptoPair "ec:p384" }}{{ $key.Private | toJwk | toJson }},
                    "signingPublicKeyJWK": {{ $key.Public | toJwk | toJson }}
                }
```

> This patch will look for `app/production/server/http/token` package in the
> given bundle and generate 2 secrets `signingPrivateKeyJWK` and `signingPublicKeyJWK`
> which will replace the old values.

```sh
$ sops -d production.yaml \
  | harp from object \
  | harp bundle patch --spec token-jwk-rotator.yaml --out rotated.bundle
$ harp bundle dump --in rotated.bundle --data-only \
  | jq
{
  ...
  "app/production/server/http/token": {
    "signingPrivateKeyJWK": "{\"kty\":\"EC\",\"kid\":\"IM_00Nm7zruMnBTxumkzT-LNiOnfGWStjznFm5diRVc=\",\"crv\":\"P-384\",\"x\":\"sH6CbEV9HqgSAxCdjjBOPgdw1xvAcMrNaOl2Vkrq9x6LfJKxEA1qioTa_1zAYbJ7\",\"y\":\"5iOnCQRyXHUPzg8RL2PV_-jdKTUpzH4v1QPBAltEdyJ_4rwJwT5oFdNm_Uqqpy2d\",\"d\":\"TE0i2Ry0Dailcs2FczPsOapePqh85uSV538M_iUNJlpkorguLpauMnxY4zSBzohd\"}",
    "signingPublicKeyJWK": "{\"kty\":\"EC\",\"kid\":\"IM_00Nm7zruMnBTxumkzT-LNiOnfGWStjznFm5diRVc=\",\"crv\":\"P-384\",\"x\":\"sH6CbEV9HqgSAxCdjjBOPgdw1xvAcMrNaOl2Vkrq9x6LfJKxEA1qioTa_1zAYbJ7\",\"y\":\"5iOnCQRyXHUPzg8RL2PV_-jdKTUpzH4v1QPBAltEdyJ_4rwJwT5oFdNm_Uqqpy2d\"}"
  },
  ...
}
```

Check the updated secrets

```sh
$ sops -d production.yaml \
  | harp from object \
  | harp bundle diff --old - --new rotated.bundle \
  | jq
[
  {
    "op": "replace",
    "type": "secret",
    "path": "app/production/server/http/token#signingPrivateKeyJWK",
    "value": "{\"kty\":\"EC\",\"kid\":\"IM_00Nm7zruMnBTxumkzT-LNiOnfGWStjznFm5diRVc=\",\"crv\":\"P-384\",\"x\":\"sH6CbEV9HqgSAxCdjjBOPgdw1xvAcMrNaOl2Vkrq9x6LfJKxEA1qioTa_1zAYbJ7\",\"y\":\"5iOnCQRyXHUPzg8RL2PV_-jdKTUpzH4v1QPBAltEdyJ_4rwJwT5oFdNm_Uqqpy2d\",\"d\":\"TE0i2Ry0Dailcs2FczPsOapePqh85uSV538M_iUNJlpkorguLpauMnxY4zSBzohd\"}"
  },
  {
    "op": "replace",
    "type": "secret",
    "path": "app/production/server/http/token#signingPublicKeyJWK",
    "value": "{\"kty\":\"EC\",\"kid\":\"IM_00Nm7zruMnBTxumkzT-LNiOnfGWStjznFm5diRVc=\",\"crv\":\"P-384\",\"x\":\"sH6CbEV9HqgSAxCdjjBOPgdw1xvAcMrNaOl2Vkrq9x6LfJKxEA1qioTa_1zAYbJ7\",\"y\":\"5iOnCQRyXHUPzg8RL2PV_-jdKTUpzH4v1QPBAltEdyJ_4rwJwT5oFdNm_Uqqpy2d\"}"
  }
]
```

Regenerate the YAML output for SOPS :

```sh
$ harp to object --in rotated.bundle \
    --format yaml
```

Generate the YAML flattened secret tree :

```yaml
app/production/database:
    password: (Pzd~O9JQ"Epm)QN-eWZAscJHxjyfQlXXq_t=HwEc1it3Bu9h+A7um777q"t4n*8
    user: app-tNtKnOgYXuuJyArI
app/production/server/http/session:
    cookieKeyB64: aV5kU292VTYvTlhJM1BrdDgzSEJUcXgrOSVIfkU2Tlp2azN4YXBhWlEpLmJjZiNkZz96VyVvQ0piN3YwbHRpNQ==
app/production/server/http/token:
    signingPrivateKeyJWK: '{"kty":"EC","kid":"IM_00Nm7zruMnBTxumkzT-LNiOnfGWStjznFm5diRVc=","crv":"P-384","x":"sH6CbEV9HqgSAxCdjjBOPgdw1xvAcMrNaOl2Vkrq9x6LfJKxEA1qioTa_1zAYbJ7","y":"5iOnCQRyXHUPzg8RL2PV_-jdKTUpzH4v1QPBAltEdyJ_4rwJwT5oFdNm_Uqqpy2d","d":"TE0i2Ry0Dailcs2FczPsOapePqh85uSV538M_iUNJlpkorguLpauMnxY4zSBzohd"}'
    signingPublicKeyJWK: '{"kty":"EC","kid":"IM_00Nm7zruMnBTxumkzT-LNiOnfGWStjznFm5diRVc=","crv":"P-384","x":"sH6CbEV9HqgSAxCdjjBOPgdw1xvAcMrNaOl2Vkrq9x6LfJKxEA1qioTa_1zAYbJ7","y":"5iOnCQRyXHUPzg8RL2PV_-jdKTUpzH4v1QPBAltEdyJ_4rwJwT5oFdNm_Uqqpy2d"}'
app/production/server/privacy:
    principal: R0FLZVROZDZlMGRUPFYsQXJRWFluVGlJYUFCTDBIT3p5a145TTRAZjNsc0B6XnhmKGo0djU4Y2AzY01vaDolcA==
app/production/vendor/mailgun:
    apiKey: "1234567890"
```

Use `--expand` to regenerate the expanded secret tree :

```sh
$ harp to object --in rotated.bundle \
    --format yaml \
    --expand
app:
    production:
        database:
            password: (Pzd~O9JQ"Epm)QN-eWZAscJHxjyfQlXXq_t=HwEc1it3Bu9h+A7um777q"t4n*8
            user: app-tNtKnOgYXuuJyArI
        server:
            http:
                session:
                    cookieKeyB64: aV5kU292VTYvTlhJM1BrdDgzSEJUcXgrOSVIfkU2Tlp2azN4YXBhWlEpLmJjZiNkZz96VyVvQ0piN3YwbHRpNQ==
                token:
                    signingPrivateKeyJWK: '{"kty":"EC","kid":"IM_00Nm7zruMnBTxumkzT-LNiOnfGWStjznFm5diRVc=","crv":"P-384","x":"sH6CbEV9HqgSAxCdjjBOPgdw1xvAcMrNaOl2Vkrq9x6LfJKxEA1qioTa_1zAYbJ7","y":"5iOnCQRyXHUPzg8RL2PV_-jdKTUpzH4v1QPBAltEdyJ_4rwJwT5oFdNm_Uqqpy2d","d":"TE0i2Ry0Dailcs2FczPsOapePqh85uSV538M_iUNJlpkorguLpauMnxY4zSBzohd"}'
                    signingPublicKeyJWK: '{"kty":"EC","kid":"IM_00Nm7zruMnBTxumkzT-LNiOnfGWStjznFm5diRVc=","crv":"P-384","x":"sH6CbEV9HqgSAxCdjjBOPgdw1xvAcMrNaOl2Vkrq9x6LfJKxEA1qioTa_1zAYbJ7","y":"5iOnCQRyXHUPzg8RL2PV_-jdKTUpzH4v1QPBAltEdyJ_4rwJwT5oFdNm_Uqqpy2d"}'
            privacy:
                principal: R0FLZVROZDZlMGRUPFYsQXJRWFluVGlJYUFCTDBIT3p5a145TTRAZjNsc0B6XnhmKGo0djU4Y2AzY01vaDolcA==
        vendor:
            mailgun:
                apiKey: "1234567890"
```

Save the output to a file, apply sops encryption and git commit/push :

```sh
$ harp to object --in rotated.bundle \
    --format yaml \
    --expand \
    --out production.yaml
$ sops -e -i \
    --age age18nnjwxyu4v45ma3dp40n42qfzhlxz5fuwuxlmkcf8x625wusze0s0qy4ug \
    production.yaml
$ git add production.yaml
$ git commit -a -m "chore(secrets): update production secrets."
$ git push
```

### Complete scenario

```sh
# Retrieve most recent secret cold state
$ git pull
# Decrypt and retrieve actual secret state
$ sops -d production.yaml \
  | harp from object \
  | harp bundle patch --spec token-jwk-rotator.yaml --out rotated.bundle
# Publish changes to Vault (can be done via GitOps after push)
$ harp bundle patch --spec cso-relocator.yaml --in rotated.bundle \
  | harp to vault
# Publish changes to SOPS
$ harp to object --in rotated.bundle \
    --format yaml \
    --expand \
    --out production.yaml
# Encrypt the production secrets
$ sops -e -i \
    --age $CI_PUBLIC_KEY \
    production.yaml
# Cleanup
$ rm -f rotated.bundle
# Push them to git
$ git add production.yaml
$ git commit -a -m "chore(ci): secrets/sample - auto-rotate JWT Token signing key."
$ git push
```
