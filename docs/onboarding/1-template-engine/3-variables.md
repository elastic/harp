# Variables

We want to generate a cryptographic key pair (private/public). The generation
returns both.

> We are using PEM encoding for cryptographic keys.

```sh
$ echo '{{ $k := cryptoPair "ec" }}{{ $k.Public | toPem }}{{ $k.Private | toPem }}' | harp template
-----BEGIN EC PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEVVvxNYwwjJB/4Wb9yL3J2B5zXl1T
FMA/7u5+O9mtaeIPq5HGW62vj48QQR2AaacF5N5FLFCnbCjbhDEx8mwChA==
-----END EC PUBLIC KEY-----
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIDBwv52f2M9EWPNZd9gugHR0ucfI5hqbC9uAuVcrKt5IoAoGCCqGSM49
AwEHoUQDQgAEVVvxNYwwjJB/4Wb9yL3J2B5zXl1TFMA/7u5+O9mtaeIPq5HGW62v
j48QQR2AaacF5N5FLFCnbCjbhDEx8mwChA==
-----END EC PRIVATE KEY-----
```

As you can see `$k` is instanced and used multiple times inside the same
template (the input string).

If you want to publish data in Vault based on generated crypto keys, you could do :

```sh
$ $EDITOR keypair.json
{
    "private": {{ $k := cryptoPair "rsa" }}{{ $k.Private | toJwk}},
    "public": {{ $k.Public | toJwk}}
}
```

If executed like this :

```sh
$ harp template --in keypair.json | jq
{
  "private": {
    "kty": "RSA",
    "kid": "hDgmLVoOFrSq_sJrJ2MYx6f5wXTyAqz3cqip6JnQu5E=",
    "n": "... omitted ...",
    "e": "... omitted ...,
    "d": "... omitted ...",
    "p": "... omitted ...",
    "q": "... omitted ...",
    "dp": "... omitted ...",
    "dq": "... omitted ...",
    "qi": "... omitted ..."
  },
  "public": {
    "kty": "RSA",
    "kid": "hDgmLVoOFrSq_sJrJ2MYx6f5wXTyAqz3cqip6JnQu5E=",
    "n": "... omitted ...",
    "e": "... omitted ..."
  }
}
```

In order to store it in Vault, you have to `flatten` the JSON object as a
`key`:`value` where `value` is a string. To do that you have to encode the JWK
representation of keys using `Base64`.

```sh
$ $EDITOR keypair.json
{
    "private": "{{ $k := cryptoPair "rsa" }}{{ $k.Private | toJwk | b64enc}}",
    "public": "{{ $k.Public | toJwk | b64enc}}"
}
```

This will generate a final secret bundle :

```json
{
  "private": "eyJrdHkiOiJSU0EiLCJraWQiOiJuM0ltZk14bGxuNGtPZFdad0UxeDhKeXVSMFBYUXBaeU9PVDV5b2Nuckp3PSIsIm4iOiIyLVB6Q0c3QzhkZ3IwTXJoaHRVMTJjd3RpMVdZck1FTUpzbUw4bi0zVTIzalI4THVOR1RCUWo5dzZYeG4
  ... omitted ...
  qLWYyQ1NyOGtDeUtIX3ZwMVhZaURjSjdYZ2dqN3FfQTdsQVU3T2RFTFFpSTRoVzVhUl9YTWF0T0JTS2VSclpKeFJQb0NBNFByWVpjOUVGdnFGOCJ9",
  "public": "eyJrdHkiOiJSU0EiLCJraWQiOiJuM0ltZk14bGxuNGtPZFdad0UxeDhKeXVSMFBYUXBaeU9PVDV5b2Nuckp3PSIsIm4iOiIyLVB6Q0c3QzhkZ3IwTXJoaHRVMTJjd3RpMVdZck1FTUpzbUw4bi0zVTIzalI4THVOR1RCUWo5dzZYeG4
  ... omitted ...
  xR0w2Mnd3U0E5TTlMYjZaclEiLCJlIjoiQVFBQiJ9"
}
```

You can publish this bundle in Vault by using the following command :

```sh
$ harp template --in keypair.json | vault kv put secrets/signingkey
Key              Value
---              -----
created_time     2020-07-28T15:33:03.902466Z
deletion_time    n/a
destroyed        false
version          1
```

So that if you retrieve the data from Vault :

```sh
$ vault kv get secrets/signingkey
====== Metadata ======
Key              Value
---              -----
created_time     2020-07-28T15:33:03.902466Z
deletion_time    n/a
destroyed        false
version          1

===== Data =====
Key        Value
---        -----
private    eyJrdHkiOiJSU0EiLCJraWQiOiJtRVJqb0dLcl9zWGotOEhiSUZZeUh1T0ZtVlF0ZHM0a1JSMFI3RlB1S0Y0PSIsIm4iOiJ2SlRyZnNjOUJweDZtbU9ja2N3SjBrSHItTVRoU2JBMEtLRTlLU25hNFAwT2Q4ODdIdGx0NGxZSVJGTWhtQ2....
OXNOUGxrWXpGTURBejlMTjN0XzJoelpsVE9vM09wWThYZ3NpU3lQY0F2cGZiSXBHLUd2ekh1d3pCM25KYlpjdyJ9
public     eyJrdHkiOiJSU0EiLCJraWQiOiJtRVJqb0dLcl9zWGotOEhiSUZZeUh1T0ZtVlF0ZHM0a1JSMFI3RlB1S0Y0PSIsIm4iOiJ2SlRyZnNjOUJweDZtbU9ja2N3SjBrSHItTVRoU2JBMEtLRTlLU25hNFAwT2Q4ODdIdGx0NGxZSVJGTWhtQ2
....
pPUnVfOEtBN211MFEiLCJlIjoiQVFBQiJ9
```

To retrieve the `private` from Vault and display `JWK` encoded key :

```sh
$ vault kv get -field private secrets/signingkey | base64 -D | jq
{
  "kty": "RSA",
  "kid": "mERjoGKr_sXj-8HbIFYyHuOFmVQtds4kRR0R7FPuKF4=",
  "n": "vJTrfsc9Bpx6mmOckcwJ0kHr-MThSbA0KKE9KSna4P0Od887Htlt4lYIRFMhmCnANsHs2ss97uKDMNu7FE0uM0EYjsRaEXbXgZtY6gATJgW3Wz9IPOcyDKReAk-RBAhRDwuq1UNMg1hVKflR9b8AtqSdfFUuPClyAwTyDkHYM4bPpQxFkZBn7hys_MDXSh3CCn1laVaMFiM0yhtpykwWC_qaSGW_fymDX7QqZQ9kryq8_5XS4zZHeXT3BkYh9Ar1zY-AAV-QE3Vtmyf4pWTknSndABqTzFc-hfdMbH2tPL4FiVuJ_C42wbEijAdBgDO8NMwDcAL9hJORu_8KA7mu0Q",
  "e": "AQAB",
  "d": "GbLloMJAA72hS5bViFzY3clUAfO6i9MyWHMYpZpplI2HwjYAZYTV36k_uSfnBRhzdELbJznZN8uweyEGjo6yBBQT56jEaWrblXL5G_JkqaLPyOSH0JzlCioAEaqMJZOIkFdTsXpZW_DWskCita2QyFMIjyAAi_xN6DFtVxoL_4FccVCW226P6F59gZ91d-HGGtyEIdIjfctgMoejc8RPWuLolleNBHPcTA34_JDpmA0r0F57TXOKdAtEHvcb5Ajb-74yyTwvj-5YNqaGkb8Rl0UIl3qKaikdpQivzL-YI-ip4s6HIEP2JJwltzbE-arwZWWvUvho6mWEi7zM7fmGrQ",
  "p": "8jEiTnZjF0PLtmDAn7dhAs2yn8Vk7AGENNGa1qKPNpTDTkrPe5tUl76HmmQvtvUr-8RSq_aYSR3se8_37pROK3ImvktLkYH265NbEkbfbiNpcOEeX_PyhEpky7M2nnP-4jpI3e4IhvNx5Qcqka9BncbxTBCiG92Xn_8ycmmSqqc",
  "q": "x1VUC-pvk7Z8tV3M7yhbhLcy93L6ReBQDxRXAyJ4a1dkJIqcY5esp4afj-boKJxroBSHS9h_13GG0GVAN3gF8OSl8Ik_gzG60DKpLZA073Ueao748Bd32iG1KaEhihyfq6aorH28BM6rLixcnBUrRgE29ze2_fML6XI70Z_8occ",
  "dp": "snd8ZT8d09X0dkcjik4SIYO8PbmynaqcZmOSatlNGRZUOQwtilMy6cLPn_h1pKdTqunHVcqX-0EesznT5C3K0H8Eh7NqUyXm8z8ZnAU3vaxAlZp9zI0xx0QetHAyLl8hkkkKyucNx4v7AJ7gQoxXmNExnDChlFEc2xytatva5P8",
  "dq": "rIc6W6XqNRu6DPDHNCjmLZSzVGH8JQblxGeCeIAZYi8dylL-0WSyV7251b-yKZRZwxCBmjAlVsA4Q5-fWWNVIQ-GYQ8qHc-pNhLjQ0CR0MC6NtjQtl5Zqj-KoiGI-hWUTenODJ43YqHOoARdk-rurYTXolpi1KLNKJ1rESE8dHk",
  "qi": "JY06MGMe4d8oD7oZe1_R6_CLRMNmasYXKfq3v-F7VFJ2Lh0WF-uUD_-ytqP2VnvsL3fAFD6JitE7dJWjK6Lmxl15B-z5oh33Ls99eNWE9vn9sNPlkYzFMDAz9LN3t_2hzZlTOo3OpY8XgsiSyPcAvpfbIpG-GvzHuwzB3nJbZcw"
}
```

## Template scope

You have to be careful with function scope, when using `.<value>` notation. Implicitly,
it's like doing `<current scope>.<value>`.

Especially in `range` iterations :

```rb
{{ range .Values.list }}
  {{ . }}
{{ end }}
```

In this case `{{ . }}` is the current item of list iteration. So if you want to
get access to `{{ .Values.<key> }}`, this is only accessible from the `document root`
scope. You have to use intermediate variable to store the handle.

```rb
{{ $values := .Values }}
{{ range .Values.list }}
  {{ . }} : {{ $values.otherVariableFromValues }}
{{ end }}
```

---

* [Previous topic](2-functions.md)
* [Index](../)
* [Next topic](4-values.md)
