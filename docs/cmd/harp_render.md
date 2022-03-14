## harp render

Render a template filesystem

### Synopsis

Generate a config filesytem from a template hierarchy or archive.

```
harp render [flags]
```

### Examples

```
  # Generate a configuration filesystem from a folder hierarchy
  harp render --in templates/database --out postgres
  
  # Generate a configuration filesystem from an archive
  harp render --in templates.tar.gz --out configMap
  
  # Test template generation
  harp render --in templates.tar.gz --dry-run
```

### Options

```
      --alt-delims                 Define '[[' and ']]' as template delimiters.
      --dry-run                    Generate in-memory only.
  -h, --help                       help for render
      --in string                  Template input path (directory or archive)
      --left-delimiter string      Template left delimiter (default to '{{') (default "{{")
      --out string                 Output path
      --right-delimiter string     Template right delimiter (default to '}}') (default "}}")
      --root string                Defines file loader root base path
  -s, --secrets-from stringArray   Specifies secret containers to load ('vault' for Vault loader or '-' for stdin or filename) (default [vault])
      --set stringArray            Specifies value (k=v)
      --set-file stringArray       Specifies value (k=filepath)
      --set-string stringArray     Specifies value (k=string)
  -f, --values stringArray         Specifies value files to load
```

### SEE ALSO

* [harp](harp.md)	 - Extensible secret management tool

