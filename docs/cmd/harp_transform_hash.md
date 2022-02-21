## harp transform hash

Hash given input

### Synopsis

Process the input to compute the hash according to selected hash algorithm.

The command input is limited to size lower than 250 MB.

Supported Algorithms:
  blake2b-256, blake2b-384, blake2b-512, md5, sha1, sha224, sha256, sha3-224, sha3-256, sha3-384, sha3-512, sha512, sha512/224, sha512/256

```
harp transform hash [flags]
```

### Examples

```
  # Compute SHA256 from stdin
  echo -n 'test' | harp transform hash
  
  # Compute SHA512 hash from a file
  harp transform hash --algorithm sha512
  
  # Compute Blake2b hash from a file with base64 encoded output
  harp transform hash --algorithm blake2b-512 --encoding base64
  
  # Check the given input integrity (default sha256 / hex)
  harp transform hash --in livecd.iso --validate 4506369c20d2a95ebad9234b7f48e0eded4ec4ee1de0cb45a195b1e38fde27f7
  
  # Check the given input integrity with specific hash algorihm and encoding
  harp transform hash --in livecd.iso --algorithm BLAKE2b_512 --encoding base64urlraw --validate dquOtQ-gj815njSbk8mGl3WUgImkflX1AaLXy6ymhk_kUpP6qXDmSC5X2l3nkTgJK9F6p3rBV6o075QZQ-HHaw
```

### Options

```
      --algorithm string   Hash algorithm to use (default "SHA256")
      --encoding string    Encoding strategy (hex, base64, base64raw, base64url, base64urlraw) (default "hex")
  -h, --help               help for hash
      --in string          Input path ('-' for stdin or filename) (default "-")
      --out string         Output path ('-' for stdout or filename) (default "-")
      --validate string    Expecetd hash to validate the output with. Decoded using the given encoding strategy.
```

### SEE ALSO

* [harp transform](harp_transform.md)	 - Transform input value using encryption transformers

