# Container sealing

> [Container sealing algorithm is inspired from KeyBase saltpack specification](https://saltpack.org/)

## Sealing Process

This algorithm implements multi recipient authenticated encryption with
`sign-then-encrypt` pattern.

* Preparation
  * Serialize `unsealed` container as protobuf as `container_content`
  * Generate a random 32 bytes array as `payload_key`
  * Generate ed25519 `ephemeral signing keypair`
  * Seal `ephemeral_signing_public_key` using a `secretbox`
    * the fixed 24 bytes nonce (`harp_container_psigk_box`)
    * with `payload_key`
  * Generate X25519 `ephemeral encryption keypair`

* Header
  * Set `ContentType` to `application/vnd.harp.v1.SealedContainer`
  * Set `EncryptionPublicKey` to `ephemeral encryption public key`
  * Set `ContainerBox`to `encrypted ephemeral signing public key`
  * For each given `recipient_public_key`
    * Derive recipient key
      * Prepare fixed 24 bytes nonce (`harp_derived_id_sboxkey0`)
      * Initialize 32 bytes zero filled array
      * Seal the zero filled array using fixed nonce, the `recipient_public_key` and the `ephemeral_encryption_private_key`
      * Get the last 32 bytes of the result and save it as `derived_recipient_key`
    * Derive recipient identifier
      * Prepare Blake2b-512 hash function with a fixed 41 bytes as key (`harp signcryption box key identifier`)
      * Hash the `derived_recipient_key`
      * Save first 32 bytes of result as `recipient_identifier`
    * Pack the `containerv1.Recipient` object
      * Set `Identifier` to `recipient_identifier`
      * Set `Key` to `recipient_key`
    * Add `Recipient` object to `recipients` list
  * Calculate `header_hash`
    * Serialize `containerv1.Header` object as protobuf byte array
    * Compute Blake2b-512 hash of serialized byte array
    * Save the hash result byte array as `header_hash`

* Content
  * Signature
    * Prepare `protected_content` by concatenating :
      * the ascii string `harp encrypted signature`
      * a zero byte (`0x00`)
      * the `header_hash` content
      * the Blake2b-512 hash result of the `container_content`
    * Sign using `ed25519` signature scheme
      * Sign `protected_content` with `ephemeral signing private key`
      * Save signature as `content_signature`
  * Encryption
    * Concatenante the `content_signature` and `container_content`
    * Seal the result with `payload_key` and the first 24 bytes of `header_hash` as nonce
    * Set `Raw` to encryption result

## Unsealing Process

* Requirements
  * A set of `identity_private_key`

* Header
  * Validate `ContentType` with `application/vnd.harp.v1.SealedContainer`
  * Retrieve `EncryptionPublicKey` and validate the required length (32 bytes)
  * For each `identity_private_key`
    * Derive recipient key
      * Prepare fixed 24 bytes nonce (`harp_derived_id_sboxkey0`)
      * Initialize 32 bytes zero filled array
      * Seal the zero filled array using fixed nonce, the `ephemral_encryption_public_key` and the `recipient_private_key`
      * Get the last 32 bytes of the result and save it as `derived_recipient_key`
    * Derive recipient identifier
      * Prepare Blake2b-512 hash function with a fixed 41 bytes as key (`harp signcryption box key identifier`)
      * Hash the `derived_recipient_key`
      * Save first 32 bytes of result as `recipient_identifier`
    * For each `Recipients` in `recipients` list
      * Compare `Identifier` with `recipient_identifier`
      * If not match continue until you find a match, unless error
      * If match, unseal the `recipient_key` from `Key` using the `derived_recipient_key`
      * Save the result as `payload_key`

* Content
  * Decryption
    * Retrieve `ephemeral_signing_public_key`
      * Open `ContainerBox` secretbox with
        * the fixed 24 bytes nonce (`harp_container_psigk_box`)
        * with `payload_key`
      * Save the result as `ephemeral_signing_public_key`
    * Calculate `header_hash`
      * Serialize `containerv1.Header` object as protobuf byte array
      * Compute Blake2b-512 hash of serialized byte array
      * Save the hash result byte array as `header_hash`
    * Open `Raw` box with
      * the first 24 bytes of `header_hash` as nonce
      * the `payload_key`
  * Signature
    * Prepare `protected_content` by concatenating :
      * the ascii string `harp encrypted signature`
      * a zero byte (`0x00`)
      * the `header_hash` content
      * the Blake2b-512 hash result of the `container_content`
    * Verify using `ed25519` signature scheme
      * Verify `protected_content` with `ephemeral_signing_public_key`
      * Save signature as `content_signature`
  * Unmarshall `payload` as `&containerv1.Container{}`

---

* [Previous topic](2-specifications.md)
* [Index](../)
* [Next topic](../3-secret-bundle/1-introduction.md)
