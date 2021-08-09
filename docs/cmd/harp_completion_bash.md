## harp completion bash

generate the autocompletion script for bash

### Synopsis


Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:
$ source <(harp completion bash)

To load completions for every new session, execute once:
Linux:
  $ harp completion bash > /etc/bash_completion.d/harp
MacOS:
  $ harp completion bash > /usr/local/etc/bash_completion.d/harp

You will need to start a new shell for this setup to take effect.
  

```
harp completion bash
```

### Options

```
  -h, --help              help for bash
      --no-descriptions   disable completion descriptions
```

### SEE ALSO

* [harp completion](harp_completion.md)	 - generate the autocompletion script for the specified shell

