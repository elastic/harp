# Harp : Crate server plugin

Allow `secret crates` to be exposed using :

* Simple `HTTP` API
* `Vault read-only` server API
* `gRPC` microservice - [Service definition](../../api/proto/harp/bundle/v1/bundle_api.proto)

## Use Case

When you have a central secret storage and you want to decorellate secret storage
from multi region secret consuming.

You could have one Vault cluster for internal secret secret storage, apply
secret deployment from Vault on secret changes by using CD workflows :

* to update relevant resources (update Database password etc.)
* produce `secret containers`
* deploy them in cloud storage

So that containers could be consumed and exposed  :

* via an HTTP server,
* or a fake Vault KV only server,
* or a gRPC microservice.

## Install

### From homebrew

```sh
brew tap elastic/harp
brew install elastic/harp/harp-server
```

### From source

```sh
export PATH=$HARP_REPO/tools/bin:$PATH
go mod vendor
mage
```

## Namespaced secrets

One backend is bound to a `namespace`. In order to register a namespace, you
have to use the following syntax :

```sh
--namespace <name>:<backend-factory-url>
```

## Transformer API

You can expose a transformer using Vault Transit HTTP API.

```sh
--transformer <name>:<key> (from harp keygen)
```

Example :

```sh
$ harp server vault \
  --transformer fernet:$(harp keygen fernet) \
  --transformer aes256-gcm96:$(harp keygen aes-256) \
  --transformer secretbox:$(harp keygen secretbox)
```

It will initalize a transformer with given name as Vault Key with given value as Key.

```sh
$ export VAULT_ADDR=http://127.0.0.1

# To encrypt
$ vault write transit/encrypt/secretbox plaintext=$(base64 <<< "my secret data")
Key           Value
---           -----
ciphertext    vault:v1:XHcLYFthNzR70a5gs/i4stRfxAkL8RXTF0U4oX13/gx8ftXZhTPMejjN+fChaemmOfYVwxeNeQ==

# To decrypt
$ vault write -format=json transit/decrypt/secretbox ciphertext=vault:v1:XHcLYFthNzR70a5gs/i4stRfxAkL8RXTF0U4oX13/gx8ftXZhTPMejjN+fChaemmOfYVwxeNeQ== | jq -r ".data.plaintext" | base64 -D
my secret data
```

> Don't forget to protect `harp-server` with a TLS configuration to protect
> your secret when accessing over network.

### Local

* `file`: directly serve a file content

### Cloud storage

> Queries to server are directly proxified to backend storage to serve content
> from cloud storage.

* `s3`: S3 bucket loader
* `azblob` : Azure Blob storage
* `gcs`: Google bucket

#### S3

URL Pattern : `s3://<endpoint>/<bucketName>/<objectKey>`

> Leave `<endpoint>` blank for AWS targeted S3 bucket, specify if you want to
> use alternative S3 compatible API endpoints (Minio, IBM, etc.)

Parameters :

* `access-key-id` (string, default "" and loaded from envrionment) : set explicitly the AWS_ACCESS_KEY_ID value
* `profile` (string, default ""): set explicitky the AWS_PROFILE value
* `region` (string, default ""): set explicitly the AWS_REGION value
* `secret-access-key` (string, default ""): set explicitly AWS_SECRET_ACCESS_KEY value
* `session-token` (string, deafult ""): set explicitly AWS_SESSION_TOKEN value
* `disable-ssl` (bool, default "false"): disable SSL usages
* `env-authentication` (bool, default "true"): force use envrionment as credentials source
* `ignore-config-creds` (bool, default "false"): ignore shared configuration creadential provider
* `ignore-ec2role-creds` (bool, default "false"): ignore EC2Role credential provider
* `ignore-env-creds` (bool, default "false"): ignore environment credentials provider
* `s3-force-path-style` (bool, default "false"): Force path style
* `s3-use-accelerate-endpoint` (bool, default "false"): Use accelerated endpoint protocol

##### AWS S3

```sh
s3:///harp/secrets?profile=bootstrap
```

##### AWS S3 (FIPS)

```sh
s3://s3-fips.us-east-1.amazonaws.com/harp/secrets?
  profile=bootstrap # Use given named profile from shared config
```

##### IBM COS

```sh
s3://s3.eu-de.cloud-object-storage.appdomain.cloud/harp/secrets?
  region=eu-de # Explicit region
  &access-key-id=$ACCESS_KEY
  &secret-access-key=$SECRET_KEY
```

##### Minio - <https://min.io/>

```sh
s3://127.0.0.1:9000/harp/secrets?
  disable-ssl=true # Disable SSL
  &access-key-id=$ACCESS_KEY
  &secret-access-key=$SECRET_KEY
  &s3-force-path-style=true # Required to support (<endpoint>/<bucketName>) path
```

#### Google Cloud Storage (gcs)

URL Pattern : `gcs://<bucketName>/<objectKey>`

Parameters :

* `prefix` (string, default "") sets the object key prefix before quering the
  cloud storage.

#### Azure Blob Storage (azblob)

URL Pattern : `azblob://<bucketName>/<objectKey>`

Environment variables :

* `AZURE_CONNECTION_STRING` (string, default "") sets the azure connection string
  to use for this backend.

Parameters :

* `prefix` (string, default "") sets the object key prefix before quering the
  cloud storage.

### Vault proxy

> Expose a Vault secret tree from server with unified API.

URL Pattern : `vault://<path>`

* `vault`: expose a secret tree from given path from Vault.

Environment variables :

* `VAULT_*` all Vault environment variables used by CLI.

### Remote bundle loader

> Load a remote bundle from selected stroage, and serve it from memory.

* `bundle` / `bundle+file` from a local bundle file
* `bundle+http` from a remote HTTP server hosted bundle file
* `bundle+https` from a remote HTTPS server hosted bundle file
* `bundle+s3` from a remote S3 bucket hosted bundle file
* `bundle+gcs` from a remote GCS bucket hosted bundle file
* `bundle+azblob` from a remote Azure Blob hosted bundle file
* `bundle+stdin` from a stdin container

It uses the same parameters as the direct file serving process,but it uses a
secret container as `<objectKey>` to retrieve it and use it for memory content
serving.

Parameters :

* `cid` (string, default "") sets the Container key to use to unseal a sealed
  container. Keys from process keyring will be used too.

## Storage transformers

> Apply content transformation before serving content to client.

* `fernet` to apply fernet encryption / decryption
* `secretbox` to apply Nacl SecretBox encryption / decryption
* `aes-gcm` to apply aes256-gcm96 encryption / decryption

Parameters :

* `key` (string) is the symmetric encryption key to use generated by `harp keygen`;
* `enc_revert` (bool, default "false") defines the transformer usage :
  * "false" => apply transformation (encode, encrypt, etc.)
  * "true" => apply reverse tranformation (decode, decrypt, etc.)

## Implementations

### Common

You can specify settings for the chosen listener you want to use, by using a
`config` file or using environment variables.

```sh
$ harp-server config new --env
export HARP_SERVER_BACKENDS="[]"
export HARP_SERVER_DEBUG_ENABLE="false"
export HARP_SERVER_INSTRUMENTATION_DIAGNOSTIC_CONFIG_GOPS_ENABLED="false"
export HARP_SERVER_INSTRUMENTATION_DIAGNOSTIC_CONFIG_GOPS_REMOTEURL=""
export HARP_SERVER_INSTRUMENTATION_DIAGNOSTIC_CONFIG_PPROF_ENABLED="true"
export HARP_SERVER_INSTRUMENTATION_DIAGNOSTIC_CONFIG_ZPAGES_ENABLED="true"
export HARP_SERVER_INSTRUMENTATION_DIAGNOSTIC_ENABLED="false"
export HARP_SERVER_INSTRUMENTATION_LISTEN=":5556"
export HARP_SERVER_INSTRUMENTATION_LOGS_LEVEL="warn"
export HARP_SERVER_INSTRUMENTATION_NETWORK="tcp"
export HARP_SERVER_KEYRING="[]"
... Servers related settings ...
```

#### Listener settings

If you look at the `HTTP` REST API settings :

```sh
# Sets the address used to start the listener
export HARP_SERVER_HTTP_LISTEN=":8080"
# Defines the network for the listener (tcp, tcp4, tcp6, unixsocket)
export HARP_SERVER_HTTP_NETWORK="tcp"
# Defines CA certificate for mTLS authentication (client authentication)
export HARP_SERVER_HTTP_TLS_CACERTIFICATEPATH=""
# Defines Certificate used for server authentication
export HARP_SERVER_HTTP_TLS_CERTIFICATEPATH=""
# Enforce client authentication
export HARP_SERVER_HTTP_TLS_CLIENTAUTHENTICATIONREQUIRED="false"
# Defines private key path
export HARP_SERVER_HTTP_TLS_PRIVATEKEYPATH=""
# Enable TLS to the given listener
export HARP_SERVER_HTTP_USETLS="false"
```

For `gRPC` listener, replace `HARP_SERVER_HTTP` by `HARP_SERVER_GRPC`.
For `Vault` listener, replace `HARP_SERVER_HTTP` by `HARP_SERVER_VAULT`.

## Secret API

### HTTP

Expose a REST API to retrieve secrets that match the given path :

```html
GET /api/v1/<namespace>/<path>
```

### Vault

Expose a Vault Server compatible API with read-only KV support.

### gRPC

Expose a gRPC (HTTP2/Protobuf) server.

## Sample server settings

### Preparation

Generate pre-shared encryption key

```sh
export CLIENT_PSK=$(harp keygen fernet)
export SERVER_PSK=$(harp keygen fernet)
```

Prepare a small `Bundle` from JSON  (secrets.json):

```sh
{
    "app/{{.Values.quality}}/ops/harp/v1.0.0/bootstrap/cluster/elasticsearch/root": {
        "user": "root-{{ randAlpha 8 }}",
        "password": {{ strongPassword | toJson }}
    },
    "app/{{.Values.quality}}/ops/harp/v1.0.0/bootstrap/cluster/elasticsearch/certificates": {
        "cacert.pem": "{{ .Files.Get "ca.pem" }}"
    }
}
```

### HTTP + S3 + E2E Encryption

Publish secret to s3

```sh
# Generate encrypted content
$ harp template --in secrets.json -set quality=production
  | harp transform encryption --key $CLIENT_PSK \
  | harp transform encryption --key $SERVER_PSK --out bootstrap
# Publish to S3
$ aws s3 cp ./bootstrap s3://MY-AWESOME-SECRET-BUCKET/secrets
```

Start the server

```sh
harp server http \
  --namespace root:s3:///MY-AWESOME-SECRET-BUCKET/secrets?key=$SERVER_PSK&enc_revert=true
```

The server does not have the full decrypted version in memory. But decrypt with
its own PSK before sending content to client (encrypted with $CLIENT_KEY).

> Pay attention to `s3://<s3-api-endpoint>/` we endpoint is empty it will use default
> AWS endpoint.

Client side decryption :

```sh
$ curl http://127.0.0.1/api/v1/root/bootstrap
  | harp transform encryption --key $CLIENT_PSK --revert
{
    "user": "root-CeQQnxAb",
    "password": "5\u003eH=?(o1k0T/R5?crok28J\u003e+!9y4~2L0"
}
```

### HTTP + mTLS + PSK encryption + unixsocket

> A little paranoid sample (`who can do the most can do the least`)

Start the server

```sh
# Setup listener
export HARP_SERVER_HTTP_LISTEN="/tmp/harp-server.sock"
export HARP_SERVER_HTTP_NETWORK="unixsocket"
export HARP_SERVER_HTTP_USETLS="true"
export HARP_SERVER_HTTP_TLS_CACERTIFICATEPATH="ca.crt"
export HARP_SERVER_HTTP_TLS_CERTIFICATEPATH="server.crt"
export HARP_SERVER_HTTP_TLS_PRIVATEKEYPATH="server.key"
export HARP_SERVER_HTTP_TLS_CLIENTAUTHENTICATIONREQUIRED="true"
# Start a crate server using a generated one
harp template --in secrets.json -set quality=production \
  | harp from jsonmap
  | harp server http \
    --namespace root:bundle+stdin://?key=$PSK
```

On client side :

```sh
curl --unix-socet /tmp/secrethub-server.sock \
  --cacert ca.crt \
  --key client.key \
  --cert client.crt \
  http://127.0.0.1/api/v1/root/app/production/ops/harp/v1.0.0/bootstrap/cluster/elasticsearch/root
  | harp transform encryption --key $PSK --revert
{
    "user": "root-CeQQnxAb",
    "password": "5\u003eH=?(o1k0T/R5?crok28J\u003e+!9y4~2L0"
}
```

### Vault + Multiple namespaces

Start the server

```sh
harp server vault \
  --namespace security:bundle://`pwd`/sealed.bundle?cid=$CONTAINER_KEY \
  --namespace legacy:s3:///bootstrap/secrets?key=$PSK&env_revert=true # Decrypt bundle before sending to client
```

Client side

```sh
export VAULT_ADDR=http://127.0.0.1:8200
# Query security namespace to use secret container
vault kv get -namespace=security platform/aws/foo/eu-central-1/rds/posgresql/adminconsole/root_creds
# Query bootstrap namespace to use s3 stored secrets
vault kv get -namespace=bootstrap secrets/simple/env
```
