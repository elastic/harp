## harp transform hash

Hash given input

### Synopsis

Process the input to compute the hash according to selected hash algoritm.

The command input is limited to size lower than 250 MB.

Supported Algorithms:
  SHA-1,SHA-512/256,BLAKE2b-256,SHA3-224,MD5,SHA3-256,SHA3-512,SHA-512,BLAKE2b-512,RIPEMD-160,MD4,SHA-256,SHA-512/224,BLAKE2b-384,SHA3-384,SHA-224

```
harp transform hash [flags]
```

### Examples

```
  # Compute SHA256 from stdin
  echo -n 'test' | harp transform hash
  
  # Compute CRC32 hash from a file
  harp transform hash --algorithm CRC32
  
  # Compute Blake2b hash from a file with base64 encoded output
  harp transform hash --algorithm BLAKE2b_512 --encoding base64
  
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

