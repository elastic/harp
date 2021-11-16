## harp to zookeeper

Publish bundle data into Apache Zookeeper

```
harp to zookeeper [flags]
```

### Options

```
      --dial-timeout duration   Zookeeper client dial timeout (default 15s)
      --endpoints stringArray   Zookeeper client endpoints (default [127.0.0.1:2181])
  -h, --help                    help for zookeeper
      --in string               Container path ('-' for stdin or filename) (default "-")
      --prefix string           Path prefix for insertion
  -s, --secret-as-leaf          Expand package path to secrets for provisioning
```

### SEE ALSO

* [harp to](harp_to.md)	 - Secret container conversion commands

