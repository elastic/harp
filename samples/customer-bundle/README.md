# Customer BundleTemplate

This defines a set of secret value that can be parametrized at generation time.
It will generate an atomic and complete bundle for a given and defined
compilation environment.

## Scenario

### Generate the bundle

Given this value file

```yaml
infra:
    aws:
        account: essp-customer1
        region: us-east-1

platform:
    essp:
        name: customer1

product:
    name: ece
    version: v1.0.0

vendor:
  recurly:
    api_key: "recurly-foo-api-123456789"
  okta:
    api_key: "okta-foo-api-123456789"
  mailgun:
    api_key: "mg-admin-9875s-sa"
```

In order to generate the bundle :

```sh
harp from template --in spec.yaml --values values.yaml --out customer1.bundle
```

### Publish the bundle

#### List secret path

```sh
$ harp bundle dump --in customer1.bundle --content-only | jq -r "keys"
[
  "app/production/customer1/ece/v1.0.0/adminconsole/authentication/otp/okta_api_key",
  "app/production/customer1/ece/v1.0.0/adminconsole/database/usage_credentials",
  "app/production/customer1/ece/v1.0.0/adminconsole/http/session",
  "app/production/customer1/ece/v1.0.0/adminconsole/mailing/sender/mailgun_api_key",
  "app/production/customer1/ece/v1.0.0/adminconsole/privacy/anonymizer",
  "app/production/customer1/ece/v1.0.0/userconsole/database/usage_credentials",
  "app/production/customer1/ece/v1.0.0/userconsole/http/certificate",
  "app/production/customer1/ece/v1.0.0/userconsole/http/session",
  "infra/aws/essp-customer1/us-east-1/rds/adminconsole/accounts/root_credentials",
  "platform/production/customer1/us-east-1/billing/recurly/vendor_api_key",
  "platform/production/customer1/us-east-1/postgresql/admiconsole/admin_credentials",
  "platform/production/customer1/us-east-1/zookeeper/accounts/admin_credentials",
  "product/ece/v1.0.0/artifact/signature/key"
]
```

or use flag `--path-only`

```sh
$ harp bundle dump --in customer1.bundle --path-only
app/production/customer1/ece/v1.0.0/adminconsole/authentication/otp/okta_api_key
app/production/customer1/ece/v1.0.0/adminconsole/database/usage_credentials
app/production/customer1/ece/v1.0.0/adminconsole/http/session
app/production/customer1/ece/v1.0.0/adminconsole/mailing/sender/mailgun_api_key
app/production/customer1/ece/v1.0.0/adminconsole/privacy/anonymizer
app/production/customer1/ece/v1.0.0/userconsole/database/usage_credentials
app/production/customer1/ece/v1.0.0/userconsole/http/certificate
app/production/customer1/ece/v1.0.0/userconsole/http/session
infra/aws/essp-customer1/us-east-1/rds/adminconsole/accounts/root_credentials
platform/production/customer1/us-east-1/billing/recurly/vendor_api_key
platform/production/customer1/us-east-1/postgresql/admiconsole/admin_credentials
platform/production/customer1/us-east-1/zookeeper/accounts/admin_credentials
product/ece/v1.0.0/artifact/signature/key
```

> This output is `xargs`-able.

#### Validate generated secret values

```sh
harp bundle dump --in customer1.bundle --content-only
```

#### Export to Vault

```sh
VAULT_ADDR=https://.......
VAULT_TOKEN=$(vault login -method=oidc -token-only)
```

You can batch import in Vault the complete bundle

```sh
harp to vault --in customer1.bundle
```

Or unitary patch secret values

```sh
harp bundle read \
    --in ec2_ssh.bundle \
    --path app/production/customer1/ece/v1.0.0/userconsole/database/usage_credentials \
    | vault kv put app/production/customer1/ece/v1.0.0/userconsole/database/usage_credentials -
```

#### Expose bundle via harp-server

You are in the situation where Vault is not available yet in your newly fresh
environment, for this situation you have to use the `harp-server` plugin.

##### Seal the bundle

In order to export the container outside your trusted zone, the container needs to
be sealed.

First prepare a recovery identity :

```sh
# Generate a recovery identity passphrase
$ harp passphrase > pass.txt
# Generate the recovery identity
$ harp container identity --passphrase $(cat pass.txt) --description "Recovery" --out recovery.json
$ cat recovery.json
```

The `recovery.json` is the identity used to seal the container and allowed to
unseal it if you forget the container key.

```json
{
  "@apiVersion": "harp.elastic.co/v1",
  "@kind": "ContainerIdentity",
  "@timestamp": "2020-10-26T17:32:38.352639Z",
  "@description": "Recovery",
  "public": "nkxAdLoiK3y_03QyrudlqKa4pV_HtBvMCZKLlhXiEDc",
  "private": {
    "encoding": "jwe",
    "content": "eyJhbGciOiJQQkVTMi1IUzUxMitBMjU2S1ciLCJjdHkiOiJqd2sranNvbiIsImVuYyI6IkEyNTZHQ00iLCJwMmMiOjUwMDAwMSwicDJzIjoiWTFBNWEwUTBlblZMU0dKTmJUTnBlUSJ9.rmIZhqU4TDXWlv2WwMHUsexrVDssciiPm4IlagB_Mamdj4eYbckKzg.7JfMfByOMj8-VY2P.I6K_Eml3Xk8K7pBTCSSlX9YqM9ZuzGAnsjYy2VWfUwtce4H1UbZ7fBjH5FiCH80HgCGf7gf5eI7BeZMZ9mBkjOZOErzAp660a4UNgfeYD2ivxAEFGjLpl74brI1fNgBymuxUyScCl12sMRwnjxdqXLN-CuKqfhAmckyKLxMan34.edLt6V03_ChtO7nlXSZPMg"
  }
}
```

Seal the secret container using recovery identity :

```sh
$ harp container seal \
    --identity $(cat recovery.json | jq -r '.public') \
    --in unsealed.bundle --out sealed.bundle
Container key : 4Vyy8xH_zNpSJbFyheC8dPgjmxs54YvLa6QZpDKE1y0
```

Now your `Bundle` is secured as a sealed secret container.

##### Expose the sealed container

With harp, you have multiple strategies to retrieve a bundle (local, s3, etc.)

Expose as Vault server :

```sh
harp server vault \
    --namespace customer1:bundle://$(pwd)/customer1-sealed.bundle\?cid=4Vyy8xH_zNpSJbFyheC8dPgjmxs54YvLa6QZpDKE1y0
```

Expose an HTTP Server :

```sh
harp server http \
    --namespace customer1:bundle://$(pwd)/customer1-sealed.bundle\?cid=4Vyy8xH_zNpSJbFyheC8dPgjmxs54YvLa6QZpDKE1y0
```

Expose a gRPC Server :

```sh
harp server grpc \
    --namespace customer1:bundle://$(pwd)/customer1-sealed.bundle\?cid=4Vyy8xH_zNpSJbFyheC8dPgjmxs54YvLa6QZpDKE1y0
```
