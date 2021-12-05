# Alternative delimiters

In previous cases, harp template engine was rendering a template to create
data. `{{` and `}}` are default template delimiters and interpreted by template
engine during template evaluation.

Sometimes, you might want to create a more complex scenario where harp engine
 is used to render a `meta` template to create a template (and then use `harp`
a second time to render the generated template to create the final data)

> For example, secrets construction based on application configuration files, an
> external list of resources, etc.

In this case, you can use alternative delimiters `[[` and `]]` and let `harp`
know using --alt-delims (`{{` and `}}` won't be interpreted)

Let's create a default template, we want to generate a `paranoidPassword` :

```sh
echo "{{ paranoidPassword }}" | harp template
EOF
```

For example, if we want to generate multiple passwords, i can create a simple template:

```sh
cat << EOF | harp template
{{ paranoidPassword }}
{{ paranoidPassword }}
{{ paranoidPassword }}
EOF
```

With alternative delimiters, we can create a meta template:

```sh
cat << EOF | harp template --alt-delims | harp template
[[ range until 3 -]]
{{ paranoidPassword }}
[[ end -]]
EOF
```

We can add an external value `count` to limit password generation count :

```sh
cat << EOF | harp template --alt-delims --set count=3 | harp template
[[ range until (.Values.count | int) -]]
{{ paranoidPassword }}
[[ end -]]
EOF
```

> Note the `int` conversion, by default all integers are `int64`, you need
> to truncate them to `int` to be able to use the `until` function.

We could set a `default` count, if the external value is not set :

```sh
cat << EOF | harp template --alt-delims --set count=50 | harp template
[[ range until (default 5 (.Values.count | int)) -]]
{{ paranoidPassword }}
[[ end -]]
EOF
```

The interest of meta template is that we can create dynamic template with
external values:

```rb
[[ $values := .Values ]]
[[ range $item := .Values.list ]]
[[ $item ]] : {{ paranoidPassword }}
[[ end ]]
```

and then invoke `harp` :

```sh
harp template --alt-delims --values value.yaml --in template.tmpl | harp template
```

You can compose a complete transformation pipeline for `values` and `template` files.
So that you can prepare, clean, organize data before the final template interpretation.

---

* [Previous topic](6-lists-and-maps.md)
* [Index](../)
* [Next topic](8-whitespace-controls.md)
