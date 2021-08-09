## harp completion zsh

generate the autocompletion script for zsh

### Synopsis


Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions for every new session, execute once:
# Linux:
$ harp completion zsh > "${fpath[1]}/_harp"
# macOS:
$ harp completion zsh > /usr/local/share/zsh/site-functions/_harp

You will need to start a new shell for this setup to take effect.


```
harp completion zsh [flags]
```

### Options

```
  -h, --help              help for zsh
      --no-descriptions   disable completion descriptions
```

### SEE ALSO

* [harp completion](harp_completion.md)	 - generate the autocompletion script for the specified shell

