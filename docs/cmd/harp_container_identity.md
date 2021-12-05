## harp container identity

Generate container identity

```
harp container identity [flags]
```

### Options

```
      --description string          Identity description
  -h, --help                        help for identity
      --key string                  Transformer key
      --out string                  Identity information output ('-' for stdout or filename)
      --passphrase string           Identity private key passphrase
      --vault-transit-key string    Use Vault transit encryption to protect identity private key
      --vault-transit-path string   Vault transit backend mount path (default "transit")
      --version uint                Select identity version (0:legacy, 1:modern, 2:nist) (default 1)
```

### SEE ALSO

* [harp container](harp_container.md)	 - Secret container commands

