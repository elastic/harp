## harp transform decompress

Decompress given input

### Synopsis

Decompress the given input stream using the selected compression algorithm.

Supported compression:
  * identity - returns the unmodified input
  * gzip
  * lzw/lzw-msb/lzw-lsb
  * lz4
  * s2/snappy
  * zlib
  * flate/deflate
  * lzma
  * zstd

```
harp transform decompress [flags]
```

### Examples

```
  # Compress a file
  harp transform decompress --in README.md.gz --out README.md --algorithm gzip
  
  # Decompress to STDOUT
  harp transform compress --in README.md.gz --algorithm gzip
  
  # Decompress from STDIN
  harp transform compress --algorithm gzip
```

### Options

```
      --algorithm string   Compression algorithm (default "gzip")
  -h, --help               help for decompress
      --in string          Input path ('-' for stdin or filename) (default "-")
      --out string         Output path ('-' for stdout or filename) (default "-")
```

### SEE ALSO

* [harp transform](harp_transform.md)	 - Transform input value using encryption transformers

