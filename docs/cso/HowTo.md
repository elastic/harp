# Cloud Secret Storage HOWTO
## or "Where do I put this darn thing, and how do I get it back out?"

## Introduction

Storing secrets (see below) is hard. Not from a purely technical standpoint - there are many, many secure ways to store them - but in the sense that secrets need to be saved in ways that they can be found logically by humans and machines. The Cloud Secret Organization specification is a way to standardize how secrets are stored logically. This document bridges the gap between the high level explanations and the daily practice of storing, retrieving, and using secrets.

### Terms used in this document

 - `secret` is any piece of information that allows access to a system. It can be a username/password combination, a key that unlocks a cryptographic store, a security certificate. It can be plain text, an encrypted blob (such as an SSL Secret Key), or some other piece of information.
 - `secret store` (or just `store`) is a file or a software program that saves and stores `secrets` in a secure manner
 - `key` is a string used to unlock a `secret store`
 - `bundle` is a file-based `secret store`. It can be used stand-alone, or as an intermediate step before populating a server-based `secret store`
 - `component` is a piece of software or a add-on to a piece of software
 - `provider` is a cloud provider - one of `aws`, `azure`, `gcp`, `ibm`, or `local`
 - `region` is a region as defined by a `provider`, or `global` to refer to all regions for that `provider`
 - `version` is a software version using semver notation - that is `vX.Y.Z` where `X` is the major release, `Y` is the minor release, and `Z` is the "patch" or "build" version.
 - `service` is an online service as defined by a cloud provider or SaaS company.
 - `quality` refers to either `dev`, `staging`, `qa`, or `production`

### Why a standard structure?

Most modern secret stores have basically one common factor: there is no standard structure. A blank store is generated and the initial administrator is able to set it up however they like. As more people and teams use it, they may not like how it was structured, and build their own structure. Small groups might not even use a structure, and just store everything at the top level, while others have a directory or document per team. After a period of time (usually a very short one) finding a piece of information becomes rather cumbersome. This also often leads to situations where there are duplicates, out-of-sync duplicates, and outdated secrets in a store.

The purpose of the Cloud Secret Organization (CSO) is to allow for a common structure (dare I say schema) for small, medium, and large teams to store and retrieve secrets in a way that makes sense and defines storeage by the purpose and use of the secret, and not the team or department that places it there initially. This also reduces overhead from maintenance, as a secret need not be moved if responsability for an application or platform changes hands.

### What do I need to get started?

 - [Elastic Harp](https://github.com/elastic/harp)
 - A set of secrets

## Secrets Organization (or where do I find things)

Before we actually store things, we need to figure out where to put them. There are six base directories for storing things: `meta`, `infra`, `platform`, `product`,`app`, and `artifact`. Under each of those is a specifier for the category of secret it is. This category might be the product name under `product` or the tier of an application. Below that the structure gets specialized for the secret types. The sections below will give both the structure and examples for each one.

### `meta` - Secrets that can compromise the entire system

The `meta` directory is used to store secrets that can compromise the entire system. This includes things like the access keys for HashiCorp Vault, metadata for the CSO itself, and so on. The basic structure is `/meta/[category]/[secret]`.

 - `/meta/vault/authentication/oidc_client_json`
 - `/meta/vault/authentication/google_client_json`
 - `/meta/vault/CSO/metadata`

### `infra` - secrets that relate to the systems and infrastructure

The `infra` directory is used to store secrets that relate to infrastructure - cloud service providers, DNS information, data center information, ssh keys, and so on. The basic structure is `/infra/[provider]/[account]/[region]/[service]/[secret]`.

 - `/infra/local/example.com/us-east/ssh/private-key`
 - `/infra/aws/elastic-cloud.com/us-east-1/route53/api-key`
 - `/infra/gcp/elastic-cloud.com/us-west1/compute/ssh-private-key`
 - `/infra/azure/elastic-cloud.com/global/alerting/api-user`

### `platform` - platform credentials

The `platform` directory is used for storing secrets relating to the software running on the infrastructure. The basic structure is `/platform/[quality]/[name]/[region]/[componet]/[secret]`.

 - `/platform/dev/example.com/us-east-1/admin/username`
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

## Storing Secrets

### How do I tell harp what to store?

Spec Files and Templates

### Importing Secrets from other places

using harp --in spec.yaml -f values.yaml & etc

## Retreiving Secrets

### Getting things out of vault
