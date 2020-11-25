# Lists and Maps

The main advantage to use the template engine is to be able to render template
based on list or map iterations. So that you can create DRY (Don't repeat yourself)
templates that finally makes final template, so secrets, more consistent.

You can define the following external value file :

```yaml
list:
  - item1
  - item2
  - item3
  - item4
map:
  key1: value1
  key2: value1
  key3: value1
  key-4: value1
```

For `list` processing, just use the `range` keyword :

```sh
$ echo "{{ range $item := .Values.list }}{{ $item }}\n{{ end }}" | harp template --values value.yaml
item1
item2
item3
item4
```

> `$item` will receive each list item value one-by-one.

If you want to get an element from a list :

> Go lists start at index `0`

```sh
$ echo "{{ index .Values.list 2 }}" | harp template --values value.yaml
item3
```

For `map` processing, same `range` keyword but with 2 variables :

```sh
$ echo "{{ range $k, $v := .Values.map }}{{ $k }}: {{ $v }}\n{{ end }}" | harp template --values value.yaml
key1: value1
key2: value2
key3: value3
key-4: value4
```

If you want to get an element from a map `without a "-" in the key name` :

```sh
$ echo "{{ .Values.key1 }}" | harp template --values value.yaml
value1
```

If you want to get an element from a map `with a "-" in the key name` :

```sh
$ echo '{{ index .Values.map "key-4" }}' | harp template --values value.yaml
value4
```

---

* [Previous topic](5-files.md)
* [Index](../)
* [Next topic](7-alternative-delimiters.md)
