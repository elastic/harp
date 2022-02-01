# Encryption algorithm

- [Encryption algorithm](#encryption-algorithm)
  - [Symmetric encryption](#symmetric-encryption)
    - [Raw output](#raw-output)
      - [Authenticated Encryption (AE)](#authenticated-encryption-ae)
        - [SecretBox](#secretbox)
      - [Authenticated Encryption with Associated Data (AEAD)](#authenticated-encryption-with-associated-data-aead)
        - [AES-GCM](#aes-gcm)
        - [AES-SIV](#aes-siv)
        - [AES-PMAC-SIV](#aes-pmac-siv)
        - [Chacha20-Poly1305](#chacha20-poly1305)
        - [XChacha20-Poly1305](#xchacha20-poly1305)
      - [Deterministic Authenticated Encryption (DAE)](#deterministic-authenticated-encryption-dae)
        - [How to use DAE?](#how-to-use-dae)
        - [DAE-AES-GCM](#dae-aes-gcm)
        - [DAE-AES-SIV](#dae-aes-siv)
        - [DAE-AES-PMAC-SIV](#dae-aes-pmac-siv)
        - [DAE-Chacha20-Poly1305](#dae-chacha20-poly1305)
        - [DAE-XChacha20-Poly1305](#dae-xchacha20-poly1305)
    - [Encoded output](#encoded-output)
      - [Fernet](#fernet)
      - [JWE](#jwe)
        - [With a key](#with-a-key)
        - [With a password/passphrase](#with-a-passwordpassphrase)
      - [PASETO](#paseto)
  - [Asymmetric encryption](#asymmetric-encryption)
    - [AGE](#age)
  - [Envelope encryption](#envelope-encryption)

## Symmetric encryption

### Raw output

The follow transformers return a RAW byte stream which need to be encoded.

#### Authenticated Encryption (AE)

##### SecretBox

Secretbox is a cipher suite based on :

* Encryption: XSalsa20 stream cipher
* Authentication: Poly1305 MAC

```sh
$ harp keygen secretbox
secretbox:FSyb2IPGl2AymGBgLVWCFfuDjbf2iQIuYTl513T1sow=
$ echo -n "test" \
    | harp transform encrypt --key paseto:HTh5iRgEEVDwpcyJEQ6pYx7S5YCGHGB729emBy2e-K4= \
    | base64
UAnsgOOawkV7uUyxBFeqauCgqf0ZbP6t8cFA/V49Xe8yTk7hRGG+Mb6bfWU=
```


#### Authenticated Encryption with Associated Data (AEAD)

##### AES-GCM

AES with Galois/Counter Mode (AES-GCM) provides both authenticated encryption
(confidentiality and authentication) and the ability to check the integrity and
authentication of additional authenticated data (AAD) that is sent in the clear.
AES-GCM is specified in NIST Special Publication 800-38D [SP800-38D].

```sh
$ harp keygen aes-gcm
aes-gcm:yOF_27OF4aokpB_6WUCBrg==
$ echo -n "test" \
    | harp transform encrypt --key aes-gcm:yOF_27OF4aokpB_6WUCBrg== \
    | base64
kCi/ReYbE1yFdS1UPsH771ToiznJjIvNfIyTnCaErNg=
```

##### AES-SIV

AES-SIV is an authenticated mode of AES which provides nonce reuse misuse resistance.
Described in RFC 5297, it combines the AES-CTR (NIST SP 800-38A) mode of encryption
with the AES-CMAC (NIST SP 800-38B) function for integrity.

```sh
$ harp keygen aes-siv
aes-siv:glmpGwiMb3bop4XwZeYvfGPZSn_87uAnC75vp73tJFAvfP6mdLX5dgTYj0cgua_tUZ7itQ-MFwNF-ZCGzr0Fxw==
$ echo -n "test" \
    | harp transform encrypt --key aes-siv:glmpGwiMb3bop4XwZeYvfGPZSn_87uAnC75vp73tJFAvfP6mdLX5dgTYj0cgua_tUZ7itQ-MFwNF-ZCGzr0Fxw== \
    | base64
1gcAdkk+fSLOwOu7JvbT2/NDddmwT0m/kWU+JTCxDOkWAM/P
```

##### AES-PMAC-SIV

AES-PMAC-SIV is a fully parallelizable variant of AES-SIV. AES-PMAC-SIV provides
effectively identical security properties as the original AES-SIV construction,
including nonce reuse misuse resistance, but also performs significantly better
on systems which provide parallel hardware implementations of AES, namely
Intel/AMD CPUs but also certain IoT devices.

```sh
$ harp keygen aes-pmac-siv
aes-pmac-siv:5W5c43ZFVX37Y4p7tlksEyyuw6R_VpF68QoczfNUnpameV_63Kif0byRd-KFA-svBW5eXy2D_9h_S02xdWKEKA==
$ echo -n "test" \
    | harp transform encrypt --key aes-pmac-siv:5W5c43ZFVX37Y4p7tlksEyyuw6R_VpF68QoczfNUnpameV_63Kif0byRd-KFA-svBW5eXy2D_9h_S02xdWKEKA== \
    | base64
2H2NOHnKrg04xNBLy4nEViKhwZZIRZ9+OgRvywmmbYSVlLCh
```

##### Chacha20-Poly1305

ChaCha20Poly1305 (RFC 8439) is an Authenticated Encryption with Associated Data
(AEAD) cipher amenable to fast, constant-time implementations in software, based
on the ChaCha20 stream cipher and Poly1305 universal hash function.

```sh
$ harp keygen chacha
chacha:dNph0d9Pj2_IMiRgBQrzExBH7899OKdyH3_T88WE1Zk=
$ echo -n "test" \
    | harp transform encrypt --key chacha:dNph0d9Pj2_IMiRgBQrzExBH7899OKdyH3_T88WE1Zk= \
    | base64
alhOtiM1GiyPqWb4thy5COzytwP0oT3b+LGiW7KZUnE=
```

##### XChacha20-Poly1305

XChaCha20Poly1305 is a modified version of ChaCha20Poly1305 created by Scott
Arciszewski that is hardened against nonce misuse.

```sh
$ harp keygen xchacha
xchacha:8ZgF0VI0lfHx8GwzDxp6fUqA1zAocPWDXRKLRMmgXUQ=
$ echo -n "test" \
    | harp transform encrypt --key xchacha:8ZgF0VI0lfHx8GwzDxp6fUqA1zAocPWDXRKLRMmgXUQ= \
    | base64
nH8U37FO0KdCKGRyzFz6TJZm9V4juTF7bOdcIu33j++qgK26On2HNkc3g+w=
```

#### Deterministic Authenticated Encryption (DAE)

> A deterministic encryption scheme (as opposed to a probabilistic encryption
> scheme) is a cryptosystem which always produces the same ciphertext for a
> given plaintext and key, even over separate executions of the encryption
> algorithm.
>
> Wikidpedia - https://en.wikipedia.org/wiki/Deterministic_encryption

Deterministic AEAD has the following properties:

* Secrecy: Nobody will be able to get any information about the encrypted
  plaintext, except the length and the equality of repeated plaintexts.
* Authenticity: Without the key it is impossible to change the plaintext
  underlying the ciphertext undetected.
* Symmetric: Encrypting the message and decrypting the ciphertext is done with
  the same key.
* Deterministic: A deterministic AEAD protects data almost as well as a normal
  AEAD. However, if you send the same message twice, an attacker can notice
  that the two messages are equal. If this is not desired, see AEAD.

`Harp` uses a pseudo-random-function (PRF) to derive the IV from the content.
This kind of encryption offers less guarantee that a probabilistic encryption
where the output is considered as random, and reveal same clear-text message
usages.
In order to achieve something which could be called `searchable-encryption`
where the algorithm is used as a reversible hash function, DAE could be used.

```go
NonceSize := 32

// Salt can be provided during the transformer creation, default to nil
salt := givenSalt | nil
// No additional data
info := nil
// Derived key length
dkLen := len(k)+NonceSize

// Stretch given key to get the required buffer size.
deriveBuf := hkdf(SHA256, key, salt, info, dkLen)

// Encryption key is equal to the given input key length
encryptionKey := derivedBuf[:len(key)]
// HMAC Key is always 32bytes
hmacKey := derivedBuf[len(key):]

// Compute nonce/iv based on the content to get deterministic IV
iv := HMACSHA256(message, hmacKey)
// Seal message with AEAD
sealed := AEAD(encryptionKey, iv, message)
```

##### How to use DAE?

In order to use DAE, you have to prefix `dae-` on supported AEAD encryption
algorithms.

```txt
dae-<aead>:<key>
```

You can provide an optional `salt` to diverge from standard encryption when
using the same key.

```txt
dae-<aead>:<key>:BASE64URL(<salt>)
```

##### DAE-AES-GCM

```sh
$ echo -n "test" \
    | harp transform encrypt --key dae-aes-gcm:yOF_27OF4aokpB_6WUCBrg== \
    | base64
SqLUCn86xNbqVcFMlSXnMRYn9qznXSsAD8B/qW7I/GU=
```

##### DAE-AES-SIV

```sh
$ echo -n "test" \
    | harp transform encrypt --key dae-aes-siv:glmpGwiMb3bop4XwZeYvfGPZSn_87uAnC75vp73tJFAvfP6mdLX5dgTYj0cgua_tUZ7itQ-MFwNF-ZCGzr0Fxw== \
    | base64
iAlNertgxaIwj8IOxlfRhyZKQknxPTopEPHBZ6jCsrJ54fzuaIIWq2vxLlaxzy9howI12A==
```

##### DAE-AES-PMAC-SIV

```sh
$ echo -n "test" \
    | harp transform encrypt --key aes-pmac-siv:5W5c43ZFVX37Y4p7tlksEyyuw6R_VpF68QoczfNUnpameV_63Kif0byRd-KFA-svBW5eXy2D_9h_S02xdWKEKA== \
    | base64
WNzc6rSYNu1IMlS9xvcfTlSsGmxZJI79+YnrGFADSfF0TLFcKfWsc8pAZHtsLuYiy/g/ug==
```

##### DAE-Chacha20-Poly1305

```sh
$ echo -n "test" \
    | harp transform encrypt --key dae-chacha:dNph0d9Pj2_IMiRgBQrzExBH7899OKdyH3_T88WE1Zk= \
    | base64
jb00tt+iUluLQEfxzz/+zcNvu7NeNyIcEehhY4zvy4o=
```

##### DAE-XChacha20-Poly1305

```sh
$ echo -n "test" \
    | harp transform encrypt --key dae-xchacha:8ZgF0VI0lfHx8GwzDxp6fUqA1zAocPWDXRKLRMmgXUQ= \
    | base64
m33d+reu9JgGuL6rlQDxoamN+nq/eDsPp8Ee4BOtXq1Leq6UvDI9pahYgjE=
```

### Encoded output

The following transformers return encoded output according to the selected
transformer specification.

#### Fernet

Fernet is basically AES128 in CBC mode with a SHA256 HMAC message authentication
code.

```sh
$ harp keygen fernet
8niskIbkop11u-_FDqSE8PEIVxv9C2bJvisc3YJwqDY=
$ echo -n "test" \
    | harp transform encrypt --key 8niskIbkop11u-_FDqSE8PEIVxv9C2bJvisc3YJwqDY=
gAAAAABh-BtD6UnmeFe3xXcw66KrdDG7bcu6kaIe-bHfWtbpLk_nVPxScS7ChZhznDjFU3st7DovGd2FZXJ-7Y0ZTLKHVOZL6w==
```

#### JWE

The JWE (JSON Web Encryption) specification standardizes the way to represent
an encrypted content in a JSON-based data structure.

* `jwe:a128kw:<base64>` to initialize a AES128 Key Wrapper with AES128 GCM Encryption transformer
* `jwe:a192kw:<base64>` to initialize a AES192 Key Wrapper with AES192 GCM Encryption transformer
* `jwe:a256kw:<base64>` to initialize a AES256 Key Wrapper with AES256 GCM Encryption transformer
* `jwe:pbes2-hs256-a128kw:<ascii>` to initialize a PBES2 key derivation function for AES128 key wrapping with AES128 GCM Encryption transformer
* `jwe:pbes2-hs384-a192kw:<ascii>` to initialize a PBES2 key derivation function for AES192 key wrapping with AES192 GCM Encryption transformer
* `jwe:pbes2-hs512-a256kw:<ascii>` to initialize a PBES2 key derivation function for AES256 key wrapping with AES256 GCM Encryption transformer


##### With a key

```sh
$ harp keygen aes-gcm --size 256
aes-gcm:GjyAYrHkFXQprLT3dQ6LCXairJKBUmLrSw86Hfetpac=
$ echo -n "test" \
    | harp transform encrypt --key jwe:a256kw:GjyAYrHkFXQprLT3dQ6LCXairJKBUmLrSw86Hfetpac=
eyJhbGciOiJBMjU2S1ciLCJlbmMiOiJBMjU2R0NNIn0.nU7WPDvptwVEKN9mRoO70xSvpQXLU6FkEML8B3NlvzDFD61rh4yo0A.2K6pjF_gdHfl8kXK.56fWkQ._TOG8llQ4stQVXdrgK2UNA
```

##### With a password/passphrase

```sh
$ harp passphrase
unblended-math-visibly-onscreen-request-expire-dragging-magenta
$ echo -n "test" \
    | harp transform encrypt --key jwe:pbes2-hs512-a256kw:unblended-math-visibly-onscreen-request-expire-dragging-magenta
eyJhbGciOiJQQkVTMi1IUzUxMitBMjU2S1ciLCJlbmMiOiJBMjU2R0NNIiwicDJjIjo1MDAwMDEsInAycyI6ImE0R1F5ZWw2akt5TTF0VU1QVWg3X3cifQ.GOjGwZHodjJicINuYxvWWKE-ENS0Cl2nFcGsA9qJvOGPTCSfoY6U-Q.hrqUQ4AWD5cSlbNc.283KnQ.rL_UFpRVDwF2WG2PjqFdfw
```

#### PASETO

> Paseto is everything you love about JOSE (JWT, JWE, JWS) without any of the
> many design deficits that plague the JOSE standards.
>
> https://github.com/paragonie/paseto

```sh
$ harp keygen paseto
paseto:HTh5iRgEEVDwpcyJEQ6pYx7S5YCGHGB729emBy2e-K4=
$ echo -n "test" \
    | harp transform encrypt --key paseto:HTh5iRgEEVDwpcyJEQ6pYx7S5YCGHGB729emBy2e-K4=
v4.local.RwujcJdC3XhWRaYHItFev5NlaC5u99iQR9njJILs2g6lcZKNecHRCvqMG81LQWbaVtfwI_qBgbB5lPSsrv9QojC_tc0
```

## Asymmetric encryption

### AGE

*Encrypt*

Key format : `age-recipients:<recipient-0>(:<recipient-n>)+`

```sh
$ echo -n "test" \
    | harp transform encrypt --key age-recipients:age1ce20pmz8z0ue97v7rz838v6pcpvzqan30lr40tjlzy40ez8eldrqf2zuxe
-----BEGIN AGE ENCRYPTED FILE-----
YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IFgyNTUxOSBpOWlwTm5JVTA5UEtFdHRI
ZU9oYzlyT0VrUEY4UGxtelM2ZjJBQUpmK2dJCnVBUWV2elVDM0RleTJZMTMrV0Fs
WmFMdVdCQmFDWG1uOHZ6blJEVWdaVDgKLS0tIDkyeTlnU0lHZWJJazQ5S1BYTU83
Q2lReW5sRm1mUUtVa3VuZCtlLzVZZXMKcl09YVsNLaCYMva21IHYtm0IvyItoyXS
UFJ6Cim7GbaB5Avu
-----END AGE ENCRYPTED FILE-----
```

*Decrypt*

Key format : `age-identity:<identity-secret-key>`

```sh
$ harp transform decrypt \
    --key age-identity:AGE-SECRET-KEY-1W8E69DQEVASNK68FX7C6QLD99KTG96RHWW0EZ3RD0L29AHV4S84QHUAP4C \
    --in encrypted.file
```

## Envelope encryption

TODO
