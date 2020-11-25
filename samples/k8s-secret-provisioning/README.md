# Kubernetes secret provisioning

## Direct rendering

Generate secret resource from YAML definition.

```yaml
{{- $suffix := randAlpha 8 -}}
apiVersion: v1
kind: Secret
metadata:
  name: my-secret
type: Opaque
data:
  username: {{ printf "user-%s" $suffix | b64enc }}
  password: {{ strongPassword | b64enc }}
```

Render and deploy :

```sh
harp template --in bundle.yaml | kubectl create -f -
```

## From secrets

Conditionally pull secrets using secret loaders :

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: my-secret
  namespace: {{ .Values.platform }}
type: Opaque
data:
  {{- with secret (printf "app/%s/%s/cst/v1.0.0/privaas/database/credentials .Value.quality .Values.platform) -}}
  {{- range $k, $v := . -}}
  {{ $k }}: {{ $v | b64enc }}
  {{- end -}}
```

Render and deploy :

```sh
harp template --in bundle.yaml \
  --set quality=production \
  --set platform=security \
  | kubectl create -f -
```

## From an existing bundle

Prepare the template :

```yaml
{{- $values := .Values -}}
{{- range $path, $secret := .Values.Secrets -}}
---
apiVersion: v1
kind: Secret
metadata:
  name: "{{ $path }}"
  namespace: {{ $values.namespace }}
type: Opaque
data: {{ range $k, $v := $secret }}
  {{ $k }}: {{ $v | b64enc }}{{ end }}
{{ end }}

```

Extract secret as JSON for injecting them as values :

```sh
harp bundle dump --content-only --out secrets.json
```

Render the manifest using `secrets.json` as Values :

```sh
$ harp template --in multiple.yaml \
  --values -:json:Secrets \ # Read from stdin, decode as json and assign the object to .Values.Secrets
  --set namespace=security \ # Define k8s namespace
---
apiVersion: v1
kind: Secret
metadata:
  name: "app/production/customer1/ece/v1.0.0/adminconsole/authentication/otp/okta-api-key"
  namespace: security
type: Opaque
data:
  API_KEY: b2t0YS1mb28tYXBpLTEyMzQ1Njc4OQ==
---
apiVersion: v1
kind: Secret
metadata:
  name: "app/production/customer1/ece/v1.0.0/adminconsole/database/usage-credentials"
  namespace: security
type: Opaque
...
```

One liner from `BundleTemplate` to multiple secrets in a given namespace.

```sh
harp from template --in ../customer-bundle/spec.yaml \
  --values ../customer-bundle/values.yaml \
  --set quality=production
  | harp bundle filter --keep app/production \ # Keep only "app/production" secrets
  | harp bundle dump --content-only \ # Dump as json
  | harp template --in multiple.yaml \ # Render secret manifest
    --values -:json:Secrets \
    --set namespace=security \
  | kubectl create -f
```
