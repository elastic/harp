# Use cases

## Generate environment shell script from Vault secret

Given this template :

```ruby
#!/bin/sh
# Admin Console
{{- with secret "app/production/operations/cst/v1.0.0/privaas/external/userconsole" }}
export PRVS_CLIENT_ADMINCONSOLE_ORGURL="{{ .organization }}"
export PRVS_CLIENT_ADMINCONSOLE_USER="{{ .user }}"
export PRVS_CLIENT_ADMINCONSOLE_PASSWORD="{{ .password }}"
{{- end }}
```

When invoking `harp` :

```sh
export VAULT_ADDR=<vault-url>
export VAULT_TOKEN="$(vault login -token-only -method='oidc')"
harp template --in privaas.env.sh.tmpl --out privaas.env.sh
```

It will pull secrets from Vault and render `privaas.env.sh`.

You can also use a secret container as source of secrets by specifying the
`--secrets-from` flag.

```sh
harp template \
  --in privaas.env.sh.tmpl \
  --out privaas.env.sh \
  --secrets-from privaas.container \
  --secrets-from vault # Fallback to vault if secret value is not found.
```

## Generate AWS profile settings based on TF variable

Given this template :

```ruby
{{- $role := "sre" }}{{ if hasKey .Values "role" }}{{ $role = default "sre" .Values.role }}{{ end -}}
{{- range $key, $account := .Values.accounts.variable.account_id.default -}}
{{- $awsNamespace := "aws" }}{{ if hasPrefix "gov" $key}}{{ $awsNamespace = "aws-gov" }}{{ end -}}
{{- $awsProfile := "commercial" }}{{ if hasPrefix "gov-" $key }}{{ if hasSuffix "-public" $key }}{{ $awsProfile = "gov-public" }}{{ else }}{{ $awsProfile = "gov" }}{{ end }}{{ end -}}
[profile {{ $key }}]
role_arn       = arn:{{ $awsNamespace }}:iam::{{ $account }}:role/xaccount/{{ $role }}
source_profile = {{ $awsProfile }}
{{ end -}}
```

When invoking `harp` :

```sh
harp template
    --in config.tpl \
    --values path-to/accounts.tf:hcl2:accounts \
    --set role=seceng
```

It will generate the following output :

```ini
[profile com-ops]
role_arn       = arn:aws:iam::12345678911234:role/xaccount/seceng
source_profile = commercial
[profile gov-main]
role_arn       = arn:aws-gov:iam::31256489123112:role/xaccount/seceng
source_profile = gov-private
[profile gov-main-public]
role_arn       = arn:aws-gov:iam::00645684152311:role/xaccount/seceng
source_profile = gov-public
...
```

---

* [Previous topic](8-whitespace-controls.md)
* [Index](../)
* [Next topic](../2-secret-container/1-introduction.md)
