## harp transform encode

Encode given input

### Synopsis

Encode the given input stream using the selected encoding strategy.

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
harp transform encode [flags]
```

### Examples

```
  # Encode stdin to base64
  echo "test" | harp transform encode --encoding base64
  
  # Encode a file using base64url
  harp transform encode --encoding base64url --in test.txt --out encoded.text
```

### Options

```
      --encoding string   Encoding strategy (default "identity")
  -h, --help              help for encode
      --in string         Input path ('-' for stdin or filename) (default "-")
      --out string        Output path ('-' for stdout or filename) (default "-")
```

### SEE ALSO

* [harp transform](harp_transform.md)	 - Transform input value using encryption transformers

