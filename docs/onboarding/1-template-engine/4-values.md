
# Values

You can inject `values` during template compilation :

* unitary value using `--set key=value` flag;
* external JSON/YAML file, using `--values=<filename>` flag;
* set an internal value with external file content, using `--set-file key=<path>`;

`Values` are exposed inside the template using the `.Values` accessor.

For unitary variables :

```sh
$ echo "Hello {{ .Values.hello }}." | harp template --set hello=harp
Hello harp.
```

For external value file :

```sh
$ $EDITOR test.yaml
cloud:
  aws:
    ec2:
      instanceType: m5.large
```

```sh
$ echo "{{.Values.cloud.aws.ec2.instanceType}}" | harp template --values test.yaml
m5.large
```

You can mix, with unitary variable to temporary override variable value:

```sh
$ echo "{{.Values.cloud.aws.ec2.instanceType}}" | harp template --values test.yaml --set cloud.aws.ec2.instanceType=t3.medium
t3.medium
```

For external file content :

```sh
$ echo "{{ .Values.certificate | sha256sum }}" | harp template --set-file certificate=ca.pem
74a8eee22f65cb319a51d034f1cbe22b35cd8de65e5ec61d0e6b033b61667948
```

## Value parsers

By default, value file parsers are detected using the file extension. You can
override the used parser when needed and also set a new root for parsed data.

```sh
--values=<filename>(:<parser>(:<root>)?)?
```

* `filename` is the normal file path to load;
* `parser` is the optional parser to user (YAML, JSON, HCL, HCL2, XML, HOCON);
* `root` is the root object name used to attach the parsed object from file;

Example :

We are going to use the `values` CLI command used to dump the `.Values` object.

```sh
$ harp values
  --values=infrastructure/common/variables/accounts.tf:hcl2:accounts
  --values=infrastructure/cluster-management/alerting/ecstaging.conf:hocon:ecstaging
{
  "accounts": {
    ... Omitted JSONified TF ...
  },
  "ecstaging": {
    ... Omitted JSONified HOCON ...
  }
}
```

---

* [Previous topic](3-variables.md)
* [Index](../)
* [Next topic](5-files.md)
