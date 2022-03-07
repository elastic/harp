## harp crate push

Push a crate

### Synopsis

Export a crate to an OCI compatible registry.

```
harp crate push [flags]
```

### Options

```
  -c, --config stringArray   Authentication config path
  -f, --cratefile string     Specification path ('-' for stdin or filename) (default "Cratefile")
  -h, --help                 help for push
      --insecure             Allow connections to SSL registry without certs
      --json                 Enable JSON output
      --out string           Output path ('-' for stdout or filename) (default "-")
  -p, --password string      Registry password
      --plain-http           Use plain http and not https
      --ref string           Container path (default "harp.sealed")
      --root string          Defines the root context path (default ".")
      --to string            Target destination (registry, oci:<path>, files:<path>) (default "registry")
  -u, --username string      Registry username
```

### SEE ALSO

* [harp crate](harp_crate.md)	 - Crate management commands

