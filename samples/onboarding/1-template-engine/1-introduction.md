# Introduction

`harp` exposes a template engine full of functions to allow you to handle
simple to complex operations.

Template language is based on [Go template language](https://blog.gopheracademy.com/advent-2017/using-go-templates/).

A template is a string, array of characters, without any type (YAML, JSON, etc.),
just a string. The final format is decided by your usages of this template language.

```sh
$ echo "{{ paranoidPassword }}" | harp template
bOLqnUZU%J@9k}df5h4h:@9a+l]hrraT3yO=VfTNT2PO_kygAcY3r2Wf4W2kNN|R
```

> `paranoidPassword` is a template function that generate a 64 chars with
> 10 digits, 10 symbols with upper and lower case and repetition allowed. It's a
> password because we decided to use this string value for password usage.

If we want to use this a a password and store it in Vault, we could do :

```sh
$ echo "{{ paranoidPassword }}" | harp template | vault kv put secrets/database password=-
Key              Value
---              -----
created_time     2020-07-28T14:55:34.125458Z
deletion_time    n/a
destroyed        false
version          1
```

So that it will generate a random string with `paranoid` pre-decided constraints
and put the `string` as a `password` property value of a `secret` addressed by
`secrets/database` path.

A `secret` is a data addressed by a path, that contains properties.

```sh
$ vault kv get secrets/database
====== Metadata ======
Key              Value
---              -----
created_time     2020-07-28T14:55:34.125458Z
deletion_time    n/a
destroyed        false
version          1

====== Data ======
Key         Value
---         -----
password    XBAQp]!VDIm5nIP3mHX0E5l-y#8gAGY1Ex!=kk+bn%g8H1shc9iH_RVXeaOTk?9h
```

A Vault entry has a `path`, `data` and `metadata` associated to the given entry.

If we want to retrieve a `data` only :

```sh
$ vault kv get -field data secrets/database
map[password:XBAQp]!VDIm5nIP3mHX0E5l-y#8gAGY1Ex!=kk+bn%g8H1shc9iH_RVXeaOTk?9h
]
```

And `password` only :

```sh
$ vault kv get -field password secrets/database
XBAQp]!VDIm5nIP3mHX0E5l-y#8gAGY1Ex!=kk+bn%g8H1shc9iH_RVXeaOTk?9h
```

> This is important to understand that Vault is a distributed highly
> available encrypted key / value store.

We have provisioned a single password in Vault using the `harp` template
engine and the Vault CLI directly. But we need to expand this usecase to allow
multiple secret provisioning as a parametrized `Bundle` of various `secret`
types.

---

* [Next topic](2-functions.md)
