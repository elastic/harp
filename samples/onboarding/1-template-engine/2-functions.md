# Functions

You can find all implemented functions in the external library import nammed
[`sprig`](http://masterminds.github.io/sprig/).

## Builtin

In order to be able to generate secret values, we have added secret generation
specialized functions.

### Secret loader

#### secret

```ruby
{{ with secret "secrets/application" }}
{{ .foo }}
{{ end }}
```

This function use parametrable secret loader. You can specify the secret data source,
by using `--secrets-from` CLI flag.

By default, it uses the `vault` secret loader.

You can specificy an secret container path, or use `-` to read secret container
from STDIN so that it will be used as secret data source.

### Password

#### customPassword

```ruby
{{ customPassword <length int> <numDigits int> <numSymbol int> <noUpper bool> <allowRepeat bool> }}
# 128 chars with 16 digits, 16 symbols with repetition
{{ customPassword 128 16 16 false true }}
```

Ouput :

```txt
o)BDz#J|PDyI!+tBKmNSE1lMqh9gfSvVG%juxf9XonBl*N:sb#tgevct9.cDcdAhpt22/MpcbEtM@yM2ofkdhyXgz*0rJOSOkHA97&R78`F1LF4gpq8ZqFntgDSH*5zD
```

#### paranoidPassword

```ruby
{{ paranoidPassword }}
# 64 chars with 10 digits, 10 symbols with upper and lower case and repetition allowed
{{ customPassword 64 10 10 false true }}
```

Output :

```txt
n4[(1[CL6HlNuK95F[qSJd5kUiK.AwV7t)WjKKttgVgn=p9(=0UbrT7vgAhy.VzZ
```

#### noSymbolPassword

```ruby
{{ noSymbolPassword }}
# Same as : 32 chars with 10 digits, no symbol with upper and lower case and repetition allowed
{{ customPassword 32 10 0 false true }}
```

Output :

```txt
V4xQxl7h6QWUr3do70ER5m377cmQaSGX
```

#### strongPassword

```ruby
{{ strongPassword }}
# Same as : 32 chars with 10 digits, 10 symbols with upper and lower case and repetition allowed
{{ customPassword 32 10 10 false true }}
```

Output :

```txt
85SXE7J{29=`^(t68:Ig!9%qU_EH@9b4
```

### Passphrase

#### customDiceware

```ruby
{{ customDiceware <wordCount> }}
# Generate diceware passphrase
{{ customDiceware 6 }}
```

Output :

```txt
brunch-starch-germinate-retool-huntsman-entourage
```

#### basicDiceware

```ruby
{{ basicDiceware }}
# Same as
{{ customDiceware 4 }}
```

Output :

```txt
grill-zit-grading-hamlet
```

#### strongDiceware

```ruby
{{ strongDiceware }}
# Same as
{{ customDiceware 8 }}
```

Output :

```txt
camper-unfilled-moonbeam-veal-vitality-snowdrop-doorman-tinsmith
```

#### paranoidDiceware

```ruby
{{ paranoidDiceware }}
# Same as
{{ customDiceware 12 }}
```

Output :

```txt
sweat-dismantle-county-unlucky-shrank-reaffirm-drainable-mustiness-appendix-scraggly-remindful-sizzling
```

### Crypto

#### cryptoKey

Generate a symmetic encryption/decryption key.

```ruby
{{ cryptoKey <type> }}
# For AES key
{{ cryptoKey "aes:256" }}
```

* `aes:128` => AES128
* `aes:256` => AES256
* `secretbox` => Curve25519 - XSalsa20 / Poly1305
* `fernet` => Fernet encryption key (used by secret service)

Output :

```txt
# AES256
KkkYfArKOMdAjxZkXltIaEUQAK342eQBqYiMZXPyqQM=
# Fernet
UloRDF4uc1-MDqaJCbU9nTG7HJcyzNjIq4zKoERsB5M=
```

#### cryptoPair

Generate asymmetic key pairs.

> Try to use the less precise key type in order to be future proof, generation
> profile is planned for dynamically generate keypair according to targeted
> requirements (fips140).

```ruby
{{ cryptoPair <type> }}
# For RSA recommended (actually RSA2048)
{{ $key := cryptoPair "rsa" }}
# Get the private key
{{ $key.Private }}
# Get the public key
{{ $key.Public }}
```

Where `type` could be :

* `rsa` , `rsa:normal` , `rsa:2048` => RSA 2048
* `rsa:strong`, `rsa:4096` => RSA 4096
* `ec`, `ec:normal` , `ec:p256` => EC P-256
* `ec:high`, `ec:p384` => EC P-384
* `ec:strong`, `ec:p521` => EC P-521
* `ssh`, `ed:25519` => Ed25519
* `naclbox` => Curve25519

#### toJwk

Encode the given cryptoKey as JWK.

```ruby
{{ $key := cryptoPair "ec:p384" }}
# Get the private key and encode it as JWK
{{ $key.Private | toJwk }}
# Get the public key and encode it as JWK
{{ $key.Public | toJwk }}
```

Output :

```ruby
# Get the private key and encode it as JWK
{{ $key.Private | toJwk }}
{
    "kty":"EC",
    "kid":"8rvz08-Aq05Vq-a40dpJFt5VwvAgdfJPGt9TKkchNUM=",
    "crv":"P-384",
    "x":"KfTYa3f9WKgg5npBsBfw6ivTJgQS0xP2KbvQHU4WtEzllvjOsz1D2WZCPq9X-aUq","y":"88SZwdKWNb3GONuO0C8LqI3aCtTBf2SCOiKgLNLinWSH_Dval0_euuCv8WRTVYcL","d":"jIcdBVkUfXs1U5SbtcmH2aqL6vXJTMmBtK9SFaoi9HDmSb7VeQSvMQZmUzDTgn9N"
}
# Get the public key and encode it as JWK
{{ $key.Public | toJwk }}
{
    "kty":"EC",
    "kid":"8rvz08-Aq05Vq-a40dpJFt5VwvAgdfJPGt9TKkchNUM=",
    "crv":"P-384",
    "x":"KfTYa3f9WKgg5npBsBfw6ivTJgQS0xP2KbvQHU4WtEzllvjOsz1D2WZCPq9X-aUq","y":"88SZwdKWNb3GONuO0C8LqI3aCtTBf2SCOiKgLNLinWSH_Dval0_euuCv8WRTVYcL"
}
```

#### toPem

Encode the given cryptoKey as PEM.

```ruby
{{ $key := cryptoPair "rsa" }}
# Get the private key and encode it as PEM
{{ $key.Private | toPem }}
# Get the public key and encode it as PEM
{{ $key.Public | toPem }}
```

Output :

```ruby
# Get the private key and encode it as PEM
# {{ $key.Private | toPem }}
"-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAvrvdBDuPyJqrkgJjC1dyCVavBZSHtbw0K1HNAM4FljNNn6NQ
zw2mSPsg52rBQgvZhXOyB4dJ6TKG0ru6yFYEOnpradeVTgOVWmAnUrjj/gLNMAk7
ig5lXbDF5IzpKC3h1dy95SNtY0ciUfpFkKwFdKPzed/gdwTNfAG3qGHScdDYQ79I
L+fHsDv/bEqoKiYmYdPtKu93VPy30b1Vq29eoa2nzPrJU/XtbU8E4RJvVAfIEXeI
bjsARukufmi01BqDobbkTAQsRnHyWEMJClYO5kZGPoZP54A+QsINuLH+EYyID9ji
j9eGHU0Q9nu9ToVVagUZFNOeb/C8FTUgTjgvFwIDAQABAoIBAQCTW785Cu9eVCkj
6FYvKANBtcgI7qUewmYn5c4VxhZI4iAhquK+5VPIZMUaJb9j0JXg5e7wpBR1Z4UE
gOeg6dvgdj5Qiq+ek2Qra8hVv+TdlbqDV065rb+g7/ckSb3FPhWRzjakNofHwOiC
H3LpgA3C/PwZ996K9ZnwKb+EBve42AMjkaL1CxZepAQ26gzgN95sIElQ7IluuTw8
/TDSY2YfxkMo9MXGCHzNYbXEZczGgh57V21Y5Z7WrsbmH4E07qDNJYF0LSY/+T6B
XhPGsoRGMTnAowNoFMcPqDIdZWaGU7v1o/tmukOjLsEdCaWhJTpQB7lSAL0DPRbv
5BvZ1RqpAoGBAPYJM5Yt1lo56HYxuEW8ywITJgRHi9POjhXrpRo9ZWdRilLYo6Zj
IT4YsFQM9oISbONn64giBMqcTMQG/rmCQEof2MRfmzQQpY8Y2iOmpDG5pO383EdQ
iYM6BMeEisDOqX0HhjpoLwoKlSJEZzUabMCZjeATgNJxJ3WjH5gns8MlAoGBAMZ1
UAblDTBGpsNxnr9YRhnBaU7LtSt9UWDu6P627hb86hkucbUDrZ+FaFkjC6AZrGHG
iIjJoDMgdoncD8LNOXnOM5P5qe6NdjNqOI9ffJlnwe5PFDQ/ac8uGMYOf6tCw6TZ
Wl9BCS27u7UuAkGeK+NIMP+M3lT/YJRzFwt9gDKLAoGBANCxgWZzvwyNbhdDmVDe
ETzTTT34Ci1BWdhSJ5uYVHlM+w3G4RlzoHDxtC+3RymRw3cpYOn6ISJTbfIhFNP6
HdpCJTZ8+kMxk51LsUzoPwJGvBV6lMaRE/ORtRgf3yoooi+BwGOul6fmzhVg/EJZ
BcJg/a0CHhVjEduA4H3Jv3tZAoGAUNG7emNTIKLVDOi7bk8DlT+HpDgfGovZVTFW
H0zd3uy2ZPTeB4ps7XbFzO8Rr+xkoBjax2Hc5JVG0NOWc41h57HKnWtiAa0IQt3y
FKkdM9fmSSdZIgHlFCNAoX+MDHGO/RYq0HnKxB4czibjcld4pgFjOt7iOBkb+rh3
3Q0J5QsCgYARPewmPn2kmg5bLd6xEf50AMVLWLU7Hfll5f31sNjqpoFnht0QCAlz
OK1V3HKLEquCNpxHXVVRr4vUjxNw068QgG5ZQ1el3DY0TIelFBo40WX3M74szqir
k1mBN96xZGjAnuhEHjnd2/xe4KYxwQnTQjEoJKY0TQc0qTcXHv3gzA==
-----END RSA PRIVATE KEY-----"

# Get the public key and encode it as PEM
# {{ $key.Public | toPem }}
"-----BEGIN RSA PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvrvdBDuPyJqrkgJjC1dy
CVavBZSHtbw0K1HNAM4FljNNn6NQzw2mSPsg52rBQgvZhXOyB4dJ6TKG0ru6yFYE
OnpradeVTgOVWmAnUrjj/gLNMAk7ig5lXbDF5IzpKC3h1dy95SNtY0ciUfpFkKwF
dKPzed/gdwTNfAG3qGHScdDYQ79IL+fHsDv/bEqoKiYmYdPtKu93VPy30b1Vq29e
oa2nzPrJU/XtbU8E4RJvVAfIEXeIbjsARukufmi01BqDobbkTAQsRnHyWEMJClYO
5kZGPoZP54A+QsINuLH+EYyID9jij9eGHU0Q9nu9ToVVagUZFNOeb/C8FTUgTjgv
FwIDAQAB
-----END RSA PUBLIC KEY-----"
```

#### encryptPem

Encrypt the given PEM with a passphrase.

```ruby
{{ $key := cryptoPair "rsa" }}
# Generate a passphrase
{{ $passphrase := paranoidDiceware }}
# Get the private key and encode it as PEM, then encrypt with passphrase
{{ $privPem := toPem $key.Private }}{{ encryptPem $privPem $passphrase }}
# Get the public key and encode it as PEM
{{ $key.Public | toPem }}
```

Output :

```ruby
# {{ $key := cryptoPair "rsa" }}{{ $passphrase := paranoidDiceware }}{{ $privPem := $key.Private | toPem }}{{ $passphrase }}\n{{ encryptPem $privPem $passphrase }}
"helmet-flashcard-context-tidiness-osmosis-sled-shimmer-jeeringly-exhale-aloof-defuse-pranker
-----BEGIN RSA PRIVATE KEY-----
Proc-Type: 4,ENCRYPTED
DEK-Info: AES-256-CBC,37b6974ddc3b42c7730cc9b66b3c46ca

d2KtoDLpW17S3waRsiCZBK+hbbmhL6/xjc61AJsf9kTshBP/5ZnbDNq0catp55r0
wNicSQUca/6iJkCEStbp7nkhobTOXlldVIBaL0ChsQkvArJ3sxqfpRFJmDwOF48u
2oKETF92ZgvZultU45iBjJ2uRrmulAJs4we8Dgo8TkjKz9yaGF5gNYloFl69oDfF
KAx5+f1dtqtaUQmKcivKmFej1e2Zn/QNOjLu5DabfycH9TcRgV+Y3NnffG9rujZE
E/SKX9fbsmoY/N8pRBLuxJ84t1Fe//qRnWGU1I2T3HVbxoZcdBiPAs1gK/oYRDlu
b+E8hDGeDNHE69vvrHdzwdvYsLhDG0rimyur2modIuITOjWep+fAOY4mQtV3+15g
OlMMQ5MQUdoZTg99oIBmuYfb1thLog+UvXsVXl53Lu1v+Oz9mUVjhy/Xc8wzLsUk
wXd6mo4maTzNxeqE6b6WtNPRlTjKGz1B+9NgWCxRJ5fEuRsA48RycxXuERSLWQXR
nTco/pQDMUppxKZnNRtEwxG21i/eZkIrapTsUlCKSiyl31/Ux1atfrOrPxXHAz9h
frnEloHfXz2p0JezQ9aKn0RQEFnGm5zPgDs9znHkFmDtG0P2F3UXFm80oNXjWFDS
JMK7tnJlS1K1iipMbxlI2b+9w1apD5uXPYYIlBDLZpMdqKPMVywmEQqhw54kqTqi
8E98EWd9FxPhqHq2tSE+dq/YS+3CGDLjNiVa8jp+vr0GOGI6gQ+1JbPdD20KCzkZ
jxPZlCDPw2GLtL+uD13aPfgIagKdTPqCoa58G81t3hKVpL7ocPQU/gg2iBmcnTkz
DAZfTY7LmYalgVyTVZaoy6YC5UouAPSdu+YhUHBR6KLWUQddboKJYdpBb9tbgGbd
qWgUPFT+7OhuA/ub9TJotlhFct4AkqEkfhgu88mx4rgYIXfQmHc3E6YMH+WcopVB
foFFSLSM+8eiw9zHVbsVaAsPBpWJb7i5UC98KpEE8aOx17nYraMuhsqOIKaohJoo
p6zb1U/aK1dOhMTBQqWYpmSYH1oo29dvzT2QlLbzQ2OLT5nIOHNXBXLrdHXUpAm+
jRSTVN02fTLiT5CFJ2XjwVRZu947piICZwU7JH/D/C6zIvinJ+4mnjPZ3AV8bZet
2XeQ66b/UiMm9lVehIA9KnEbasjqLqyulGaUmO+Zmd4AcCoKcP137Q+ma72MExqk
Qu4J1goWiL3ZGdFMf1wdzNf/XONzlLK2zvysvmNhnT3NWtj+QdM+yui3px2QbI09
Fq6fGOFuCCP35U0tFtI5kT/I6Bm0y/eMBuoGPQvfNf/48VlwswILHabT+x0SfYJH
rh9v505XT/jP7YVdETgRNTju5NIZmEgoos4aB7zjZcIKpUxz/dyOZ7DtGHcrAnrq
ZB8vJycdejWmvUhmXIchqdRrDd0WTZlikIU3I6zeP1uHoQNBCrKV/bsKnLItleo4
4gvZvd9gYEnn4mx+7Q96XmiidV02PIZiZykZR2vmGE+BTEtFb6S+q4ZylBhq9/o+
rOEinWY5gWfVV+515pCdVWOOSyfwzCx0imnIKC2d4fkMsrNM1xmAZtm/+MYudss9
-----END RSA PRIVATE KEY-----"
```

#### toSSH

Encode the given key for OpenSSH usages.

```ruby
{{ $key := cryptoPair "ssh" }}
# Get the private key and encode it as OpenSSH private key
{{ $key.Private | toSSH }}
# Get the public key and encode it as PEM
{{ $key.Public | toPem }}
```

Output :

```ruby
# {{ $key := cryptoPair "ssh" }}{{ $key.Private | toSSH }}\n{{ $key.Public | toPem }}\n{{ $key.Public | toSSH }}
"-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtz
c2gtZWQyNTUxOQAAACAzYJV8yhDVUKmycpjGBCuO8rO9vbZleBhEvTuLmAVxkAAA
AIh29lFUdvZRVAAAAAtzc2gtZWQyNTUxOQAAACAzYJV8yhDVUKmycpjGBCuO8rO9
vbZleBhEvTuLmAVxkAAAAECvfToeBka1FO6I6jovwvZXEXEei9ACZM5ImPTzuxAM
DTNglXzKENVQqbJymMYEK47ys729tmV4GES9O4uYBXGQAAAAAAECAwQF
-----END OPENSSH PRIVATE KEY-----

-----BEGIN PUBLIC KEY-----
MCowBQYDK2VwAyEAM2CVfMoQ1VCpsnKYxgQrjvKzvb22ZXgYRL07i5gFcZA=
-----END PUBLIC KEY-----

ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDNglXzKENVQqbJymMYEK47ys729tmV4GES9O4uYBXGQ"
```

---

* [Previous topic](1-introduction.md)
* [Index](../)
* [Next topic](3-variables.md)
