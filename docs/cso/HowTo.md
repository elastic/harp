# Cloud Secret Storage HOWTO
## or "Where do I put this darn thing, and how do I get it back out?"

## Introduction

Storing secrets (see below) is hard. Not from a purely technical standpoint - there are many, many secure ways to store them - but in the sense that secrets need to be saved in ways that they can be found logically by humans **and** machines. The Cloud Secret Organization (CSO) specification is a way to standardize how secrets are stored logically, and the `harp` tool is designed to aid in storing, retrieving, and validating secrets in a CSO compliant manner. This document provides a simple and practical introduction to CSO and storing, retrieving, and using secrets with `harp`.

### Terms used in this document

 - `secret` is any piece of information that allows access to a system. It can be a username/password combination, a key that unlocks a cryptographic store, a security certificate. It can be plain text, an encrypted blob (such as an SSL Secret Key), or some other piece of information.
 - `secret store` (or just `store`) is a file or a software program that saves and stores `secrets` in a secure manner
 - `key` is a string used to unlock a `secret store`
 - `bundle` is a file-based `secret store`. It can be used stand-alone, or as an intermediate step before populating a server-based `secret store`
 - `component` is a piece of software or a add-on to a piece of software
 - `provider` is a cloud provider - one of `aws`, `azure`, `gcp`, `ibm`, or `local`
 - `region` is a region as defined by a `provider`, or `global` to refer to all regions for that `provider`. The `local` provider only supports the `global` region.
 - `version` is a software version using semver notation - that is `vX.Y.Z` where `X` is the major release, `Y` is the minor release, and `Z` is the "patch" or "build" version.
 - `service` is an online service as defined by a cloud provider or SaaS company.
 - `quality` refers to either `dev`, `staging`, `qa`, or `production`

### Why a standard structure?

Most modern secret stores have basically one common factor: there is no standard structure. A blank store is generated and the initial administrator is able to set it up however they like. As more people and teams use it, they may not like how it was structured, and build their own structure. Small groups might not even use a structure, and just store everything at the top level, while others have a directory or document per team. After a period of time (usually a very short one) finding a piece of information becomes rather cumbersome. This also often leads to situations where there are duplicates, out-of-sync duplicates, and outdated secrets in a store.

The purpose of the Cloud Secret Organization (CSO) is to allow for a common structure (dare I say schema) for small, medium, and large teams to store and retrieve secrets in a way that makes sense and defines storeage by the purpose and use of the secret, and not the team or department that places it there initially. This also reduces overhead from maintenance, as a secret need not be moved if responsability for an application or platform changes hands.

### What do I need to get started?

 - [Elastic Harp](https://github.com/elastic/harp)
 - A set of secrets
 - A text editor, like `vim`
 - `jq` and `yq` for validations

## Secrets Organization (or where do I find things)

Before we actually store things, we need to figure out where to put them. There are six base directories for storing things: `meta`, `infra`, `platform`, `product`,`app`, and `artifact`. Under each of those is a specifier for the category of secret it is. This category might be the product name under `product` or the tier of an application. Below that the structure gets specialized for the secret types. The sections below will give both the structure and examples for each one.

### `meta` - Secrets that can compromise the entire system

The `meta` directory is used to store secrets that can compromise the entire system. This includes things like the access keys for HashiCorp Vault, metadata for the CSO itself, and so on. The basic structure is `/meta/[category]/[secret]`.

 - `/meta/vault/authentication/oidc_client_json`
 - `/meta/vault/authentication/google_client_json`
 - `/meta/vault/CSO/metadata`

### `infra` - secrets that relate to the systems and infrastructure

The `infra` directory is used to store secrets that relate to infrastructure - cloud service providers, DNS information, data center information, ssh keys, and so on. The basic structure is `/infra/[provider]/[account]/[region]/[service]/[secret]`.

 - `/infra/local/example.com/global/ssh/private-key`
 - `/infra/aws/elastic-cloud.com/us-east-1/route53/api-key`
 - `/infra/gcp/elastic-cloud.com/us-west1/compute/ssh-private-key`
 - `/infra/azure/elastic-cloud.com/global/alerting/api-user`

### `platform` - platform credentials

The `platform` directory is used for storing secrets relating to the software running on the infrastructure. The basic structure is `/platform/[quality]/[name]/[region]/[componet]/[secret]`.

 - `/platform/dev/example.com/global/admin/username`
 - `/platform/production/elastic-cloud.com/aws-us-west-1/rds/password`
 - `/platform/production/elastic-cloud.com/gcp-us-east1/rds/password`
 - `/platform/production/elastic-cloud.com/azure-eastus2/rds/password`
 - `/platform/qa/elastic-cloud.com/global/admin/okta-api-token`

### `product` - product keys and registration information

The `product` directory is used for product-related secrets. The basic structure is `/product/[name]/[version]/[component]/[secret]`.

 - `/product/macosx/v11.0.2/antivirus/enterprise/registration-key`
 - `/product/elastic/v7.10.1/xpack/license/serialnumber`
 - `/product/kibana/v7.9.0/auth/okta-plugin/api-token-secret`
 - `/product/redhat/v8.3.0/rhn/login/username`
 - `/product/redhat/v8.3.0/rhn/login/password`

### `app` - application specific credentials

The `app` directory is used for application secrets. The basic structure is `/app/[quality]/[platform]/[product]/[version]/[component]/[secret]`.

 - `/app/dev/ubuntu/v18.04/fips/ppa/url`
 - `/app/dev/ubuntu/v18.04/fips/ppa/username`
 - `/app/dev/ubuntu/v18.04/fips/ppa/password`
 - `/app/qa/rundeck/v3.3.7/web/admin/password`
 - `/app/production/rundeck/v3.3.7/web/admin/password`

### `artifact` - checksums and reports

The `artifact` directory is used for release artifacts. The basic structure is `/artifact/[component]/[unique id]/[secret]`

 - `/artifact/docker/sha256:[some sha256sum]/security/attestations/scan_report`
 - `/artifact/iso/md5:[some md5sum]/installation/product-key`

## Storing (and Retreiving) Secrets

Now that we have some idea how to structure the secrets, we need to store them. The tool `harp` uses yaml and json to determine what to store based on the structure above.

### How do I tell `harp` what to store?

`Harp` uses specification documents called `spec` files to determine what to store where. These can be formatted as either YAML or JSON. The format allows `harp` to validate both the location (based on the directory paths above) and the format of a secret. If needed, the `spec` file will instruct `harp` on how to generate a secret based on specific criteria if one does not exist. It is possible, with the right syntax, to populate a completely blank `secret store` with all the values needed for a new service deployment automatically.

The spec file is documented in full in the [https://github.com/elastic/harp/tree/main/docs/onboarding/1-template-engine](Harp Onboarding Docs). There is a LOT in there, and we cannot possible cover all the cases in this document.

### A Sample set of secrets

Let us say that we have a set of secrets to store. We know the secrets, but we need to represent them in a way that `harp` can read in. In this case, we will use a YAML file name `values.yaml`.

```yaml
secrets:
  infra:
    local:
      account: "example.com"
      global:
        ssh_private_key: privatekeystring
    aws:
      account: "elastic-cloud.com"
      us_east_1:
        route53:
          api_key: awsapikey
  platform:
    production:
      elastic_cloud.com:
        aws_us_west_1:
          rds:
            password: mypassword
        gcp-us-east1:
          rds:
            password: myotherpassword
        azure-eastus2:
          rds:
            password: azurepassword
  app:
    dev:
      ubuntu:
        v18.04:
          fips:
            ppa:
              url: "https://ppa.ubuntulinux.net/fips"
              username: someuser
              password: somepassword
```
### A sample `spec.yaml` file

The following `spec.yaml` defines the location of several secrets and the values to populate them with.

```yaml
apiVersion: harp.elastic.co/v1
kind: BundleTemplate
meta:
  name: samplespec
  owner: cloud-operations@elastic.co
  description: Keys for AWS
spec:
  namespaces:
    infrastructure:
      - provider: local
        description: Common keys for local resources
        account: "example.com"
        regions:
          - name: global
            services:
              - type: compute
                name: ssh
                description: ssh keys
                secrets:
                  - suffix: private_key
                    description: "SSH Private Key"
                    template: |-
                      {
                          "private_key": "{{ .Values.secrets.infra.local.global.ssh_private_key }}"
                      }
      - provider: aws
        description: aws keys
        account: "elastic-cloud.com"
        regions:
          - name: us-east-1
            services:
              - type: compute
                name: route53
                description: route53 api key
                secrets:
                  - suffix: apikey
                    description: "Route53 API Key"
                    template: |-
                      {
                          "api_key": "{{ .Values.secrets.infra.aws.us_east_1.route53.api_key }}"
                      }
```

### OK, Now what?

Great question! Now it is time to actually store those secrets in a locked bundle.

### Working with Bundles

We can start by making an unlocked bundle.

```bash
harp from bundle-template --in example.spec.yaml --values example.values.yaml --out example.bundle
```

This will create an unencrypted file `example.bundle` in the local directory. If there are any syntax errors in the spec file or the values file, harp **will** exit with an error. Using a tool like `yq` as a linter on the input files can be very informative to at least verify the syntax and structure of the files themselves.

Be aware that the errors are not always clear at first, and you may need to ask for help to understand them.

You can look an an unencrypted bundle with the `bundle dump` command in `harp`. This will return a JSON representation of all the data in the bundle file, including the template it was generated from. This can be used to verify the data befor encrypting.

```bash
harp bundle dump --in example.bundle
```

Once a bundle is created, it can be encrypted with a base64 encoded `key`.

Be aware that the command line below is NOT secure. You should probably use a `secret store` or some other secure tool for generating and storing the `key`.

```bash
harp bundle encrypt --in example.bundle --out example.encrypted.bundle --key [some base64 string]
```

In order to decerypt the bundle, you can use the same key with the `decrypt` command.

```bash
harp bundle decrypt --in example.encrypted.bundle --out example.bundle --key [some base64 string]
```

Since the `--in` and `--out` commands can also read and write to the console with `-` as the target, you can pipe one bundle command to the next like so.

```bash
harp from bundle-template --in example.spec.yaml --values example.values.yaml --out - | harp bundle encrypt --in - --out example.encrypted.bundle --key [some base64 string]
```

This creates a new bundle and pipes the data into the encrypt command, resulting in an encrypted bundle.

If we want to be SUPERFANCY, we can add some `jq` parsing to the `decrypt` command and pull a single value out.

```bash
harp bundle decrypt --in example.encrypted.bundle --out - --key [some base64 string] | harp bundle dump --in - | jq '.packages|map(select(.name == "infra/aws/elastic-cloud.com/us-east-1/compute/route53/apikey"))| .[].secrets.data[].value'
```

