## harp bundle patch

Apply patch to the given bundle

```
harp bundle patch [flags]
```

### Options

```
  -h, --help                     help for patch
      --in string                Container input ('-' for stdin or filename) (default "-")
      --out string               Container output ('-' for stdout or a filename)
      --set stringArray          Specifies value (k=v)
      --set-file stringArray     Specifies value (k=filepath)
      --set-string stringArray   Specifies value (k=string)
      --spec string              Patch specification path ('-' for stdin or filename)
      --stop-at-rule-id string   Stop patch evaluation before the given rule ID
      --stop-at-rule-index int   Stop patch evaluation before the given rule index (0 for first rule) (default -1)
      --values stringArray       Specifies value files to load
```

### SEE ALSO

* [harp bundle](harp_bundle.md)	 - Bundle commands

