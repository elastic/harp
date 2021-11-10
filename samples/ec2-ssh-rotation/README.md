# AWS EC2 SSH Rotation

Generate and provision 2 SSH key pairs (RSA / Ed25519) in Vault, and then deploy
to AWS account using Terraform.

## Scenario

### Secret requirements

Describes your secret requirements using the `BundleTemplate` specification and
secret micro-template functions.

```yaml
apiVersion: harp.elastic.co/v1
kind: BundleTemplate
meta:
  name: "Ec2ssh"
  owner: cloud-security@elastic.co
  description: "EC2 SSH Key"
spec:
  namespaces:
    infrastructure:
    - provider: "aws"
      account: "{{ .Values.infra.aws.account }}"
      description: "AWS Account"
      regions:
      - name: "global"
        services:
        - type: "ec2"
          name: "default"
          description: "Authentication for all EC2 instances"
          secrets:
          - suffix: "ssh/rsa_keys"
            description: "Private SSH keys for connection"
            template: |-
              {
                "private": {{ $sshKey := cryptoPair "rsa" }}{{ $sshKey.Private | toSSH | toJson }},
                "public": "{{ $sshKey.Public | toSSH | trim }} cloud-security@elastic.co"
              }
          - suffix: "ssh/ed25519_keys"
            description: "Private SSH keys for connection"
            template: |-
              {
                "private": {{ $sshKey := cryptoPair "ssh" }}{{ $sshKey.Private | toSSH | toJson }},
                "public": "{{ $sshKey.Public | toSSH | trim }} cloud-security@elastic.co"
              }
```

In this example, we specify :

* The secret owner that will be used in case of incident related to this secret values
* A bundle description to describe secret bundle usage
* That `infrastructure account` path component is an external value passed during generation
* To create 2 secret paths :
  * `infra/aws/{{ infra.aws.account }}/global/ec2/default/ssh/rsa_keys` to hold the RSA key pair
  * `infra/aws/{{ infra.aws.account }}/global/ec2/default/ssh/ed25519_keys` to hold the Ed25519 key pair
* To format RSA key secret value as a JSON object
  * `private` attribute to hold the RSA private key encoded for SSH usages and transform to JSON string
  * `public` attribute to hold the RSA public key encoded for SSH usages and trim white spaces
* To format Ed25519 key secret value as a JSON object
  * `private` attribute to hold the Ed25519 private key encoded for SSH usages and transform to JSON string
  * `public` attribute to hold the Ed25519 public key encoded for SSH usages and trim white spaces

### Generate the bundle

Given this value file

```yaml
infra:
  aws:
    account: security
```

In order to generate the bundle :

```sh
harp from template --in spec.yaml --values values.yaml --out ec2_ssh.bundle
```

Or by passing directly the variable value :

```sh
harp from template --in spec.yaml --set infra.aws.account=security --out ec2_ssh.bundle
```

### Publish the bundle

#### List secret path

```sh
$ harp bundle dump --in ec2_ssh.bundle| jq -r ".packages[].name"
infra/aws/security/global/ec2/default/ssh/rsa_keys
infra/aws/security/global/ec2/default/ssh/ed25519_keys
```

#### Validate generated secret values

```sh
harp bundle read --in ec2_ssh.bundle --path infra/aws/security/global/ec2/default/ssh/rsa_keys | jq
```

```json
{
  "private": "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEAoZMWQ2ahTD5bYw5ojIxO91mHzcdgVa2PspwvpHmVdkCnAhSV\ndfurx4vOQdJ2fkf9iftZ90xnnEn3nWfgLkJYXrz/M0qgPCjP9HzFlncHKvvwfBfy\ne4lyDQ190jL1jCokHMuqpGfXac7fOI80d2zSYSf0rCJBzIq6mHPzm1wLZKXwRc6e\nsv3SdQKpieEL0AS7wmRYrFD8+u5KVymc0lK9M0pPvCBS9U1suud/51Pi5n5jjUgo\n9b57Y1JxKNxIfaRuxpW8YPmHOfjB2yRL13G+ztr7lq4XeKMDRej2nwD9y3VpcHuO\nWwVYlWLpJh5A5CQvGuIUql3RhJhXr0YMuVQIwwIDAQABAoIBAQCe4eGJo9lG/Sam\ncJu0YaChMMQPQVhkyAg/LcDXrsuffhH8RLt4SmPwFHVdWpU0tpwF3EnqmZJlKIM6\noNPdCVaWyqj1ChQpNETR+QPfYuqEhTxE+tNyUYT6euLYGU5GZ4YdXtVNV+OG42uz\n1XZVXYg/C6hYwFMNzPmSQqsxgcCeVdLCX1NDXAkOIvUnneGSLxrCLD1g3SMjcHqd\nxp2B3OSJ6mz8YRgQqEoVh/KLDMAuT76+W3ybfL6TCpF1XpE6R/L95js6nh8U6a9k\nNhfuPS9O6RJ4CAwKmZrbAhkPg8MwjrI3iP2HphNIRRIfCXxRjk7fSFnv6SKa256l\nRtxxrNRZAoGBAModYQ7hYS5hjSc/sxRDRgVJSwY+CYBgq3+SSUu7NulSE02LJILw\nr9agYWr0cwCGo+p7z+XqseEdQgz+bxImFUp54EStOUXamDP+/TpWTYxytaBXHuyW\nSjXncoNvS4mtNDfbwO2YFOA+HwnkuG1hoILwCVXsh0DyI4cuyFJlMdh3AoGBAMym\nx/Vt9qtWOw7l1ky/xoPHyKgbTViSIIKdHaMjS1wLtSLPL/PHQ8WZHIgMZ19LKV05\ntYUkxuUjVDe5FinEHZonHgo7XuCtQKutVAf7OV0E2fODXZ+raAO+lSHglGeRcz0V\nhUtniFCHQARvvr0S2zC3I2ezowRf+3vsI/4uxbEVAoGAT+WLP6milAYmGXTZ4tGx\nNVfC26Xcda5BPT3j0ZchXtx+GvI9LbHdoNdnizM03ulalM/64CWFybqaKK7P03nQ\ngs9o810SveVqs2tCSTRilXmnCMpHxDio/2QN5Z0yXCtm8Anj20h6QCbueCe9LCgi\nnoArAJdu5CKUHtVEhSXRrYMCgYEAq4d2vb0nLMCfy4LUtYtfxgBjrJMFpyEDYrZx\nqtTgSwv9DGn/1SHFKg+FHHrZAcQrrVm7TRdgJZoQ8ouNigA4l4YF5amRglt0gvBK\nKE5m7BIu463NgRDXo5vPv49Ok+gTYLVy/ZqPZH+YJp/KjQsK8K/vWvHzxqz0Sg/I\nszlctWkCgYAf/aoJzHPFPRCZqG8oxDXglWz8OMKd3eaC+WLD8XyMXFbhT+4S05yp\nfOy1Z7+PIxMYHgIMl/ftrTPROWqXQ/uzpF5RSAXB4nB7LFnDjDEBZ/HCi+DFf00A\nqbDIC5lQwSGmGcELUqpK/9aqBUrI5i2jFSW/RGhHpyTXtfGTh9Q6+w==\n-----END RSA PRIVATE KEY-----\n",
  "public": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQChkxZDZqFMPltjDmiMjE73WYfNx2BVrY+ynC+keZV2QKcCFJV1+6vHi85B0nZ+R/2J+1n3TGecSfedZ+AuQlhevP8zSqA8KM/0fMWWdwcq+/B8F/J7iXINDX3SMvWMKiQcy6qkZ9dpzt84jzR3bNJhJ/SsIkHMirqYc/ObXAtkpfBFzp6y/dJ1AqmJ4QvQBLvCZFisUPz67kpXKZzSUr0zSk+8IFL1TWy653/nU+LmfmONSCj1vntjUnEo3Eh9pG7Glbxg+Yc5+MHbJEvXcb7O2vuWrhd4owNF6PafAP3LdWlwe45bBViVYukmHkDkJC8a4hSqXdGEmFevRgy5VAjD cloud-security@elastic.co"
}
```

> With direct field

```sh
$ harp b read --in production.bundle --path infra/aws/security/global/ec2/default/ssh/rsa_keys --field private
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAoZMWQ2ahTD5bYw5ojIxO91mHzcdgVa2PspwvpHmVdkCnAhSV
dfurx4vOQdJ2fkf9iftZ90xnnEn3nWfgLkJYXrz/M0qgPCjP9HzFlncHKvvwfBfy
e4lyDQ190jL1jCokHMuqpGfXac7fOI80d2zSYSf0rCJBzIq6mHPzm1wLZKXwRc6e
sv3SdQKpieEL0AS7wmRYrFD8+u5KVymc0lK9M0pPvCBS9U1suud/51Pi5n5jjUgo
9b57Y1JxKNxIfaRuxpW8YPmHOfjB2yRL13G+ztr7lq4XeKMDRej2nwD9y3VpcHuO
WwVYlWLpJh5A5CQvGuIUql3RhJhXr0YMuVQIwwIDAQABAoIBAQCe4eGJo9lG/Sam
cJu0YaChMMQPQVhkyAg/LcDXrsuffhH8RLt4SmPwFHVdWpU0tpwF3EnqmZJlKIM6
oNPdCVaWyqj1ChQpNETR+QPfYuqEhTxE+tNyUYT6euLYGU5GZ4YdXtVNV+OG42uz
1XZVXYg/C6hYwFMNzPmSQqsxgcCeVdLCX1NDXAkOIvUnneGSLxrCLD1g3SMjcHqd
xp2B3OSJ6mz8YRgQqEoVh/KLDMAuT76+W3ybfL6TCpF1XpE6R/L95js6nh8U6a9k
NhfuPS9O6RJ4CAwKmZrbAhkPg8MwjrI3iP2HphNIRRIfCXxRjk7fSFnv6SKa256l
RtxxrNRZAoGBAModYQ7hYS5hjSc/sxRDRgVJSwY+CYBgq3+SSUu7NulSE02LJILw
r9agYWr0cwCGo+p7z+XqseEdQgz+bxImFUp54EStOUXamDP+/TpWTYxytaBXHuyW
SjXncoNvS4mtNDfbwO2YFOA+HwnkuG1hoILwCVXsh0DyI4cuyFJlMdh3AoGBAMym
x/Vt9qtWOw7l1ky/xoPHyKgbTViSIIKdHaMjS1wLtSLPL/PHQ8WZHIgMZ19LKV05
tYUkxuUjVDe5FinEHZonHgo7XuCtQKutVAf7OV0E2fODXZ+raAO+lSHglGeRcz0V
hUtniFCHQARvvr0S2zC3I2ezowRf+3vsI/4uxbEVAoGAT+WLP6milAYmGXTZ4tGx
NVfC26Xcda5BPT3j0ZchXtx+GvI9LbHdoNdnizM03ulalM/64CWFybqaKK7P03nQ
gs9o810SveVqs2tCSTRilXmnCMpHxDio/2QN5Z0yXCtm8Anj20h6QCbueCe9LCgi
noArAJdu5CKUHtVEhSXRrYMCgYEAq4d2vb0nLMCfy4LUtYtfxgBjrJMFpyEDYrZx
qtTgSwv9DGn/1SHFKg+FHHrZAcQrrVm7TRdgJZoQ8ouNigA4l4YF5amRglt0gvBK
KE5m7BIu463NgRDXo5vPv49Ok+gTYLVy/ZqPZH+YJp/KjQsK8K/vWvHzxqz0Sg/I
szlctWkCgYAf/aoJzHPFPRCZqG8oxDXglWz8OMKd3eaC+WLD8XyMXFbhT+4S05yp
fOy1Z7+PIxMYHgIMl/ftrTPROWqXQ/uzpF5RSAXB4nB7LFnDjDEBZ/HCi+DFf00A
qbDIC5lQwSGmGcELUqpK/9aqBUrI5i2jFSW/RGhHpyTXtfGTh9Q6+w==
-----END RSA PRIVATE KEY-----

```

> Or via dump command

```sh
harp b dump --in ec2_ssh.bundle --content-only | jq -r '.["infra/aws/security/global/ec2/default/ssh/ed25519_keys"]' | jq
```

```json
{
  "private": "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtz\nc2gtZWQyNTUxOQAAACC8qE+2GoBkj01x6N2DL+rRZD2WXGB5HDtD0ttn8ThvkAAA\nAIh8q3I6fKtyOgAAAAtzc2gtZWQyNTUxOQAAACC8qE+2GoBkj01x6N2DL+rRZD2W\nXGB5HDtD0ttn8ThvkAAAAEDQqC+n4eADwDU0ZayMIgy+N0iRLLfdEalJEcQMxDL5\n8byoT7YagGSPTXHo3YMv6tFkPZZcYHkcO0PS22fxOG+QAAAAAAECAwQF\n-----END OPENSSH PRIVATE KEY-----\n",
  "public": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILyoT7YagGSPTXHo3YMv6tFkPZZcYHkcO0PS22fxOG+Q cloud-security@elastic.co"
}
```

#### Export to Vault

```sh
VAULT_ADDR=https://.......
VAULT_TOKEN=$(vault login -method=oidc -token-only)
```

You can batch import in Vault the complete bundle

```sh
harp to vault --in ec2_ssh.bundle --prefix security
```

> `prefix` will add the given prefix to each path.

Or unitary patch secret values

```sh
harp bundle read \
    --in ec2_ssh.bundle \
    --path infra/aws/security/global/ec2/default/ssh/rsa_keys \
    | vault kv put infra/aws/security/global/ec2/default/ssh/rsa_keys -
```

### Deploy AWS changes

Once deployed to Vault, just pull secret from Vault and deploy the new keys.

AWS Key deployment TF script :

```ruby
data "vault_generic_secret" "ssh_rsa_keypair" {
  path = "infra/aws/security/global/ec2/default/ssh/rsa_keys"
}

resource "aws_key_pair" "operation" {
  key_name   = "secops-key"
  public_key = data.vault_generic_secret.ssh_rsa_keypair.data["public"]
}
```

This script used Vault as source of truth, pull the secret value and deploy an
SSH key pair in AWS. It must be executed via a pipeline or a Rundeck job to
automate key deployment. So that in case of incident, just regenerate the SSH key
using the bundle specification, push the secret to Vault and execute secret
deployment pipeline.

In order to retrieve the private key :

```sh
vault kv read -field=private infra/aws/security/global/ec2/default/ssh/rsa_keys > secops.pem
ssh -i secops.pem user@host
```
