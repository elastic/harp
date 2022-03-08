## harp crate copy

Copy a crate

### Synopsis

Copy a crate from one source to another.

```
harp crate copy [flags]
```

### Examples

```
  # Copy a crate from a registry to file for debugging purpose
  harp crate copy --from-ref <registry>/<image>:<tag> --to files:out
```

### Options

```
      --from string            Target destination (registry, oci:<path>, files:<path>) (default "registry")
      --from-insecure          Allow connections to SSL registry without certs
  -p, --from-password string   Registry password
      --from-plain-http        Use plain http and not https
      --from-ref string        Source image reference
  -u, --from-username string   Registry username
  -h, --help                   help for copy
      --to string              Target destination (registry, oci:<path>, files:<path>) (default "file")
      --to-insecure            Allow connections to SSL registry without certs
      --to-password string     Registry password
      --to-plain-http          Use plain http and not https
      --to-ref string          Source image reference
      --to-username string     Registry username
```

### SEE ALSO

* [harp crate](harp_crate.md)	 - Crate management commands

