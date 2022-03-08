## harp lint

Configuration linter commands

### Synopsis

Validate input YAML/JSON content with the selected JSONSchema definition.

```
harp lint [flags]
```

### Examples

```
  # Validate a JSON dump with schema detection from STDIN
  harp lint
  
  # Validate a BundleTemplate from a file
  harp lint --schema BundleTemplate --in template.yaml
  
  # Validate a RuleSet
  harp lint --schema RuleSet --in ruleset.yaml
  
  # Validate a BundlePatch
  harp lint --schema BundlePatch --in patch.yaml
  
  # Display a schema definition
  harp lint --schema Bundle --schema-only
```

### Options

```
  -h, --help            help for lint
      --in string       Container input ('-' for stdin or filename) (default "-")
      --out string      Container output ('' for stdout or filename)
      --schema string   Override schema detection for validation (Bundle|BundleTemplate|RuleSet|BundlePatch
      --schema-only     Display the JSON Schema
```

### SEE ALSO

* [harp](harp.md)	 - Extensible secret management tool

