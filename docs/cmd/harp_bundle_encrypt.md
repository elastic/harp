## harp bundle encrypt

Encrypt secret values

### Synopsis

Apply package content encryption.

For confidentiality purpose, bundle package value can be encrypted before
the container sealing. It offers confidentiality properties so that the
final consumer must know an additional decryption key to be allowed to
read the package value even if it can unseal the container.

All package properties (name, labels, annotations) remain a clear-text
message. Only package values (secret K/V) are encrypted.

This act as in-transit/in-use encryption.

Annotations:

* harp.elastic.co/v1/package#encryptionKeyAlias=<alias> - Set this
  annotation on packages to reference a key alias.

```
harp bundle encrypt [flags]
```

### Examples

```
  # Encrypt a whole bundle from STDIN and produce output to STDOUT
  harp bundle encrypt --key <transformer key>
  
  # Encrypt partially a bundle using the annotation matcher from STDIN and
  # produce output to STDOUT
  harp bundle encrypt --key-alias <alias>:<transformer key> --key-alias <alias-2>:<transformer key 2>
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

