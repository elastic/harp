<!--
 Copyright 2022 Thibault NORMAND

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
-->

# Signature Transformers

## Overview

| Transformer | Format | Algorithms | Use Case |
|------------|--------|-----------|----------|
| `jws:` | Compact JWT | EdDSA, ES256/384/512, RS256/384/512, PS256/384/512 | API tokens, standard JWTs |
| `paseto:` | PASETO v4 | Ed25519 | Modern alternative to JWT |
| `raw:` | Binary | Ed25519, ECDSA, RSA | Detached signatures, custom formats |

## Key Generation

### CLI Method

```bash
# Generate Ed25519 key (recommended)
harp keygen jwk --algorithm EdDSA

# Generate with other algorithms
harp keygen jwk --algorithm ES256  # ECDSA P-256
harp keygen jwk --algorithm RS256  # RSA 2048
harp keygen jwk --algorithm PS256  # RSA-PSS
```

### Template Method

```
{{- $key := cryptoPair "ed25519" -}}
{{ $key.Private | toJwk }}  # Includes alg field automatically
```

**Both methods produce JWK with `alg` field** - ready for signing.

## Signing Operations

### Basic JWS Signing

Keys must be **base64url-encoded JWK** with `jws:` prefix:

```bash
# Sign a message
JWK="$(harp keygen jwk --algorithm EdDSA)"
KEY_B64=$(echo "$JWK" | harp transform encode --encoding base64url --in -)
echo -n "message" | harp transform sign --key "jws:$KEY_B64" --deterministic

# Output: eyJhbGciOiJFZERTQSJ9.dGVzdCBtZXNzYWdl.signature...
```

### Verification

```bash
# Verify signature
echo "eyJhbGciOiJFZERTQSJ9.dGVzdCBtZXNzYWdl.signature..." | \
  harp transform verify --key "jws:$PUBLIC_KEY_B64"

# Output: test message
```

## Algorithm Details

### Supported Algorithms Matrix

| Algorithm | Key Type | Signature Size | Speed | Security | FIPS Compatible |
|-----------|----------|----------------|-------|----------|----------------|
| **EdDSA** | Ed25519 | 64 bytes | ⚡⚡⚡ | High | No |
| **ES256** | ECDSA P-256 | ~72 bytes | ⚡⚡ | Medium-High | Yes |
| **ES384** | ECDSA P-384 | ~104 bytes | ⚡⚡ | High | Yes |
| **ES512** | ECDSA P-521 | ~139 bytes | ⚡⚡ | Very High | Yes |
| **RS256** | RSA-2048 | 256 bytes | ⚡ | Medium | Yes |
| **RS384** | RSA-2048 | 256 bytes | ⚡ | Medium-High | Yes |
| **RS512** | RSA-2048 | 256 bytes | ⚡ | High | Yes |
| **PS256** | RSA-PSS | 256 bytes | ⚡ | Medium-High | Yes |
| **PS384** | RSA-PSS | 256 bytes | ⚡ | High | Yes |
| **PS512** | RSA-PSS | 256 bytes | ⚡ | Very High | Yes |

## Complete Examples

See [Signing Workflows](4-signing-workflows.md) for complete end-to-end examples including:
- Vault storage integration
- Deterministic vs non-deterministic signing
- All algorithm variants

* [Previous topic](2-encryption.md)
* [Index](../)
* [Next topic](4-signing-workflows.md)
