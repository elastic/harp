# Mozilla SOPS integration

In this sample, you will use `sops` encrypted bundle to convert it to a CSO compliant
Vault secret tree.

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
                apiKey: {{ .Values.mailgun.apikey }}
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

## Generate environments

### Production

Generate `production` identity for secret decryption :

```sh
$ age-keygen -o app-production.age
Public key: age12r9nsmgxf7329jlwrv548pldhghx0jk7raxtgnz8x6j5jkp5ygzqzsmqkt
```

> This key is handled only by the subject who is able to read the production secrets.

Prepare `production` secrets :

```sh
harp template --in template.yaml --set env=production --set mailgun.apikey=0123456789 --out production.yaml
```

### Staging

Generate `staging` identity for secret decryption :

```sh
$ age-keygen -o app-staging.age
Public key: age136gsn220gwz8qjthl0ypyuxy06d7lkhmjjgd86g7r7uap6f3a48s02cfex
```

> This key is handled only by the subject who is able to read the staging secrets.

Prepare `staging` secrets :

```sh
harp template --in template.yaml --set env=staging --set mailgun.apikey=9876543210 --out staging.yaml
```

## Seal secrets with sops

Seal `production` secrets with production identity

```sh
sops --age age12r9nsmgxf7329jlwrv548pldhghx0jk7raxtgnz8x6j5jkp5ygzqzsmqkt -i -e production.yaml
```

Seal `staging` secrets with staging identity

```sh
sops --age age136gsn220gwz8qjthl0ypyuxy06d7lkhmjjgd86g7r7uap6f3a48s02cfex -i -e staging.yaml
```

## Deploy secrets to Git as a cold storage

It will store the encrypted yaml file as a Git repository. It can be useful to
get a monolithing state of active secrets. But it has a huge drawback when you
want to audit / authorize access from a secret consumer point of view.

```sh
$ git add production.yaml
$ git commit production.yaml -m "chore(ci): update production secrets"
$ git push
```

## Resource provisioning

> Can be done via a git commit hook for GitOps oriented secret management.

Use the generated yaml files to provision your corresponding resources (accounts / databases / etc.)

## Deploy consumer secrets to Vault

> Can be done via a git commit hook for GitOps oriented secret management.

Check the sops bundle conversion.

```sh
$ export SOPS_AGE_KEY_FILE=app-production.age
$ sops -d production.yaml | harp from yaml | harp bundle dump --data-only | jq
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

We need to rewrite the secret paths to append the product key prefix to be compliant with CSO Secret naming convention.
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
$ sops -d production.yaml | harp from yaml | harp bundle patch --spec cso-relocator.yaml | harp bundle dump --data-only | jq
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
$ sops -d production.yaml | harp from yaml | harp bundle patch --spec cso-relocator.yaml | harp to vault
```

### Application access profile

Start by defining the Application role first.

```yaml
apiVersion: harp.elastic.co/terraformer/v1
kind: AppRoleDefinition
meta:
  name: "app"
  owner: "cloud-security@elastic.co"
  description: "app service approle & policy"
spec:
  selector:
    platform: "security"
    product: "sops-sample"
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
# Description: "app service approle & policy"
# Issues:
# ---
#
# ------------------------------------------------------------------------------

# Create the policy
data "vault_policy_document" "service-app-production" {
  # Application secrets
  rule {
    description  = "Database connnection settings"
    path         = "app/data/production/security/sops-sample/v1.0.0/microservice-1/database"
    capabilities = ["read"]
  }

  rule {
    description  = "Privacy anonymizer"
    path         = "app/data/production/security/sops-sample/v1.0.0/microservice-1/server/privacy"
    capabilities = ["read"]
  }

  rule {
    description  = "HTTP Session related secrets"
    path         = "app/data/production/security/sops-sample/v1.0.0/microservice-1/server/session"
    capabilities = ["read"]
  }

  rule {
    description  = "JWT Token provider related secrets"
    path         = "app/data/production/security/sops-sample/v1.0.0/microservice-1/server/token"
    capabilities = ["read"]
  }

  rule {
    description  = "Mailgun vendor"
    path         = "app/data/production/security/sops-sample/v1.0.0/microservice-1/vendor/mailgun"
    capabilities = ["read"]
  }
}

# Register the policy
resource "vault_policy" "service-app-production" {
  name   = "service-app-production"
  policy = data.vault_policy_document.service-app-production.hcl
}

# ------------------------------------------------------------------------------
#
# Register the backend role
resource "vault_approle_auth_backend_role" "app-production" {
  backend   = "service"
  role_name = "app-production"

  token_policies = [
    "service-default",
    "service-app-production",
  ]
}
```
