// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package crypto

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	_ "golang.org/x/crypto/blake2b"
)

func TestAlgorithmForECDSACurve(t *testing.T) {
	tests := []struct {
		name  string
		curve elliptic.Curve
		want  string
	}{
		{
			name:  "P-256",
			curve: elliptic.P256(),
			want:  "ES256",
		},
		{
			name:  "P-384",
			curve: elliptic.P384(),
			want:  "ES384",
		},
		{
			name:  "P-521",
			curve: elliptic.P521(),
			want:  "ES512",
		},
		{
			name:  "unsupported curve",
			curve: nil,
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := algorithmForECDSACurve(tt.curve)
			if got != tt.want {
				t.Errorf("algorithmForECDSACurve() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlgorithmForKey(t *testing.T) {
	// Generate test keys
	rsaPriv, rsaPub, err := generateKeyPair("rsa")
	if err != nil {
		t.Fatalf("unable to generate rsa key: %v", err)
	}

	ecPriv, ecPub, err := generateKeyPair("ec")
	if err != nil {
		t.Fatalf("unable to generate ec key: %v", err)
	}

	edPriv, edPub, err := generateKeyPair("ssh")
	if err != nil {
		t.Fatalf("unable to generate ed25519 key: %v", err)
	}

	tests := []struct {
		name string
		key  interface{}
		want string
	}{
		{
			name: "RSA private key",
			key:  rsaPriv,
			want: "RS256",
		},
		{
			name: "RSA public key",
			key:  rsaPub,
			want: "RS256",
		},
		{
			name: "ECDSA private key (P-256)",
			key:  ecPriv,
			want: "ES256",
		},
		{
			name: "ECDSA public key (P-256)",
			key:  ecPub,
			want: "ES256",
		},
		{
			name: "Ed25519 private key",
			key:  edPriv,
			want: "EdDSA",
		},
		{
			name: "Ed25519 public key",
			key:  edPub,
			want: "EdDSA",
		},
		{
			name: "unsupported key type",
			key:  "invalid",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := algorithmForKey(tt.key)
			if got != tt.want {
				t.Errorf("algorithmForKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToJWK(t *testing.T) {
	priv, pub, err := generateKeyPair("rsa")
	if err != nil {
		t.Error("unable to generate rsa key")
		return
	}

	tests := []struct {
		name    string
		args    interface{}
		want    string
		wantErr bool
	}{
		{
			name:    "nil",
			args:    nil,
			wantErr: true,
		},
		{
			name:    "private",
			args:    priv,
			wantErr: false,
		},
		{
			name:    "public",
			args:    pub,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ToJWK(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToJWK() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestToJWK_AlgorithmField(t *testing.T) {
	tests := []struct {
		name        string
		keyType     string
		wantAlg     string
		keyModifier func(interface{}, interface{}) interface{} // Optional: modify generated key
	}{
		{
			name:    "RSA private key includes alg=RS256",
			keyType: "rsa",
			wantAlg: "RS256",
			keyModifier: func(priv, pub interface{}) interface{} {
				return priv
			},
		},
		{
			name:    "RSA public key includes alg=RS256",
			keyType: "rsa",
			wantAlg: "RS256",
			keyModifier: func(priv, pub interface{}) interface{} {
				return pub
			},
		},
		{
			name:    "ECDSA P-256 private key includes alg=ES256",
			keyType: "ec",
			wantAlg: "ES256",
			keyModifier: func(priv, pub interface{}) interface{} {
				return priv
			},
		},
		{
			name:    "ECDSA P-256 public key includes alg=ES256",
			keyType: "ec",
			wantAlg: "ES256",
			keyModifier: func(priv, pub interface{}) interface{} {
				return pub
			},
		},
		{
			name:    "Ed25519 private key includes alg=EdDSA",
			keyType: "ssh",
			wantAlg: "EdDSA",
			keyModifier: func(priv, pub interface{}) interface{} {
				return priv
			},
		},
		{
			name:    "Ed25519 public key includes alg=EdDSA",
			keyType: "ssh",
			wantAlg: "EdDSA",
			keyModifier: func(priv, pub interface{}) interface{} {
				return pub
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate key pair
			pub, priv, err := generateKeyPair(tt.keyType)
			if err != nil {
				t.Fatalf("unable to generate %s key: %v", tt.keyType, err)
			}

			// Select the key to test
			var testKey interface{}
			if tt.keyModifier != nil {
				testKey = tt.keyModifier(priv, pub)
			} else {
				testKey = priv
			}

			// Convert to JWK
			jwkJSON, err := ToJWK(testKey)
			if err != nil {
				t.Fatalf("ToJWK() failed: %v", err)
			}

			// Parse JWK JSON to validate alg field
			var jwk map[string]interface{}
			if err := json.Unmarshal([]byte(jwkJSON), &jwk); err != nil {
				t.Fatalf("Failed to parse JWK JSON: %v", err)
			}

			// Verify alg field exists
			alg, ok := jwk["alg"]
			if !ok {
				t.Errorf("JWK missing 'alg' field. Got JWK: %s", jwkJSON)
				return
			}

			// Verify alg field has correct value
			algStr, ok := alg.(string)
			if !ok {
				t.Errorf("'alg' field is not a string. Got type: %T, value: %v", alg, alg)
				return
			}

			if algStr != tt.wantAlg {
				t.Errorf("ToJWK() alg = %v, want %v", algStr, tt.wantAlg)
			}

			// Additional validation: verify expected JWK fields
			if _, ok := jwk["kty"]; !ok {
				t.Errorf("JWK missing 'kty' field")
			}
			if _, ok := jwk["kid"]; !ok {
				t.Errorf("JWK missing 'kid' field")
			}
		})
	}
}

func TestToJWK_ECDSACurveVariants(t *testing.T) {
	tests := []struct {
		name    string
		curve   elliptic.Curve
		wantAlg string
	}{
		{
			name:    "P-256 curve produces ES256",
			curve:   elliptic.P256(),
			wantAlg: "ES256",
		},
		{
			name:    "P-384 curve produces ES384",
			curve:   elliptic.P384(),
			wantAlg: "ES384",
		},
		{
			name:    "P-521 curve produces ES512",
			curve:   elliptic.P521(),
			wantAlg: "ES512",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate ECDSA key with specific curve using crypto/rand
			privKey, err := ecdsa.GenerateKey(tt.curve, cryptorand.Reader)
			if err != nil {
				t.Fatalf("Failed to generate ECDSA key: %v", err)
			}

			// Test private key
			jwkJSON, err := ToJWK(privKey)
			if err != nil {
				t.Fatalf("ToJWK() failed for private key: %v", err)
			}

			var jwk map[string]interface{}
			if err := json.Unmarshal([]byte(jwkJSON), &jwk); err != nil {
				t.Fatalf("Failed to parse JWK JSON: %v", err)
			}

			if alg, ok := jwk["alg"].(string); !ok || alg != tt.wantAlg {
				t.Errorf("Private key: ToJWK() alg = %v, want %v", jwk["alg"], tt.wantAlg)
			}

			// Test public key
			jwkJSON, err = ToJWK(&privKey.PublicKey)
			if err != nil {
				t.Fatalf("ToJWK() failed for public key: %v", err)
			}

			if err := json.Unmarshal([]byte(jwkJSON), &jwk); err != nil {
				t.Fatalf("Failed to parse JWK JSON: %v", err)
			}

			if alg, ok := jwk["alg"].(string); !ok || alg != tt.wantAlg {
				t.Errorf("Public key: ToJWK() alg = %v, want %v", jwk["alg"], tt.wantAlg)
			}
		})
	}
}

func TestToJWK_RSAKeyTypes(t *testing.T) {
	// Generate RSA key
	privKey, err := rsa.GenerateKey(cryptorand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	tests := []struct {
		name string
		key  interface{}
	}{
		{
			name: "RSA private key",
			key:  privKey,
		},
		{
			name: "RSA public key",
			key:  &privKey.PublicKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwkJSON, err := ToJWK(tt.key)
			if err != nil {
				t.Fatalf("ToJWK() failed: %v", err)
			}

			var jwk map[string]interface{}
			if err := json.Unmarshal([]byte(jwkJSON), &jwk); err != nil {
				t.Fatalf("Failed to parse JWK JSON: %v", err)
			}

			// Verify alg field
			if alg, ok := jwk["alg"].(string); !ok || alg != "RS256" {
				t.Errorf("ToJWK() alg = %v, want RS256", jwk["alg"])
			}

			// Verify kty field
			if kty, ok := jwk["kty"].(string); !ok || kty != "RSA" {
				t.Errorf("ToJWK() kty = %v, want RSA", jwk["kty"])
			}
		})
	}
}

func TestToJWK_Ed25519KeyTypes(t *testing.T) {
	// Generate Ed25519 key
	pubKey, privKey, err := ed25519.GenerateKey(cryptorand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate Ed25519 key: %v", err)
	}

	tests := []struct {
		name string
		key  interface{}
	}{
		{
			name: "Ed25519 private key",
			key:  privKey,
		},
		{
			name: "Ed25519 public key",
			key:  pubKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwkJSON, err := ToJWK(tt.key)
			if err != nil {
				t.Fatalf("ToJWK() failed: %v", err)
			}

			var jwk map[string]interface{}
			if err := json.Unmarshal([]byte(jwkJSON), &jwk); err != nil {
				t.Fatalf("Failed to parse JWK JSON: %v", err)
			}

			// Verify alg field
			if alg, ok := jwk["alg"].(string); !ok || alg != "EdDSA" {
				t.Errorf("ToJWK() alg = %v, want EdDSA", jwk["alg"])
			}

			// Verify kty field
			if kty, ok := jwk["kty"].(string); !ok || kty != "OKP" {
				t.Errorf("ToJWK() kty = %v, want OKP", jwk["kty"])
			}

			// Verify crv field
			if crv, ok := jwk["crv"].(string); !ok || crv != "Ed25519" {
				t.Errorf("ToJWK() crv = %v, want Ed25519", jwk["crv"])
			}
		})
	}
}

func TestToPEM(t *testing.T) {
	rsaPriv, rsaPub, err := generateKeyPair("rsa")
	if err != nil {
		t.Error("unable to generate rsa key")
		return
	}

	ecPriv, ecPub, err := generateKeyPair("ec")
	if err != nil {
		t.Error("unable to generate ec key")
		return
	}

	edPriv, edPub, err := generateKeyPair("ssh")
	if err != nil {
		t.Error("unable to generate ssh key")
		return
	}

	tests := []struct {
		name    string
		args    interface{}
		want    string
		wantErr bool
	}{
		{
			name:    "nil",
			args:    nil,
			wantErr: true,
		},
		{
			name:    "RSA private",
			args:    rsaPriv,
			wantErr: false,
		},
		{
			name:    "RSA public",
			args:    rsaPub,
			wantErr: false,
		},
		{
			name:    "EC private",
			args:    ecPriv,
			wantErr: false,
		},
		{
			name:    "EC public",
			args:    ecPub,
			wantErr: false,
		},
		{
			name:    "SSH private",
			args:    edPriv,
			wantErr: false,
		},
		{
			name:    "SSH public",
			args:    edPub,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ToPEM(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToPEM() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestEncryptPEM(t *testing.T) {
	_, rsaPriv, err := generateKeyPair("rsa")
	if err != nil {
		t.Error("unable to generate rsa key")
		return
	}

	rsaPrivPem, err := ToPEM(rsaPriv)
	if err != nil {
		t.Error("unable to generate rsa key PEM")
		return
	}

	type args struct {
		pemData    string
		passphrase string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "nil",
			args:    args{},
			wantErr: true,
		},
		{
			name: "nil pem data",
			args: args{
				pemData:    "",
				passphrase: "foo",
			},
			wantErr: true,
		},
		{
			name: "empty passphrase",
			args: args{
				pemData:    rsaPrivPem,
				passphrase: "",
			},
			wantErr: true,
		},
		{
			name: "passphrase too short",
			args: args{
				pemData:    rsaPrivPem,
				passphrase: "foo",
			},
			wantErr: true,
		},
		{
			name: "valid",
			args: args{
				pemData:    rsaPrivPem,
				passphrase: "clash-cement-plywood-repeater-shrubbery-landscape-aghast-sulfur",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := EncryptPEM(tt.args.pemData, tt.args.passphrase)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncryptPEM() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestToSSH(t *testing.T) {
	rsaPub, rsaPriv, err := generateKeyPair("rsa")
	if err != nil {
		t.Error("unable to generate rsa key")
		return
	}

	ecPub, ecPriv, err := generateKeyPair("ec")
	if err != nil {
		t.Error("unable to generate ec key")
		return
	}

	edPub, edPriv, err := generateKeyPair("ssh")
	if err != nil {
		t.Error("unable to generate ssh key")
		return
	}

	tests := []struct {
		name    string
		args    interface{}
		want    string
		wantErr bool
	}{
		{
			name:    "nil",
			args:    nil,
			wantErr: true,
		},
		{
			name:    "RSA private",
			args:    rsaPriv,
			wantErr: false,
		},
		{
			name:    "RSA public",
			args:    rsaPub,
			wantErr: false,
		},
		{
			name:    "EC private",
			args:    ecPriv,
			wantErr: false,
		},
		{
			name:    "EC public",
			args:    ecPub,
			wantErr: false,
		},
		{
			name:    "SSH private",
			args:    edPriv,
			wantErr: false,
		},
		{
			name:    "SSH public",
			args:    edPub,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ToSSH(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToSSH() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestFromJWK(t *testing.T) {
	type args struct {
		jwk string
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name:    "blank",
			wantErr: true,
		},
		{
			name: "valid - private",
			args: args{
				jwk: `{	"kty": "EC", "d": "KtNle6xh0XBGhJbJEzP-5TiWdB6_dVkoWeWeo-VUVUI", "crv": "P-256", "x": "eoZzawRZk9sL9pkNYIKJJU34FyckdDAQg7LM2z0wez4", "y": "3Z6Z3vv1QQmQ3S5_4aeFnqrENhOBmBreXGYsbbLTLh8"	}`,
			},
			wantErr: false,
		},
		{
			name: "valid - public",
			args: args{
				jwk: `{	"kty": "EC", "crv": "P-256", "x": "eoZzawRZk9sL9pkNYIKJJU34FyckdDAQg7LM2z0wez4", "y": "3Z6Z3vv1QQmQ3S5_4aeFnqrENhOBmBreXGYsbbLTLh8"	}`,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FromJWK(tt.args.jwk)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromJWK() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestEncryptJWE(t *testing.T) {
	type args struct {
		key     string
		payload interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "blank",
			wantErr: false,
		},
		{
			name: "claims",
			args: args{
				key: "test",
				payload: map[string]interface{}{
					"sub": "test",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := EncryptJWE(tt.args.key, tt.args.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncryptJWE() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_EncryptDecryptJWE(t *testing.T) {
	claims := map[string]interface{}{
		"sub": "test",
	}

	jwe, err := EncryptJWE("test", claims)
	if err != nil {
		t.Fatalf("unable to encrypt claims: %v", err)
	}

	got, err := DecryptJWE("test", jwe)
	if err != nil {
		t.Fatalf("unable to decrypt claims: %v", err)
	}

	if report := cmp.Diff(claims, got); report != "" {
		t.Errorf("%s", report)
	}
}

func TestToJWS(t *testing.T) {

	_, ecPriv, err := generateKeyPair("ec")
	if err != nil {
		t.Error("unable to generate ec key")
		return
	}

	type args struct {
		payload interface{}
		privkey interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "blank",
			args: args{
				privkey: ecPriv,
			},
			wantErr: false,
		},
		{
			name: "claims",
			args: args{
				privkey: ecPriv,
				payload: map[string]interface{}{
					"sub": "test",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ToJWS(tt.args.payload, tt.args.privkey)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToJWS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
