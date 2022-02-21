## harp transform decode

Decode given input

### Synopsis

Decode the given input stream using the selected decoding strategy.

Supported codecs:
  * identity - returns the unmodified input
  * hex/base16 - returns the hexadecimal encoded input
  * base32 - returns the Base32 encoded input
  * base32hex - returns the Base32 with extended alphabet encoded input
  * base64 - returns the Base64 encoded input
  * base64raw - returns the Base64 encoded input without "=" padding
  * base64url - returns the Base64 encoded input using URL safe characters
  * base64urlraw - returns the Base64 encoded input using URL safe characters without "=" padding
  * base85 - returns the Base85 encoded input

```
harp transform decode [flags]
```

### Examples

```
  # Decode base64 from stdin
  echo "dGVzdAo=" | harp transform decode --encoding base64
  
  # Decode base64url from a file
  harp transform decode --in test.txt --encoding base64url
```

### Options

```
      --encoding string   Encoding strategy (default "identity")
  -h, --help              help for decode
      --in string         Input path ('-' for stdin or filename) (default "-")
      --out string        Output path ('-' for stdout or filename) (default "-")
```

### SEE ALSO

* [harp transform](harp_transform.md)	 - Transform input value using encryption transformers

