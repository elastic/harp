## harp template

Read a template and execute it

```
harp template [flags]
```

### Options

```
      --alt-delims                 Define '[[' and ']]' as template delimiters.
  -h, --help                       help for template
      --in string                  Template input path ('-' for stdin or filename) (default "-")
      --left-delimiter string      Template left delimiter (default to '{{') (default "{{")
      --out string                 Output file ('-' for stdout or a filename)
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

