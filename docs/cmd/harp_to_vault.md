## harp to vault

Push a secret container in Hashicorp Vault

```
harp to vault [flags]
```

### Options

```
  -h, --help                  help for vault
      --in string             Container path ('-' for stdin or filename) (default "-")
      --namespace string      Vault namespace
      --prefix string         Vault backend prefix
      --with-metadata         Push container metadata as secret data
      --with-vault-metadata   Push container metadata as secret metadata (requires Vault >=1.9)
      --worker-count int      Active worker count limit (default 4)
```

### SEE ALSO

* [harp to](harp_to.md)	 - Secret container conversion commands

