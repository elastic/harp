## harp from etcd3

Extract KV pairs from CoreOS Etcdv3 KV Store

```
harp from etcd3 [flags]
```

### Options

```
      --ca-file string                 TLS CA Certificate file path
      --cert-file string               TLS Client certificate file path
      --dial-timeout duration          Etcd cluster dial timeout (default 15s)
      --endpoints stringArray          Etcd cluster endpoints (default [http://localhost:2379])
  -h, --help                           help for etcd3
      --insecure-skip-verify           Disable TLS certificate verification
      --key-file string                TLS Client private key file path
      --key-passphrase string          TLS Client private key passphrase
  -k, --last-path-item-as-secret-key   Use the last path element as secret key
      --out string                     Container output path ('-' for stdout) (default "-")
      --password string                Etcd cluster connection password
      --paths strings                  Exported base paths
      --tls                            Enable TLS
      --username string                Etcd cluster connection username
```

### SEE ALSO

* [harp from](harp_from.md)	 - Secret container generation commands

