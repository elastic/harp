# Harp : From Zero to Hero

## Requirements

### Tools

```sh
brew install vault # Hashicorp Vault CLI and Server
brew install jq    # Use to query / parse / beautify json data
brew install yq    # Use to query / parse yaml data
```

To install `stable` harp :

```sh
brew install elastic/harp/harp # Install harp
```

To install `devel` harp (compiled from `main` branch)

```sh
brew install --head elastic/harp/harp # Install harp
```

### Environment

In a dedicated console, start a Vault server in developer mode

```sh
VAULT_DEV_ROOT_TOKEN_ID=myroot vault server -dev
```

In each console opened, don't forget to add these environment variables :

```sh
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=myroot
```

Prepare Vault secrets backends :

```sh
vault secrets enable -version=2 -path=infra kv
vault secrets enable -version=2 -path=platform kv
vault secrets enable -version=2 -path=product kv
vault secrets enable -version=2 -path=app kv
vault secrets enable -version=2 -path=artifact kv
vault secrets enable -version=2 -path=legacy kv
```

## Let's go

### Template engine

1. [Introduction](1-template-engine/1-introduction.md)
1. [Functions](1-template-engine/2-functions.md)
1. [Variables](1-template-engine/3-variables.md)
1. [Values](1-template-engine/4-values.md)
1. [Files](1-template-engine/5-files.md)
1. [Lists and Maps](1-template-engine/6-lists-and-maps.md)
1. [Alternative delimiters](1-template-engine/7-alternative-delimiters.md)
1. [Whitespace controls](1-template-engine/8-whitespace-controls.md)
1. [Use Cases](1-template-engine/9-usecases.md)

### Secret Container

1. [Introduction](2-secret-container/1-introduction.md)
2. [Specifications](2-secret-container/2-specifications.md)
3. [Cryptographic seal](2-secret-container/3-seal.md)

### Secret Bundle

1. [Introduction](3-secret-bundle/1-introduction.md)
2. [Bundle](3-secret-bundle/2-bundle.md)
3. [BundleTemplate](3-secret-bundle/3-template.md)
4. [BundlePatch](3-secret-bundle/4-patch.md)

### Value Transformers

1. [Introduction](4-value-transformers/1-introduction.md)
2. [Encryption](4-value-transformers/2-encryption.md)
3. [Signature](4-value-transformers/3-signature.md)
4. [Signing Workflows](4-value-transformers/4-signing-workflows.md)

### Secret Workflow

> Section in development
