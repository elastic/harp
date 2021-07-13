# Functions

You can find all implemented functions in the external library import nammed
[`sprig`](http://masterminds.github.io/sprig/).

## Builtin

In order to be able to generate secret values, we have added secret generation
specialized functions.

### Encoders

#### b64urlenc / b64urldec

Apply BASE64 URL encoding to given input.

```ruby
{{ paranoidPassword | b64urlenc }}
fjYySGJoa00iQkdTaXRUQ2d-RVgwfHMwI2tvcG5Yc0xne3RfQV9HZU5YQ3ZTT243XWUyeDVqNjVNQnRMJEdzNA==
```

Decode a BASE64 URL encoded string.

```ruby
{{ "fjYySGJoa00iQkdTaXRUQ2d-RVgwfHMwI2tvcG5Yc0xne3RfQV9HZU5YQ3ZTT243XWUyeDVqNjVNQnRMJEdzNA==" | b64urldec }}
~62HbhkM"BGSitTCg~EX0|s0#kopnXsLg{t_A_GeNXCvSOn7]e2x5j65MBtL$Gs4
```

#### bech32enc / bech32dec

[Bech32](https://en.bitcoin.it/wiki/Bech32) is an encoding used for many wallet
address in blockchain space.
You can encode a binary array with a human readable prefix (HRP), very useful to
encode crypto-material and keep ownership visible to humans.

> This is the encoding used by container sealing identities.

```ruby
{{ bechenc <HRP> <[]BYTE> }}
```

For example with an Ed25519 Public key:

```ruby
{{ $key := cryptoPair "ed25519" }}
{{ bech32enc "security" $key.Public }}
security19f29qq5vq73tdrhspdzkqcdf2exewg2g6xcxe5h74y72qsv7c00sx57ny0
```

#### shellescape

Apply Shell escaping strategy to allow a string to be safely used in a shell script.

```ruby
{{ paranoidPassword | shellescape }}
'tGO48jRkfOiXv8=p?eV^wi7tqJz`ABeQy1ZXk2WE(E1XWuS6%$j+X>QVx93W*WEY'
```

#### urlPathEscape / urlPathUnescape

Apply url character escaping strategy for components used in path

```ruby
https://ingester.es.cloud/{{ tenant | urlPathEscape }}/api/v1
```

#### urlQueryEscape / urlQueryUnescape

Apply url character escaping strategy for components used in query

```ruby
https://logstash:{{ paranoidPassword | urlQueryEscape }}@ingester.es.cloud:1234
https://logstash:K3iDayow9%5Cav67HawD6%210k~8lhcm8oLVUBt2wE%3E%5DLBJQJVj%3AfIx%2Fuo%40%7B%3D6kvgXHK@ingester.es.cloud:1234%
```

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
* `ssh`, `ed25519` => Ed25519
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

#### fromJwk

Decode a JWK encoded key.

```ruby
{{ $key := fromJwk .Values.jwk }}
# Convert JWK encoded key to native one
{{ $key.Private | toJwk }}
# Get the public key and encode it as JWK
{{ $key.Public | toJwk }}
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
-----BEGIN ENCRYPTED PRIVATE KEY-----
MIIFNjBgBgkqhkiG9w0BBQ0wUzAyBgkqhkiG9w0BBQwwJQQQQvKLOxmTrmsNvR6x
skoU4wIDAYagMAwGCCqGSIb3DQIJAAAwHQYJYIZIAWUDBAEqBBBzz11Ee0eilGVC
rnT7s8ITBIIE0DHRRvy/8XsDWZ64b0huVdQpv3BiXeUATb0c+i1neZo4btaR5prG
nz/LK6/XbLgUkzgcC5cYEfi5bTkPqFDi4raa8gBryGF95k7akm4XJJY3Os1MIrCS
MQinWfjr8WqkbPmMe5hZyv6PBuPOPrBcw5M/rhTPIePshvvyWi4cXv1871MCDB5p
93jhNTNqGB/sbqSbW/qlanYhhaPZuuUZld4Jae9Y9WAvz42T/C/uj6LEQE0+lOGa
Y1rNYWdkkeoeBhHj8GHfYydTDusxzbMz37VCtdPJQnNtQU6zaCdReC0wVbLcFwWA
NYghMAjoSFrWhB26htq+Ob+DMLlhbAnVu9iTuEV3C/qe/sFrN2iUjN6lBQh1SISR
gKHj8dclduMgSQiIKqk+rB89wz3dJIspq93AAzFKbx2MSD9LP70EM2PRaezIjnGR
irWx6L0xuhPWPCn5oSJpvw7gVuxn7jN6SK/EyMP6fEEd4dN6tNh5XsVEY0jXEw03
o/MVDeEcmewZwzFpKeIiJYd6IqnwScLqWFeLNZPfi8EKv/hOFVr3fRDH98FKVJer
xeMA73wBeBG7zPcQga14AlPrJBUHAeH6SObbd0G4k7QDntuYerVU8GQ+5akJw5HL
1a83uKi85CqtB2QnCFbdhdgOV7X8rxOBd8jlHOzXXkm57sk8sfKWYGzrqtltbGjm
jgbODDJEX2Jmk/DVhvr60J6m3Y8pyvNIoBK3V9qu2wezRsONX5ZFS6VXan0OBcjQ
1EwdgMkSci01Gs5oj9s3MPM9FS65arKTCsE9riWfUqLkJBZ6/mHwaqbw1LoYFqev
WSvl9IyBCGt8YZVnvggYBc4hmmfubn3XhbjZGPDJwh1beOP3TOlTQqPaLsaTCB9F
NBp3WbSW4Ff3O2c0/gdmLmIMed2nLNVJZSVO8rfhV8k18iDjKwm5hboeuY0CucR1
7+k/HrrqxsjixhQuyaY5SeGn8AWmJi7tdStExMwiCX9t5zgzL+V1pM/Xf6SG7dij
EcIfyPrVWGk//0lvBo6jSjtOh8n+JM/9rPQQEngHUhI+h/8QY06AVSCpbgo2vQJl
Us6Gcc+YBUD04D4lLmI7nSBWNS46Zilirtkmwp61EG/ysB+5MW8oN4212QqM8yrV
n0UTIsrJn8X2sQ+yLX4qEzA1bH26pJIkwKHfQktb+RrqokgF9uB9kI9UzkBIN8ns
ZQ3B1+bqzE9z0EWdGd79yNH1SW6LVDDaEVVeV1lrSBSkhAAS86+TUf5rNKhYeqU9
m3dzPJAl3MTWhXpxQ/olJVb3fES2ZT5EGxVlt36N3p2tUf/GFwaUi8FyNYenGEn0
R7Vi0TdmOLrhJJI+PzqXceYcLh13MD3XRHbpBR/FqnAf7cVAnxiqjRVgrHHAAq0+
y3fGPusAt2ANEvl2Hk1qFqQX1oNhRkoqJpBgDc6ouqNV4rEdx+kwBDSoZ7ahWzr0
lt3bNj4XJ8a74WKdilbWmpM4zonXMovOJ8e2lhXF069B1X6QijSyrXZaOjEobJvH
mIXCH8ZhRiB/qfLllwLKzdKh961/Mm4mxIC1/FqraGtlC0jkZXZR69huLNNUBGry
JFS6EYe2rXuOqxSTzurUTPC4U3bBwtCwTpG/YVAzIkiL7BDfhxB0X5aG
-----END ENCRYPTED PRIVATE KEY-----"
```

#### encryptJwe

Encrypt input using JWE.

```ruby
{{ $key := cryptoPair "rsa" }}
# Get the private key and encode it as PEM
{{ $pk := toPem $key.Private }}
# Encrypt private key
{{ encryptJwe $passphrase $pk }}
```

#### decryptJwe

Decrypt input encoded as JWE.

```ruby
{{ $key := cryptoPair "rsa" }}
# Get the private key and encode it as PEM
{{ $pk := toPem $key.Private }}
# Encrypt private key
{{ $encrypted := encryptJwe $passphrase $pk }}
# Decrypt JWE
{{ decryptJwe $passphrase $encrypted }}
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

#### toJws

Create a JWT.

```ruby
{{ $key := cryptoPair "ec" }}
{{ $claims := fromJson "{\"sub\":\"test\"}" }}
{{ toJws $claims $key.Private }}
```

---

* [Previous topic](1-introduction.md)
* [Index](../)
* [Next topic](3-variables.md)
