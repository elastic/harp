## harp from vault

Pull a list of Vault K/V paths as a secret container

```
harp from vault [flags]
```

### Options

```
  -h, --help                help for vault
      --namespace string    Vault namespace
      --out string          Container output ('-' for stdout or filename)
      --path stringArray    Vault backend path (and recursive)
      --paths-from string   Path to read path from ('-' for stdin or filename)
      --with-metadata       Pull bundle metadata from Vault (default true)
      --worker-count int    Active worker count limit (default 4)
```

### SEE ALSO

* [harp from](harp_from.md)	 - Secret container generation commands

