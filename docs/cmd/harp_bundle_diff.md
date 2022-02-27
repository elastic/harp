## harp bundle diff

Display bundle differences

### Synopsis

Compute Bundle object differences.

Useful to debug a BundlePatch application and watch for a Bundle alteration.

```
harp bundle diff [flags]
```

### Examples

```
  # Diff a bundle from STD and a file based one
  harp bundle diff --old - --new rotated.bundle
  
  # Generate a BundlePatch from differences
  harp bundle diff --old - --new rotated.bundle --patch --out rotation.yaml
```

### Options

```
  -h, --help         help for diff
      --new string   Container path ('-' for stdin or filename)
      --old string   Container path ('-' for stdin or filename)
      --out string   Output ('-' for stdout or filename) (default "-")
      --patch        Output as a bundle patch
```

### SEE ALSO

* [harp bundle](harp_bundle.md)	 - Bundle commands

