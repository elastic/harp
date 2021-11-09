# Extend harp using CLI plugins

Heavily inspired by `kubectl` [behavior](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/#writing-kubectl-plugins).

## Scenario

You have repetitive task with harp atomic operations, and you want to assemble
them to create macro commands.

### Prepare the plugin

The plugin manager relies on an executable filename convention. This file could
be written in any language, compiled or not.

For the example, we are going to add a plugin using `bash`.

Create a file named `harp-myplugin` with this content :

```sh
#!/bin/bash

# optional argument handling
if [[ "$1" == "version" ]]
then
    echo "1.0.0"
    exit 0
fi

# optional argument handling
if [[ "$1" == "new-ssh-key" ]]
then
    echo '{{ $key := cryptoPair "rsa" }}{"private":{{ $key.Private | toJwk }},"public":{{ $key.Public | toJwk}}}' | harp template
    exit 0
fi

echo "I am a plugin named harp-myplugin"
```

### Deploy it

Give the execution permission to the file

```sh
chmod +x harp-myplugin
```

Put the file referenced by `PATH` environment variable.

```sh
mv harp-essp /usr/local/bin/harp-myplugin
```

### Check plugin discoverability

```sh
$ harp plugin list
The following compatible plugins are available:

<PATH>/harp-myplugin
<PATH>/harp-server
```

### Try it

```sh
$ harp myplugin
I am a plugin named harp-myplugin
```

Another command

```sh
$ harp myplugin new-ssh-key | jq
{
  "private": {
    ... JWK ...
  },
  "public": {
    ... JWK ...
  }
}
```
