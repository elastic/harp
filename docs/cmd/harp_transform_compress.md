## harp transform compress

Compress given input

### Synopsis

Compress the given input stream using the selected compression algorithm.

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
harp transform compress [flags]
```

### Examples

```
  # Compress a file
  harp transform compress --in README.md --out README.md.gz --algorithm gzip
  
  # Compress to STDOUT
  harp transform compress --in README.md --algorithm gzip
  
  # Compress from STDIN
  harp transform compress --algorithm gzip
```

### Options

```
      --algorithm string   Compression algorithm (default "gzip")
  -h, --help               help for compress
      --in string          Input path ('-' for stdin or filename) (default "-")
      --out string         Output path ('-' for stdout or filename) (default "-")
```

### SEE ALSO

* [harp transform](harp_transform.md)	 - Transform input value using encryption transformers

