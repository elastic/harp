# Act as Vault server with namespace emulation

Expose multiple harp crates as separated Vault namespaces.

## Scenario

### Prepare bundles

> Use the bundle template from `customer-bundle` sample.

#### Staging bundle

Generate bundle using specification and set quality as `staging`.

```sh
$ harp from template \
    --in spec.yaml \
    --values values.yaml \
    --set quality=staging
```

#### Production bundle

Generate bundle using specification and set quality as `production`.

```sh
$ harp from template \
    --in spec.yaml \
    --values values.yaml \
    --set quality=production
```

### List available secret path

> Some secret paths are quality related

```go
app/{{.quality}}/customer1/ece/v1.0.0/adminconsole/authentication/otp/okta_api_key
app/{{.quality}}/customer1/ece/v1.0.0/adminconsole/database/usage_credentials
app/{{.quality}}/customer1/ece/v1.0.0/adminconsole/http/session
app/{{.quality}}/customer1/ece/v1.0.0/adminconsole/mailing/sender/mailgun_api_key
app/{{.quality}}/customer1/ece/v1.0.0/adminconsole/privacy/anonymizer
app/{{.quality}}/customer1/ece/v1.0.0/userconsole/database/usage_credentials
app/{{.quality}}/customer1/ece/v1.0.0/userconsole/http/certificate
app/{{.quality}}/customer1/ece/v1.0.0/userconsole/http/session
infra/aws/essp-customer1/us-east-1/rds/adminconsole/accounts/root_credentials
platform/{{.quality}}/customer1/us-east-1/billing/recurly/vendor_api_key
platform/{{.quality}}/customer1/us-east-1/postgresql/admiconsole/admin_credentials
platform/{{.quality}}/customer1/us-east-1/zookeeper/accounts/admin_credentials
product/ece/v1.0.0/artifact/signature/key
```

### Expose as Vault namespaces

```sh
harp server vault \
    --namespace staging:bundle://$(pwd)/staging.bundle \
    --namespace production:bundle://$(pwd)/production.bundle
```

> You can load bundle remotely from S3 also.

### Query with Vault CLI

```sh
export VAULT_ADDR=http://127.0.0.1:8200
```

```sh
$ vault kv get -namespace=staging infra/aws/essp-customer1/us-east-1/rds/adminconsole/accounts/root_credentials
====== Data ======
Key         Value
---         -----
password    eWVVN0JCXEtVUm5qInV6NXYjRUtvbUF2RjguPlxlZ355SWZLWFdmeWU5SG1sN2pyZXMyQzJUKjB4VWNgM0ckMg==
user        dbroot-qNsb7eHc
```

```sh
$ vault kv get -namespace=production infra/aws/essp-customer1/us-east-1/rds/adminconsole/accounts/root_credentials
====== Data ======
Key         Value
---         -----
password    dkhYNjg3dVBhaEp9a2tKOTxjMm5mMlxyfUUvQkc8WmxwQ1dCaSFtNW5HbldpbmJFOTJkQn03Ri1lYT9FdkNjZQ==
user        dbroot-rwMvn1hW
```
