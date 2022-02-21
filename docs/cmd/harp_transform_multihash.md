## harp transform multihash

Multiple hash  from given input

### Synopsis

Process the input to compute the hashes according to selected hash algorithms.

The command input is limited to size lower than 250 MB.

Supported Algorithms:
  blake2b-256, blake2b-384, blake2b-512, md4, md5, ripemd160, sha1, sha224, sha256, sha3-224, sha3-256, sha3-384, sha3-512, sha512, sha512/224, sha512/256

```
harp transform multihash [flags]
```

### Examples

```
  # Compute md5, sha1, sha256, sha512 in one read from a file
  harp transform multihash --in livecd.iso
  
  # Compute sha256, sha512 only
  harp transform multihash --algorithm sha256 --algorithm sha512 --in livecd.iso
  
  # Compute sha256, sha512 only with JSON output
  harp transform multihash --json --algorithm sha256 --algorithm sha512 --in livecd.iso
```

### Options

```
      --algorithm strings   Hash algorithms to use (default [md5,sha1,sha256,sha512])
  -h, --help                help for multihash
      --in string           Input path ('-' for stdin or filename) (default "-")
      --json                Display multihash result as json
      --out string          Output path ('-' for stdout or filename) (default "-")
```

### SEE ALSO

* [harp transform](harp_transform.md)	 - Transform input value using encryption transformers

