# Act as a readonly Vault server with mTLS

Use harp as a Hashicorp Vault API compatible service to use Vault environment
tools (consul-template, etc.).

> Use and prepare vault compatible tools during bootstrap phase when vault is
> not ready yet, or not accessible.

## Scenario

### Prepare your bundle

> Use the bundle template from `customer-bundle` sample.

```sh
harp from template \
    --in spec.yaml \               # Input specification
    --values values.yaml \         # Static values (vendored secrets, etc.)
    --set quality=production \     # Inline value override
    --out production.bundle        # Output bundle
```

### List available secret path

```sh
$ harp bundle dump --in customer1-production.bundle --content-only | jq -r "keys"
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

### Expose the bundle as read-only Vault server

Map a bundle to `root` namespace as default.

```sh
harp-server vault \
  --namespace root:bundle://$(pwd)/customer1-production.bundle
```

This will expose a Vault HTTP API based on secret created in the bundle.

### Read using Vault CLI

```sh
export VAULT_ADDR=http://127.0.0.1:8200
```

> All secret path will be interpreted as a KVv2 backend as required by the CSO.
> These paths must be symmetric with secrets deployed in Vault.

```sh
$ vault kv get app/production/customer1/ece/v1.0.0/userconsole/database/usage_credentials
====== Data ======
Key         Value
---         -----
dbname      userconsole
host        sample-instance.abc2defghije.us-west-2.rds.amazonaws.com
options     sslmode=require&application_name=userconsole
password    KzRRPEtLNHp0d1BEW1k2cEUqSD5MUFFwKzl1Q045MVNtc1hqUz9NdntwaGsyUDRHRlM4SWtUJktFdzNYWSh2Zw==
port        5432
username    dbuser-userconsole-YBP3lOMy
```

### Protect fake vault server with mutual TLS

#### Generate CA

CA settings `ca-config.json`.

```sh
cat << EOF > ca-config.json
{
  "signing": {
    "default": {
      "expiry": "8760h"
    },
    "profiles": {
      "server": {
        "usages": ["signing", "key encipherment", "server auth"],
        "expiry": "8760h"
      },
      "client": {
        "usages": ["signing","key encipherment","client auth"],
        "expiry": "8760h"
      }
    }
  }
}
EOF
```

Generate the CA key and certificate

```sh
cfssl print-defaults csr | cfssl gencert -initca - | cfssljson -bare harp-ca
```

#### Generate server certificate

```sh
$ echo '{}' | cfssl gencert -ca=harp-ca.pem -ca-key=harp-ca-key.pem -config=ca-config.json \
    -hostname="secrets.customer1.aws.elstc.co,localhost,127.0.0.1" -profile server - | cfssljson -bare server
```

#### Generate client certificate

```sh
$ echo '{}' | cfssl gencert -ca=harp-ca.pem -ca-key=harp-ca-key.pem -config=ca-config.json \
    -hostname="localhost,127.0.0.1" -profile client - | cfssljson -bare client
```

#### Start fake Vault server with TLS settings

Take a look to settings via the command

```sh
$ harp config new --env
...
export HARP_SERVER_VAULT_LISTEN=":8200"
export HARP_SERVER_VAULT_NETWORK="tcp"
export HARP_SERVER_VAULT_TLS_CACERTIFICATEPATH=""
export HARP_SERVER_VAULT_TLS_CERTIFICATEPATH=""
export HARP_SERVER_VAULT_TLS_CLIENTAUTHENTICATIONREQUIRED="false"
export HARP_SERVER_VAULT_TLS_PRIVATEKEYPATH=""
export HARP_SERVER_VAULT_USETLS="false"
...
```

In order to setup the fake vault server to use our freshly new pki, you have to
set environment variables.

```sh
export HARP_SERVER_VAULT_TLS_CACERTIFICATEPATH="./pki/harp-ca.pem"
export HARP_SERVER_VAULT_TLS_CLIENTAUTHENTICATIONREQUIRED="true"
export HARP_SERVER_VAULT_USETLS="true"
export HARP_SERVER_VAULT_TLS_CERTIFICATEPATH="./pki/server.pem"
export HARP_SERVER_VAULT_TLS_PRIVATEKEYPATH="./pki/server-key.pem"
```

Then start the fake Vault server

```sh
harp-server vault \
  --namespace root:bundle://$(pwd)/customer1-production.bundle
```

#### Use vault CLI

> More details about CLI [environment variables](https://www.vaultproject.io/docs/commands).

Setup vault CLI to enabe TLS client authentication

```sh
export VAULT_ADDR=https://127.0.0.1:8200
export VAULT_CACERT=harp-ca.pem
export VAULT_CLIENT_CERT=client.pem
export VAULT_CLIENT_KEY=client-key.pem
```

Read secret value

```sh
$ vault kv get app/production/customer1/ece/v1.0.0/userconsole/database/usage_credentials
====== Data ======
Key         Value
---         -----
dbname      userconsole
host        sample-instance.abc2defghije.us-west-2.rds.amazonaws.com
options     sslmode=require&application_name=userconsole
password    M0NBU0V4bHFVaXloNnl4I2lDanozXWpHTFdEOj5iTzlNdHx0L0M3cXtPVDZkajNtXjVDKjl1eTNaemVUVWldaw==
port        5432
username    dbuser-userconsole-t2me7nrH
```

#### Use consul-template

Setup vault CLI to enabe TLS client authentication

```sh
export VAULT_ADDR=https://127.0.0.1:8200
export VAULT_CACERT=harp-ca.pem
export VAULT_CLIENT_CERT=client.pem
export VAULT_CLIENT_KEY=client-key.pem
```

Prepare a template

```sh
$ cat << EOF > database.yaml
---
{{\$quality := (or (env "QUALITY") "production") -}}
{{\$quality}}:{{with (printf "app/%s/customer1/ece/v1.0.0/userconsole/database/usage_credentials" \$quality) | secret }}
    adapter: postgresql
    host: {{ .Data.data.host }}
    username: {{ .Data.data.username }}
    password: {{ .Data.data.password | base64Decode }}
{{end}}
EOF
```

Execute the template and merge values

```sh
QUALITY=production consul-template -template "database.yaml:out.txt" -once
```

This will produce the following result

```sh
$ cat out.txt
---
production:
    adapter: postgresql
    host: sample-instance.abc2defghije.us-west-2.rds.amazonaws.com
    username: dbuser-userconsole-t2me7nrH
    password: 3CASExlqUiyh6yx#iCjz3]jGLWD:>bO9Mt|t/C7q{OT6dj3m^5C*9uy3ZzeTUi]k
```
