## harp bundle decrypt

Decrypt secret values

### Synopsis

Decrypt a bundle content.

For confidentiality purpose, bundle package value can be encrypted before
the container sealing. It offers confidentiality properties so that the
final consumer must know an additional decryption key to be allowed to
read the package value.

All package properties (name, labels, annotations) remain a clear-text
message. Only package values (secret K/V) is encrypted.

In order to decrypt the package value, harp uses the value encryption
transformers. The required key must be provided in a format understandable
by the encryption transformer factory.

This act as in-transit/in-use encryption.

```
harp bundle decrypt [flags]
```

### Examples

```
  # Decrypt a bundle from STDIN and produce output to STDOUT
  harp bundle decrypt --key <transformer key>
  
  # Decrypt a bundle from STDIN using multiple transformer keys
  harp bundle decrypt --key <transformer key 1> --key <transformer key 2>
  
  # Decrypt a bundle from STDIN and ignore secrets which could not be decrypted
  # with given transformer key (partial decryption / authorization by key)
  harp bundle decrypt --skip-not-decryptable --key <transformer-key>
  
  # Decrypt a bundle from STDIN and produce output to a file
  harp bundle decrypt --key <transformer key> --out decrypted.bundle
```

### Options

```
  -h, --help                   help for decrypt
      --in string              Container input ('-' for stdin or filename)
      --key strings            Secret value decryption key. Repeat to add multiple keys to try.
      --out string             Container output ('-' for stdout or filename)
  -s, --skip-not-decryptable   Skip not decryptable secrets without raising an error.
```

### SEE ALSO

* [harp bundle](harp_bundle.md)	 - Bundle commands

