# BundleTemplate

A `BundleTemplate` allows developers to describe their own secret requirements by
designing secret values with `harp` template engine attached to a CSO
compliant secret path.

A `BundleTemplate` doesn't contains secret values, and could be parametrized for
common secret generation profile (region bound bundle).

It helps `developers` and `secret operators` to have a common language to
manage secrets in time, and with auditable requirements that could be stored as
code, `BundleTemplate is a Secret as Code definition`. It will be used as
specification to generate a secret container with a `Bundle` containing rendered
secret paths and matching values.

## Specifications

This [specification](https://github.com/elastic/harp/blob/main/api/proto/harp/bundle/v1/template.proto) declares Bundle generation template object structure.

The specification is written in YAML, in order to be human readable, but
internally converted as protobuf object.

### CSO Secret Path Naming Convention

Cloud Secret Organization is the secret management program deployed at Elastic, to
solve secret management problems.

One of them concerns the secret storage, and more precisely the secret path
building process.

Too many people fall in the anti-pattern which consist in organizing secrets
by team secrets ownership, or mirroring the company organization tree but
this is wrong ! Secrets are just managed by teams and owned by softwares.

> The policy applied for secret access authorization can be considered as
> `proof of ownership`, not the fact that the secret is stored in a folder
> nammed by the team name.
>
> Don't forget that a company is an organic life form that is in constant move.
> It's easier to modify policies than redesigning a complete secret tree.

We designed a risk-based classification ordered in "rings". Each ring evaluate
a threat realization impact level when a ring secrets leak occurs.

* `meta` (R0) - Secret used to protect secret storage (Vault authentication, etc.)
* `infrastructure` (R1) - Infrastructure secrets can lead to multiple platform compromise.
* `platform` (R2) - Platform secrets can lead to multiple product compromise.
* `product` (R3) - Product related secrets shared by all instance of the same product.
* `application` (R4) - Application secrets are instancied product in a given platform for a given quality stage.
* `artifact` (R5) - Artifact secrets are secrets attached to an artifact (docker image, archive, etc.)

When building a `BundleTemplate`, you have to distribute the secrets accross the rings
according to their sensitivity to leaks.

> The CSO CheatSheet can be viewed here - <https://ela.st/cso>

### YAML

`BundleTemplate` uses Kubernetes-like YAML descriptors.

```yaml
apiVersion: harp.elastic.co/v1
kind: BundleTemplate
meta:
  name: "Ecebootstrap"
  owner: syseng@elstc.co
  description: "ECE Secret Bootstrap"
spec:
  namespaces:
    infrastructure:
  ... Omitted ...
```

* The `apiVersion` must be `harp.elastic.co/v1`
* The `kind` must be `BundleTemplate`
* The `meta.name` must not be empty
* The `meta.owner` must be an email
* The `meta.description` must not be empty
* The `spec` describes the concrete `BundleTemplate` secret template according
  namespaces (rings) used.

#### Infrastructure secrets

> Protobuf definition - [InfrastructureNS](https://github.com/elastic/harp/blob/main/api/proto/harp/bundle/v1/template.proto#L83)

An `infrastructure` is a consitent set of resource provided by an infrastructure
provider (IaaS).

`Infrastructure secrets` are sensitive secrets attached to these resources.

```yaml
spec:
  namespaces:
    infrastructure:
    # AWS ----------------------------------------------------------------------
    - provider: "aws"
      account: "{{ .Values.infra.aws.account }}"
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
            template: |-
              {
                "user": "dbroot-{{ randAlphaNum 8 }}",
                "password": "{{ paranoidPassword | b64enc }}"
              }
```

This will generate a secret like the following one :

```json
{
    "infra/aws/essp-customer1/us-east-1/rds/adminconsole/accounts/root_credentials": {
        "password": "OHxLcGI4WzZuP0V+UFJcdXEkX0M0XXhiSlF6VzFtZjZhZHc3RFFCM1Y4a2x6aHVYclhwcy41Q15QU29ZSmRzdw==",
        "user": "dbroot-Lta1Swae"
    }
}
```

#### Platform secrets

> Protobuf definition - [PlatformNS](https://github.com/elastic/harp/blob/main/api/proto/harp/bundle/v1/template.proto#L117)

A `platform` is a regionalized consistent set of infrastructure resources working
together in order to provide services to products.

> Platform regions are not necessarily infrastrcuture regions. A logical overlay
> of physical infrastructure regions could be designed.

```yaml
spec:
  selector:
    quality: "production"
    platform: "customer1"
  namespaces:
    platform:
      # EMEA-1
      - region: "emea-1"
        components:
        # Zookeeper ------------------------------------------------------------
        - name: "zookeeper"
          secrets:
          - suffix: "accounts/admin_credentials"
            description: "Zookeeper administrative access account"
            template: |-
              {
                "username": "zkadmin-{{ randAlphaNum 8 }}",
                "password": "{{ paranoidPassword | b64enc}}"
              }

        # PostgreSQL -----------------------------------------------------------
        - name: "postgresql"
          secrets:
          - suffix: "admiconsole/admin_credentials"
            description: "PostgreSQL Administrative access on Database only"
            template: |-
              {
                "username": "dbadmin-{{ randAlphaNum 8 }}",
                "password": "{{ paranoidPassword | b64enc }}"
              }

        # Billing --------------------------------------------------------------
        - name: "billing"
          secrets:
          - suffix: "recurly/vendor_api_key"
            description: "Recurly API Key for invoice generation"
            # Not generated by workflow, requested at generation time.
            vendor: true
            template: |-
              {
                "API_KEY": "{{ .Values.vendor.recurly.api_key }}"
              }
            content:
              "ca.pem": |-
                {{ .Files.Get "billing/ca.pem" }}
```

This will generate 3 secrets like the following ones:

```json
{
  "platform/production/customer1/emea-1/billing/recurly/vendor_api_key": {
    "API_KEY": "recurly-foo-api-123456789",
    "ca.pem": "-----BEGIN CERTIFICATE-----\nMIIFazCCA1OgAwIBAgIRAIIQz7DSQONZRGPgu2OCiwAwDQYJKoZIhvcNAQELBQAw\nTzELMAkGA1UEBhMCVVMxKTAnBgNVBAoTIEludGVybmV0IFNlY3VyaXR5IFJlc2Vh\ncmNoIEdyb3VwMRUwEwYDVQQDEwxJU1JHIFJvb3QgWDEwHhcNMTUwNjA0MTEwNDM4\nWhcNMzUwNjA0MTEwNDM4WjBPMQswCQYDVQQGEwJVUzEpMCcGA1UEChMgSW50ZXJu\nZXQgU2VjdXJpdHkgUmVzZWFyY2ggR3JvdXAxFTATBgNVBAMTDElTUkcgUm9vdCBY\nMTCCAiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoCggIBAK3oJHP0FDfzm54rVygc\nh77ct984kIxuPOZXoHj3dcKi/vVqbvYATyjb3miGbESTtrFj/RQSa78f0uoxmyF+\n0TM8ukj13Xnfs7j/EvEhmkvBioZxaUpmZmyPfjxwv60pIgbz5MDmgK7iS4+3mX6U\nA5/TR5d8mUgjU+g4rk8Kb4Mu0UlXjIB0ttov0DiNewNwIRt18jA8+o+u3dpjq+sW\nT8KOEUt+zwvo/7V3LvSye0rgTBIlDHCNAymg4VMk7BPZ7hm/ELNKjD+Jo2FR3qyH\nB5T0Y3HsLuJvW5iB4YlcNHlsdu87kGJ55tukmi8mxdAQ4Q7e2RCOFvu396j3x+UC\nB5iPNgiV5+I3lg02dZ77DnKxHZu8A/lJBdiB3QW0KtZB6awBdpUKD9jf1b0SHzUv\nKBds0pjBqAlkd25HN7rOrFleaJ1/ctaJxQZBKT5ZPt0m9STJEadao0xAH0ahmbWn\nOlFuhjuefXKnEgV4We0+UXgVCwOPjdAvBbI+e0ocS3MFEvzG6uBQE3xDk3SzynTn\njh8BCNAw1FtxNrQHusEwMFxIt4I7mKZ9YIqioymCzLq9gwQbooMDQaHWBfEbwrbw\nqHyGO0aoSCqI3Haadr8faqU9GY/rOPNk3sgrDQoo//fb4hVC1CLQJ13hef4Y53CI\nrU7m2Ys6xt0nUW7/vGT1M0NPAgMBAAGjQjBAMA4GA1UdDwEB/wQEAwIBBjAPBgNV\nHRMBAf8EBTADAQH/MB0GA1UdDgQWBBR5tFnme7bl5AFzgAiIyBpY9umbbjANBgkq\nhkiG9w0BAQsFAAOCAgEAVR9YqbyyqFDQDLHYGmkgJykIrGF1XIpu+ILlaS/V9lZL\nubhzEFnTIZd+50xx+7LSYK05qAvqFyFWhfFQDlnrzuBZ6brJFe+GnY+EgPbk6ZGQ\n3BebYhtF8GaV0nxvwuo77x/Py9auJ/GpsMiu/X1+mvoiBOv/2X/qkSsisRcOj/KK\nNFtY2PwByVS5uCbMiogziUwthDyC3+6WVwW6LLv3xLfHTjuCvjHIInNzktHCgKQ5\nORAzI4JMPJ+GslWYHb4phowim57iaztXOoJwTdwJx4nLCgdNbOhdjsnvzqvHu7Ur\nTkXWStAmzOVyyghqpZXjFaH3pO3JLF+l+/+sKAIuvtd7u+Nxe5AW0wdeRlN8NwdC\njNPElpzVmbUq4JUagEiuTDkHzsxHpFKVK7q4+63SM1N95R1NbdWhscdCb+ZAJzVc\noyi3B43njTOQ5yOf+1CceWxG1bQVs5ZufpsMljq4Ui0/1lvh+wjChP4kqKOJ2qxq\n4RgqsahDYVvTH9w7jXbyLeiNdd8XM2w9U/t7y0Ff/9yi0GE44Za4rF2LN9d11TPA\nmRGunUHBcnWEvgJBQl9nJEiU0Zsnvgc/ubhPgXRR4Xq37Z0j4r7g1SgEEzwxA57d\nemyPxgcYxn/eR44/KJ4EBs+lVDR3veyJm+kXQ99b21/+jh5Xos1AnX5iItreGCc=\n-----END CERTIFICATE-----"
  },
  "platform/production/customer1/emea-1/postgresql/admiconsole/admin_credentials": {
    "password": "aGQyd2k+M3BiUzhjdExjWGQkUzR5WEpFeVI3KXg+ankxYmxGc3ZUWlo3MH1bRnU2LGZVSkxFfjZ4JWZqeih4Vg==",
    "username": "dbadmin-0GISldgQ"
  },
  "platform/production/customer1/emea-1/zookeeper/accounts/admin_credentials": {
    "password": "bUtLJDdnaDJbdHRhRUdwaDBjKmhSNzBvdTNUMmZzZ0dwL1I9WkJreVQ2a0lQdHN2WFRtUSx7WTd6PSZQWU43ew==",
    "username": "zkadmin-VO8HYMFL"
  }
}
```

* `quality` path compoent is extracted from `selector.quality`
* `platform` path component is extracted from `selector.platform`

#### Product secrets

> Protobuf definition - [ProductComponentNS](https://github.com/elastic/harp/blob/main/api/proto/harp/bundle/v1/template.proto#L139)

A `product` in CSO, is a virtual concept for non-instanciable resources, and
not related to execution.

> Same secrets for all platform/stage ? => it could be a product secret.

```yaml
spec:
  selector:
    product: "ece"
    version: "v1.0.0"
  namespaces:
    product:
    - name: artifact
      description: "Artifact related secrets"
      secrets:
      - suffix: "signature/key"
        description: "Artifact signature crypto key encoded using JWK"
        template: |-
          {
            "privateKey": "{{ $sigKey := cryptoPair "rsa" }}{{ $sigKey.Private | toJwk | b64enc }}",
            "publicKey": "{{ $sigKey.Public | toJwk | b64enc }}"
          }
```

This will generate a secret like the following one :

```json
{
  "product/ece/v1.0.0/artifact/signature/key": {
    "privateKey": "eyJrdHkiOiJSU0EiLCJraWQiOiJNczQzbDN4dVBCUzhHemxWeU0yajFGSDlHeTM0MFM0bl9sYjNuNE1nNTkwPSIsIm4iOiJtM0dLRVVhX1VSX2tJM3RXaVJUVnFOQ0pueGh4TjNYSm1rS19yV3ZUeE1rMl9taXVRU29qckpwRmlGTjI4ZWJnRWpZYnkyM2VodFlEQi1SVFVaZU1jVUFZb3VaWnBlYlpsOGhWNjBPT0dmSnBxRmVKTmlnVmxUendkN1Y5RWFlTUVhbDVvUDQwbURneW5ocElaNVNVMzhCd1dvd0Vlb2JPSW8xek11UjFwMktOSmxhT0RpR1hPTDVVVzlkT1dOTVF1VUFUVFNhZElZVnZRYmlwMWZMNl96aU9udkIwdWd2TW1qQjRIdXhEMHVsRlo2TFhpdVBqX0ZkN1g4Y2VMVld0R0doTFBKb09vUFp3LTluTzVwcDRMb1NjZFFXNGxCVjY3MVJQdThCb3JYYUREampqVVVpN2NGemR1eXFnNkhHZWk3UlFpY25POHhhTDhRekxLNUNnUFEiLCJlIjoiQVFBQiIsImQiOiJtZGhLdlFTWFQ1UlB1R3BXNEQtVm80b1oyek5Xd253NmR3bV9LY1hCaDA5YXRYc25rLWtfLTVHSVpmLXRob2RwbDd5ano3aEMtSktSMTFxOHQ5RlZON1VuYlBxdEdZeWNLU1FuSFR6MFJHdnU5S1VHY1dwRXlqclJDTG5BT2h2b2ZvYU1rYkZtbm8xb1U2QlJydXFZV2NmZHExQlBFbkdmVFFWUVVidFpVcnFsWl9HWmM5UGxxV1FnQUk5RFhXSGE5MEI0RzI1WHY2Wi1wWFhhOXpZMUZsQVF2TUo2MkIwUUxaWmppdzFfZ3pzN0w1RlpEX0JnTkhuTlZnNS0yaXo4bEZFMjZsemJVS2pCY0F3OGc2LVhKdVNRTGV2ZUQ5VElRUlZSN0hNOU43VzlNR1p6eWpfOUFyMHQ1S1BaVE9HYTctUkFNTjd0VmVDWTFkaHFBVF9KaFEiLCJwIjoieXhSVEN4cnowVktfR1lIVDUzSExxRHN0U2RCYVplcmpCSEtVbHQtQnVQQmY4ZjUyYUZZMGNZWGlXMGZBLTF3UlBCSzdJdWtycDJWUWlURHlQRjBXeVlxSnN4X3dheTRQd29uczNST1E3TkwyTzJZZEJFUzZLZEZPNEttMDc5MlhBX3NtMEVub0xCSEZUWG52Rkp0YWN3dThZMk1pRUQ0THIyMjZVczBld3Q4IiwicSI6IndfTmQ4RWNxTnc0d0dlX01hR2RVZXNINzRlb2RKY3ZhZE5XS1pZaHg2WHAxWEtoNzRackUyU0lKcGJpU3p3N3pLanVacGN0dGNkTWY3ZzBjWnV5Q1dpZGFTeEtPUy1RMGVkempkMWdnVVdPdUJTcmFIZnB6UXZyOXhPbnk1b3hwTUNkMUJ2bjlsdXk2Q0R6V1FBN3MxbVM4Sk9nY002ZXV6c3pVZDBQaFBHTSIsImRwIjoiUEdzZGtjNUFfLVBvYXdSUE1TcVA5c3MwWENPYTRYdVNjdjVMNnQ4d1R2OWs3REJTdGhQX29rNjgyMzlya04wQlc2Z08tUUg2Tk9GVnBwdGpWa1l6dzE1dVBWYWhScUg3bWx0Q2x6dDlBSmg3SFl6eDBSVkpkYXVLRmhrbmRiMnRja2ZFY20tcW5ZSGotM3J0Z0duXzdQNXUyX3JnWllpd0hVOC1BZWg0NEcwIiwiZHEiOiJFZU9DM08teVEtcHdxNzFfbkx4cU12YklwdnczZ3Y3VVI3eEM4VGYtcGtELXUtSEp4WFBhcXJQM3k0QkpMc3ZfbVFodDQzdnAxdTFlU2Q0NmpJN2s2NVFTSXk1amZUd3RLajduS1RzTFlFTElYVUpuUFR0akVHZFhpWVdPSGt3TlFrOG4yT1l6cDNhZkdTZHNxOVp3LXJXaGs0RDVLaUlSekdGWXVEYWpObDgiLCJxaSI6Inc3aV9aOS1uRUNaTWRBNVRoLVppbjEzdkhqbGo4VWxJQmNMVU12MC10WVpJSDhNMDR3bnB2cTYtbUVwbFY1V3FYNEpTcGRPV05QQzRxSmlyOWVMXzFqYXJidWpBOHlOSVNsQzJDRUZHVEgwbjdnYkJFZFI2RTRaSG1GUjlFdzFPVVBTTTRZQjJtbl9Pd2ViMkR6bjFINkl4TE9lb0FERWpuWlVLRVNOcDBhSSJ9",
    "publicKey": "eyJrdHkiOiJSU0EiLCJraWQiOiJNczQzbDN4dVBCUzhHemxWeU0yajFGSDlHeTM0MFM0bl9sYjNuNE1nNTkwPSIsIm4iOiJtM0dLRVVhX1VSX2tJM3RXaVJUVnFOQ0pueGh4TjNYSm1rS19yV3ZUeE1rMl9taXVRU29qckpwRmlGTjI4ZWJnRWpZYnkyM2VodFlEQi1SVFVaZU1jVUFZb3VaWnBlYlpsOGhWNjBPT0dmSnBxRmVKTmlnVmxUendkN1Y5RWFlTUVhbDVvUDQwbURneW5ocElaNVNVMzhCd1dvd0Vlb2JPSW8xek11UjFwMktOSmxhT0RpR1hPTDVVVzlkT1dOTVF1VUFUVFNhZElZVnZRYmlwMWZMNl96aU9udkIwdWd2TW1qQjRIdXhEMHVsRlo2TFhpdVBqX0ZkN1g4Y2VMVld0R0doTFBKb09vUFp3LTluTzVwcDRMb1NjZFFXNGxCVjY3MVJQdThCb3JYYUREampqVVVpN2NGemR1eXFnNkhHZWk3UlFpY25POHhhTDhRekxLNUNnUFEiLCJlIjoiQVFBQiJ9"
  }
}
```

* `product` path compoent is extracted from `selector.product`
* `version` path component is extracted from `selector.version`

#### Application secrets

> Protobuf definition - [ApplicationComponentNS](https://github.com/elastic/harp/blob/main/api/proto/harp/bundle/v1/template.proto#L151)

An `application` is an instance of a `product` running an a `platform` at a
`quality` stage level.

> Application secrets are shared by all instances of the same software. Secrets
> must not be identic for different platform / quality.

```yaml
spec:
  selector:
    quality: "production"
    platform: "customer1"
    product: "ece"
    version: "v1.0.0"
  namespaces:
    application:
    # AdminConsole -------------------------------------------------------------
    - name: "adminconsole"
      description: "Administration UI console"
      secrets:
      - suffix: "authentication/otp/okta_api_key"
        description: "Okta API Key for OTP validation"
        # Not generated by workflow, requested at generation time.
        vendor: true
        template: |-
          {
            "API_KEY": "{{ .Values.vendor.okta.api_key }}"
          }
      - suffix: "database/usage_credentials"
        description: "PostgreSQL database account for component usage"
        template: |-
          {
            "host": "sample-instance.abc2defghije.us-west-2.rds.amazonaws.com",
            "port": "5432",
            "options": "sslmode=require&application_name={{ .Data.Component }}",
            "username": "dbuser-{{ .Data.Component }}-{{ randAlphaNum 8 }}",
            "password": "{{ paranoidPassword | b64enc }}",
            "dbname": "{{ .Data.Component }}"
          }
```

This will generate 2 secrets like the following ones :

```json
{
  "app/production/customer1/ece/v1.0.0/adminconsole/authentication/otp/okta_api_key": {
    "API_KEY": "okta-foo-api-123456789"
  },
  "app/production/customer1/ece/v1.0.0/adminconsole/database/usage_credentials": {
    "dbname": "adminconsole",
    "host": "sample-instance.abc2defghije.us-west-2.rds.amazonaws.com",
    "options": "sslmode=require&application_name=adminconsole",
    "password": "dHQiMXhQI3VxbDRHaUxAUWRiSzRwe1RNOEN2YnhrMnY6VlF6VXpLUlF4byoyXnIyV300bm5jOFBnXHFvdl9hNw==",
    "port": "5432",
    "username": "dbuser-adminconsole-qc7GUKPN"
  }
}
```

* `quality` path component is extract from `selector.quality`
* `platform` path component is extract from `selector.platform`
* `product` path component is extracted from `selector.product`
* `version` path component is extracted from `selector.version`

## Usages

A `BundleTemplate` uses the `harp` template engine to render a `Bundle`
object stored in a secret container.

> Extracted from [Customer BundleTemplate](../../customer-bundle)

```sh
$ harp from template --in essp.yaml \
  --values=values.yaml \
  --set quality=production | harp bundle dump --path-only
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

---

* [Previous topic](2-bundle.md)
* [Index](../)
* [Next topic](4-patch.md)
