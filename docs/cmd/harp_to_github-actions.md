## harp to github-actions

Export all secrets to Github Actions as repository secrets.

```
harp to github-actions [flags]
```

### Examples

```
$ export GITHUB_TOKEN=ghp_###############
$ harp to gha --in secret.container --owner elastic --owner harp --secret-filter "COSIGN_*"
```

### Options

```
  -h, --help                   help for github-actions
      --in string              Container path ('-' for stdin or filename) (default "-")
      --owner string           Github owner/organization
      --repository string      Github repository
      --secret-filter string   Specify secret filter as Glob (*_KEY, private*) (default "*")
```

### SEE ALSO

* [harp to](harp_to.md)	 - Secret container conversion commands

