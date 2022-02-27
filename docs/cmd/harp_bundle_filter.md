## harp bundle filter

Filter package names

### Synopsis

Create a new Bundle based on applied package matchers.

Filtering a Bundle consists in reducing the Bundle packages using a matcher
applied on the Bundle and Package model to select them, and export them
in another Bundle.

In order to filter packages, you can use :
* a package name selector
* a JMES query
* a REGO policy
* a Set of CEL expressions

Bundle package filtering capabilities are the root of the secret management
by contract. Filter commands can be pipelined to produce complex filtering
pipelines and target the appropriate secrets.

TIP: Use this command to debug your BundlePatch matchers.

```
harp bundle filter [flags]
```

### Examples

```
  # Exclude specific packages by name from STDIN bundle to STDOUT.
  harp bundle filter --exclude "$app/(staging|production)/*"
  
  # Exclude specific packages by name from file bundle to STDOUT
  harp bundle filter --in customer.bundle --exclude "$app/(staging|production)/*"
  
  # Keep specific packages by name
  harp bundle filter --keep "$app/(staging|production)/*"
  
  # Filter packages using a JMES query (context is the package)
  harp bundle filter --query "labels.deprecated == 'true'"
  
  # Filter packages using a JMES query (context is the package) to a file based bundle.
  harp bundle filter --query "labels.deprecated == 'true'" --out deprecated.bundle
  
  # Filter packages using a REGO policy
  harp bundle filter --policy deprecated.rego
  
  # Filter packages using a CEL matcher expressions (associated with AND logic if multiple)
  harp bundle filter --cel "p.match_secret('*Key')"
  
  # Reverse the matcher logic
  harp bundle filter --not <matcher>
```

### Options

```
      --cel stringArray       CEL expression as package filter (multiple)
      --exclude stringArray   Exclude path
  -h, --help                  help for filter
      --in string             Container input ('-' for stdin or filename)
      --keep stringArray      Keep path
      --not                   Reverse filter logic expression
      --out string            Container path ('-' for stdout or filename)
      --policy string         OPA Rego policy file as package filter
      --query string          JMESPath query used as package filter
```

### SEE ALSO

* [harp bundle](harp_bundle.md)	 - Bundle commands

