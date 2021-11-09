# Files

You can inject a complete file system during template rendering, this can be
helpful when handling numerous files without explicit reference and inclusion as
it could be done using `.Values`.

The file loader root path must be specified while using template engine `--root`,
it will load all files recursively from the root in memory.

> Pay attention to large root file system when using `--root` flag.

All files are loaded and represent a map accessible via `.Files` handle where
key is the relative path to root of the file and the value the `bytes` content
of the file crawled.

You have to use the `.Files` handle to access file management functions :

## .Files.GetBytes(string)

Fetch the contents of a file as a byte array.

```yaml
{{.Files.GetBytes "foo"}}
```

Equivalent of :

```yaml
{{index .Files $path}}
```

## .Files.Get(string)

Fetch the contents of a file as a string. It is designed to be called in a
template.

```yaml
{{.Files.Get "foo"}}
```

## .Files.Glob(string)

Glob takes a glob pattern and returns another files object only containing
matched files.

```yaml
{{ range $name, $content := .Files.Glob("foo/**") }}
{{ $name }}: |
{{ .Files.Get($name) | indent 4 }}{{ end }}
```

## .Files.AsConfig()

`AsConfig` returns a Files group and flattens it to a YAML map suitable for
including in the 'data' section of a Kubernetes ConfigMap definition.
Duplicate keys will be overwritten, so be aware that your file names
(regardless of path) should be unique.

The output will not be indented, so you will want to pipe this to the
'indent' template function.

```yaml
data:
{{ (.Files.Glob "certs/**").AsConfig | toYaml | indent 4 }}
```

```sh
$ harp template --root=. --in certs.yaml
  intermediate_ca.crt: |
    -----BEGIN CERTIFICATE-----
    MIIBoTCCAUegAwIBAgIRAPdzPAjBHfxh0kKwDelUAjcwCgYIKoZIzj0EAwIwGjEY
    ...
    ATWYAiAlNYS/2X7G5IXRYHDGKyyW7tIXhzWegHVAXR2gbwdd/Q==
    -----END CERTIFICATE-----
  root_ca.crt: |
    -----BEGIN CERTIFICATE-----
    MIIBdjCCAR2gAwIBAgIQLGDbCrk+QDh5W7bdfFgsZDAKBggqhkjOPQQDAjAaMRgw
    ...
    qJiLIgwbayaxAh9ZDboH9Uq+gmqHYgqQ23Cu34ea7pNN7SG0G2iWQyDq
    -----END CERTIFICATE-----
```

## .Files.AsSecret()

`AsSecrets` returns the base64-encoded value of a Files object suitable for
including in the 'data' section of a Kubernetes Secret definition.
Duplicate keys will be overwritten, so be aware that your file names
(regardless of path) should be unique.

The output will not be indented, so you will want to pipe this to the
'indent' template function.

```yaml
data:
  {{ (.Files.Glob "certs/*").AsSecrets | toYaml | indent 2 }}
```

When executing `harp` :

```sh
$ harp template --root . --in certs.yaml
data:
  intermediate_ca.crt: LS0tLS1CRUdJTiBDRVJUSU.....JRklDQVRFLS0tLS0K
  root_ca.crt: LS0tLS1CRUdJTiBDRVJ.....VElGSUNBVEUtLS0tLQo=
```

## .Files.Lines()

`Lines` return each line of a named file (split by "\n") as a slice, so it can
be ranged over in your templates.

```yaml
{{ range .Files.Lines "foo/bar.html" }}
{{ . }}{{ end }}
```

---

* [Previous topic](4-values.md)
* [Index](../)
* [Next topic](6-lists-and-maps.md)
