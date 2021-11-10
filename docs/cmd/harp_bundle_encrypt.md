## harp bundle encrypt

Encrypt secret values

```
harp bundle encrypt [flags]
```

### Options

```
  -h, --help                        help for encrypt
      --in string                   Container input ('-' for stdin or filename)
      --key string                  Secret value encryption key for full bundle encryption
      --key-alias strings           Secret value encryption key for partial bundle encryption ('alias:key')
      --out string                  Container output ('-' for stdout or filename)
  -s, --skip-unresolved-key-alias   Skip unresolved key alias during partial bundle encryption
```

### SEE ALSO

* [harp bundle](harp_bundle.md)	 - Bundle commands

