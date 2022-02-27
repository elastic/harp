## harp bundle lint

Lint the bundle using the given RuleSet spec

### Synopsis

Apply a RuleSet specification to the given bundle.

This command is used to check a Bundle structure (Package => Secrets).
A control gate could be implemented with this command to enforce a bundle
structure by decoupling the bundle content and the usage contract.

```
harp bundle lint [flags]
```

### Examples

```
  # Lint a bundle from STDIN
  harp bundle lint --spec cso.yaml
```

### Options

```
  -h, --help          help for lint
      --in string     Container input ('-' for stdin or filename) (default "-")
      --spec string   RuleSet specification path ('-' for stdin or filename)
```

### SEE ALSO

* [harp bundle](harp_bundle.md)	 - Bundle commands

