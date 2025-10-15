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

# Signing Workflows

## Workflow 1: Bundle Template → Vault Storage → CLI Signing

### Step 1: Create Bundle Template

Create a bundle template file `service-keys.yaml`:

```yaml
apiVersion: harp.elastic.co/v1
kind: BundleTemplate
meta:
  name: "signing-keys"
  owner: test@elastic.co
  description: "signing-keys example"
spec:
  selector:
    platform: "examplePlatform"
    quality: "dev"
    product: "service-signing-keys"
    version: "v1.0.0"
  namespaces:
    application:
      - name: "signing-keys"
        description: "signing keys for services"
        secrets:
        - suffix: "crypto/signing-key"
          description: "example signing key"
          template: |
            {{- $key := cryptoPair "ed25519" -}}
            {{- $jwk := $key.Private | toJwk | fromJson -}}
            {
              "private_key": {{ $key.Private | toJwk | toJson }},
              "public_key": {{ $key.Public | toJwk | toJson }},
              "algorithm": {{ $jwk.alg | toJson }},
              "service": "service-signing-keys",
              "rotation_date": "{{ now | date "2006-01-02" }}",
              "expires_at": "{{ now | dateModify "8760h" | date "2006-01-02" }}"
            }
```

### Step 2: Generate and Store in Vault

```bash
# Render bundle template and push directly to Vault
harp from template --in service-keys.yaml | \
  harp to vault

# Or save bundle first (recommended for audit trail)
harp from template --in service-keys.yaml --out service-keys.bundle
harp to vault --in service-keys.bundle

# View what was generated (optional)
harp bundle dump --in service-keys.bundle --data-only | jq .
```

### Step 3: Retrieve and Sign

```bash
# Pull keys from Vault for specific service
PRIVATE_KEY=$(harp from vault --path "app/dev/examplePlatform/service-signing-keys/v1.0.0/signing-keys/crypto/signing-key" \
| harp bundle read --path "app/dev/examplePlatform/service-signing-keys/v1.0.0/signing-keys/crypto/signing-key" --field private_key \
| jq .)

# Encode for signing
PRIVATE_B64=$(echo "$PRIVATE_KEY" | harp transform encode --encoding base64url --in -)

# Sign a message
echo -n "API request payload" | harp transform sign \
  --key "jws:$PRIVATE_B64" \
  --deterministic

# Output: eyJhbGciOiJFZERTQSIsImtp...
```

### Step 4: Verify Signature

```bash
# Extract public key using piped workflow
PUBLIC_KEY=$(harp from vault --path "app/dev/examplePlatform/service-signing-keys/v1.0.0/signing-keys/crypto/signing-key" \
| harp bundle read --path "app/dev/examplePlatform/service-signing-keys/v1.0.0/signing-keys/crypto/signing-key" --field public_key \
| jq .)

# Encode for verification
PUBLIC_B64=$(echo "$PUBLIC_KEY" | harp transform encode --encoding base64url --in -)

# Verify signature
# echo "eyJhbGciOiJFZERTQSIsImtp.. | harp transform verify --key "jws:$PUBLIC_B64"
echo "$OUTPUT_FROM_STEP_3" | harp transform verify \
  --key "jws:$PUBLIC_B64"

# Output: API request payload
```

### Step 5: Key Rotation (Bonus)

```bash
# Generate new keys with updated template
harp from template --in service-keys.yaml --out service-keys-new.bundle

# Compare old and new keys
harp bundle diff \
  --old service-keys.bundle \
  --new service-keys-new.bundle

# Update Vault with new keys
harp to vault --in service-keys-new.bundle

# Archive old bundle for rollback capability
cp service-keys.bundle service-keys-backup-$(date +%Y%m%d).bundle
```

---

## Workflow 2: All Algorithms in One Template

```ruby
{{- $keys := dict -}}
{{- $algMap := dict "ed25519" "EdDSA" "rsa" "RS256" "ec:p256" "ES256" "ec:p384" "ES384" -}}
{{- range $type, $alg := $algMap -}}
  {{- $key := cryptoPair $type -}}
  {{- $jwk := $key.Private | toJwk | fromJson -}}
  {{- $_ := set $keys $alg (dict
      "kty" $jwk.kty
      "kid" $jwk.kid
      "alg" $jwk.alg
      "private_b64" ($key.Private | toJwk | b64enc)
      "public_b64" ($key.Public | toJwk | b64enc)
  ) -}}
{{- end -}}
{{ $keys | toJson }}
```

**Output structure:**

```json
{
  "EdDSA": {
    "kty": "OKP",
    "kid": "abc123...",
    "alg": "EdDSA",
    "private_b64": "eyJrdHkiOiJPS1Ai...",
    "public_b64": "eyJrdHkiOiJPS1Ai..."
  },
  "RS256": { ... },
  "ES256": { ... },
  "ES384": { ... }
}
```

---

## Workflow 3: Deterministic vs Non-Deterministic Signing

### Deterministic (Reproducible)

```bash
# Same input + key = same signature (RFC 6979)
echo -n "test" | harp transform sign --key "jws:$KEY" --deterministic
# Always produces: eyJhbGciOiJFZERTQSJ9.dGVzdA.gCushVU...

# Use cases:
# - Testing and validation
# - Signature caching
# - Reproducible builds
```

### Non-Deterministic (Random Nonce)

```bash
# Each signature is unique due to random nonce
echo -n "test" | harp transform sign --key "jws:$KEY"
# Produces: eyJhbGciOiJFZERTQSIsIm5vbmNlIjoiUnlmQiJ9.dGV...

# Use cases:
# - Protection against replay attacks
# - Time-sensitive operations
# - Production security
```

* [Previous topic](3-signature.md)
* [Index](../)
