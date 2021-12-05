# Specifications

## Implementation

### File format

The secret container is using protobuf as on-the-wire format.

```txt
magic (4 bytes) || version  (2 bytes) || payload (til EOF)
```

> `||` means concatenation, these characters are not used during serialization.

* The `magic` is fixed to `0x53CB3701` to recognize a secret container.
* The `version` is fixed to`0x0002' to recognize serialization format version.
* The `payload` is a protobuf serialized `harp.container.v1.Container` object.

### Secret Container

Container header definition :

```protobuf
// Header describes container headers.
message Header {
  // Content encoding describes the content encoding used for raw.
  // Unspecified means no encoding.
  string content_encoding = 1;
  // Content type is the serialization method used to serialize 'raw'.
  // Unspecified means "application/vnd.harp.protobuf".
  string content_type = 2;
  // Ephemeral public key used for encryption.
  bytes encryption_public_key = 3;
  // Container box contains public signing key encrypted with payload key.
  bytes container_box = 4;
  // Recipient list for identity bound secret container.
  repeated Recipient recipients = 6;
  // Seal strategy
  uint32 seal_version = 7;
}
```

* The `content_encoding` is a string which defines encoding used to store `raw`
  content. (i.e. `gzip`, `compress`)
* The `content_type` is the serialization method used to serialize `raw`.
* The `encryption_public_key` is the ephemeral x25519 public key used for
  container encryption.
* The `container_box` is the signature public key encrypted with the payload key.
* The `recipients` is a NaCL `box` that contains the x25519 private used for
  encryption protected using the passphrase during `sealing` process.
* The `seal_version` indicates the algorithm used to seal the container.

Recipient definition :

```protobuf
// Recipient describes container recipient informations.
message Recipient {
  // Recipient identifier
  bytes identifier = 1;
  // Encrypted copy of the payload key for recipient.
  bytes key = 2;
}

```

* The `identitifer` is the anonymized recipient identifier.
* The `key` is a NaCL `box` that contains the payload key encrypted using the
  recipient public key, so that only the owner of the private key can open the
  box.

Container definition :

```protobuf
// Container describes the container attributes.
message Container {
  // Container headers.
  Header headers = 1;
  // Raw hold the complete serialized object in protobuf.
  bytes raw = 2;
}
```

* The `headers` are the secret container headers object.
* The `raw` are serialized data.

## Container Modes

### Sealed

```go
struct Container {
    Headers: &containerv1.Header{
        ContentType: "application/vnd.harp.v1.SealedContainer",
        ContentEncoding: "gzip",
        EncryptionPublicKey: [32]byte{ ... }, // X25519 public encryption key
        ContainerBox: [64]byte{ ... }, // NaCL box containing signing public key encrypted with payload key.
        Recipients: []&containerv1.Recipient{
            {
                Identifier: []byte{ ... }, // Anonymized identifier
                Key: []byte{ ... }, // Nacl secretbox with payload key encrypted
            }
        }
    },
    Raw: []byte{ ... } // Data with attached ed25519 signature as prefix encrypted using payload key.
}
```

#### Container identities

Container identities are `X25519` key pairs where public key is used as a
`recipient` and private key as container key to unseal the secret container.

> No container identity revocation list is implemented, if a key is compromised
> all containers sealed with this key are compromised, and need to be rotated.

##### Ephemeral Container Key

For immutability principle, the sealing process generates a new Container Identity
at each execution. It means that all the container consumers must know the new
container to be able to unseal it.

In order to `seal` a secret container, you can use the following commands :

Generate a passphrase first. This passphrase will be used to encode a recovery
content in case of container key loss.

```sh
harp passphrase > passphrase.txt
harp container identity --passphrase $(cat pass.txt) --description "Recovery" --out recovery.json
```

Seal the container using the generated passphrase for recovery :

```sh
$ harp container seal --in unsealed.container \
    --identity-file recovery.json
    --out sealed.container
Container Key: .....
```

##### Deterministic Container Key

Seal the container using a deterministic container key derived from a master key.
This will prevent modification of container consumers after each container seals.

Generate a master key :

> Keep this key as a high sensitive secret.

```sh
harp keygen master-key > master.key
```

Seal the secret container using deterministic container key derivation (DCKD) :

```sh
$ harp container seal --in unsealed.container \
    --identity-file recovery.json \
    --dckd-master-key $(cat master.key) \
    --dckd-target "essp:ms-46:2020-08-31" \
    --out sealed.container
Container key : ....
```

* The `dckd-master-key` flag defines the root key used for derivation.
* The `dckd-target` flag defines an arbitrary string acting as a salt for Key
  Derivation Function.

### Unsealed

```go
struct Container {
    headers: &containerv1.Header{
        ContentType: "application/vnd.harp.v1.Bundle",
        ContentEncoding: "gzip",
    },
    Raw: []byte{ ... } // Protobuf marshalled bundlev1.Bundle object
}
```

#### Unseal with Container Key

```sh
$ harp container unseal --in sealed.container \
    --out unsealed.container \
    --key $(cat container.key)
```

#### Container Key Recovery

If you have lost the container key, you can recover it by using the passphrase used
during the `sealing` process.

```sh
$ harp container recover --in recovery.json --passphrase $(cat pass.txt)
Container key: ....
```

Then use the recovered Container Key as usual.

---

* [Previous topic](1-introduction.md)
* [Index](../)
* [Next topic](3-seal.md)
