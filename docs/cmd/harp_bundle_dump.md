## harp bundle dump

Dump as JSON

### Synopsis

Inspect a Bundle object.

Harp Bundles is a structure designed to hold additional properties associated
to a path (package name) and values (secrets). For your pipeline usages, you
can store annotations, labels and user data which can be consumed and/or
produced during the secret management pipeline execution.

The Bundle object specification can be consulted here -	https://ela.st/harp-spec-bundle

```
harp bundle dump [flags]
```

### Examples

```
  # Dump a JSON representation of a Bundle object from STDIN
  harp bundle dump
  
  # Dump a JSON map containing package name as key and associated secret kv
  harp bundle dump --data-only
  
  # Dump a JSON map containing package name as key and associated metadata
  harp bundle dump --metadata-only
  
  # Dump all package paths as a list (useful for xargs usage)
  harp bundle dump --path-only
  
  # Dump a Bundle using a JMEFilter query
  harp bundle dump --query <jmesfilter query>
  
  # Dump a bundle content excluding the template used to generate
  harp bundle dump --skip-template
```

### Options

```
      --content-only    Display content only (data-only alias)
      --data-only       Display data only
  -h, --help            help for dump
      --in string       Container input ('-' for stdin or filename)
      --metadata-only   Display metadata only
      --path-only       Display path only
      --query string    Specify a JMESPath query to format output
      --skip-template   Drop template from dump
```

### SEE ALSO

* [harp bundle](harp_bundle.md)	 - Bundle commands

