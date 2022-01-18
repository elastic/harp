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
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"

	"github.com/pkg/errors"
	"go.step.sm/crypto/pemutil"

	// Import Blake2b
	_ "golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/ssh"
	jose "gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"

	"github.com/elastic/harp/build/fips"
	"github.com/elastic/harp/pkg/sdk/security/crypto/bech32"
	"github.com/elastic/harp/pkg/sdk/types"
)

// ToJWK encodes given key using JWK.
func ToJWK(key interface{}) (string, error) {
	// Check key
	if types.IsNil(key) {
		return "", fmt.Errorf("unable to encode nil key")
	}

	// Wrap key
	keyWrapper := jose.JSONWebKey{Key: key, KeyID: ""}

	// Don't process Ed25519 keys
	if fips.Enabled() {
		switch key.(type) {
		case ed25519.PrivateKey, ed25519.PublicKey:
			return "", errors.New("ed25519 key processing is disabled in FIPS Mode")
		}
	}

	// Generate thumbprint
	thumb, err := keyWrapper.Thumbprint(crypto.SHA512_256)
	if err != nil {
		return "", err
	}

	// Assign thumbprint
	keyWrapper.KeyID = base64.URLEncoding.EncodeToString(thumb)

	// Marshal private as JSON
	payload, err := keyWrapper.MarshalJSON()
	if err != nil {
		return "", err
	}

	// No error
	return string(payload), nil
}

// FromJWK parses a JWK and return wrapped keys.
func FromJWK(jwk string) (interface{}, error) {
	var k jose.JSONWebKey

	// Decode JWK
	if err := json.Unmarshal([]byte(jwk), &k); err != nil {
		return nil, fmt.Errorf("unable to decode JWK: %w", err)
	}

	// Don't process Ed25519 keys
	if fips.Enabled() {
		switch k.Key.(type) {
		case ed25519.PrivateKey, ed25519.PublicKey:
			return "", errors.New("ed25519 key processing is disabled in FIPS Mode")
		}
	}

	if k.IsPublic() {
		// No error
		return struct {
			Public interface{}
		}{
			Public: k.Key,
		}, nil
	}

	// No error
	return struct {
		Private interface{}
		Public  interface{}
	}{
		Private: k.Key,
		Public:  k.Public().Key,
	}, nil
}

// ToPEM encodes the given key using PEM.
func ToPEM(key interface{}) (string, error) {
	// Check key
	if types.IsNil(key) {
		return "", fmt.Errorf("unable to encode nil key")
	}

	// Don't process Ed25519 keys
	if fips.Enabled() {
		switch key.(type) {
		case ed25519.PrivateKey, ed25519.PublicKey:
			return "", errors.New("ed25519 key processing is disabled in FIPS Mode")
		}
	}

	// Delegate to smallstep library
	pemBlock, err := pemutil.Serialize(key, pemutil.WithPKCS8(true))
	if err != nil {
		return "", fmt.Errorf("unable to serialize input as PEM: %w", err)
	}

	return string(pem.EncodeToMemory(pemBlock)), nil
}

// KeyToBytes encodes the given crypto key as a byte array.
func KeyToBytes(key interface{}) ([]byte, error) {
	// Check key
	if types.IsNil(key) {
		return nil, fmt.Errorf("unable to encode nil key")
	}

	var (
		out []byte
		err error
	)
	switch k := key.(type) {
	// Private keys ------------------------------------------------------------
	case *rsa.PrivateKey, *ecdsa.PrivateKey:
		out, err = x509.MarshalPKCS8PrivateKey(k)
		if err != nil {
			return nil, err
		}
	case ed25519.PrivateKey:
		if fips.Enabled() {
			return nil, errors.New("ed25519 private key processing is disabled in FIPS Mode")
		}
		out = []byte(k)
	// Public keys ------------------------------------------------------------
	case *rsa.PublicKey:
		out, err = x509.MarshalPKIXPublicKey(k)
		if err != nil {
			return nil, err
		}
	case *ecdsa.PublicKey:
		out = elliptic.MarshalCompressed(k.Curve, k.X, k.Y)
	case ed25519.PublicKey:
		if fips.Enabled() {
			return nil, errors.New("ed25519 private key processing is disabled in FIPS Mode")
		}
		out = []byte(k)
	default:
		return nil, fmt.Errorf("given key type is not supported")
	}

	return out, nil
}

// EncryptPEM returns an encrypted PEM block using the given passphrase.
func EncryptPEM(pemData, passphrase string) (string, error) {
	// Check passphrase
	if len(passphrase) < 32 {
		return "", fmt.Errorf("passphrase must contains more than 32 characters, usage of a diceware passphrase is recommended")
	}

	// Decode PEM
	out, err := pemutil.Parse([]byte(pemData))
	if err != nil {
		return "", fmt.Errorf("unable to parse input PEM data: %w", err)
	}

	// Encrypt PEM
	encryptedBlock, err := pemutil.Serialize(out, pemutil.WithPKCS8(true), pemutil.WithPassword([]byte(passphrase)))
	if err != nil {
		return "", fmt.Errorf("unable to export encrypted PEM: %w", err)
	}

	// Build output
	outPem := pem.EncodeToMemory(encryptedBlock)

	// No error
	return string(outPem), nil
}

// ToSSH encodes the given key as SSH key.
func ToSSH(key interface{}) (string, error) {
	var result []byte

	// Check key
	if types.IsNil(key) {
		return "", fmt.Errorf("unable to encode nil key")
	}

	switch k := key.(type) {
	// Public keys ------------------------------------------------------------
	case *rsa.PublicKey, *ecdsa.PublicKey, ed25519.PublicKey:
		if _, ok := k.(ed25519.PublicKey); ok && fips.Enabled() {
			return "", errors.New("ed25519 public key processing is disabled in FIPS Mode")
		}
		pubKey, err := ssh.NewPublicKey(k)
		if err != nil {
			return "", fmt.Errorf("unable to convert key as ssh public key: %w", err)
		}
		result = ssh.MarshalAuthorizedKey(pubKey)
	// Private keys --------------------------------------------------------
	default:
		if _, ok := k.(ed25519.PrivateKey); ok && fips.Enabled() {
			return "", errors.New("ed25519 private key processing is disabled in FIPS Mode")
		}
		pemBlock, err := pemutil.Serialize(key, pemutil.WithOpenSSH(true))
		if err != nil {
			return "", fmt.Errorf("unable to encode SSH key: %w", err)
		}
		result = pem.EncodeToMemory(pemBlock)
	}

	// No error
	return string(result), nil
}

// EncryptJWE encrypts input as JWE token.
func EncryptJWE(key string, payload interface{}) (string, error) {
	// Prepare encrypter
	encrypter, err := jose.NewEncrypter(jose.A128GCM, jose.Recipient{Algorithm: jose.PBES2_HS256_A128KW, Key: key}, nil)
	if err != nil {
		return "", fmt.Errorf("unable to initialize JWE encrypter: %w", err)
	}

	// Marshal payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("unable to marshal content: %w", err)
	}

	// Encrypt
	object, err := encrypter.Encrypt(payloadBytes)
	if err != nil {
		return "", fmt.Errorf("unable to encrypt payload: %w", err)
	}

	// Return JWE
	return object.CompactSerialize()
}

// DecryptJWE decrypt a JWE token.
func DecryptJWE(key, token string) (interface{}, error) {
	// Parse JWE token
	object, err := jose.ParseEncrypted(token)
	if err != nil {
		return "", fmt.Errorf("unable to parse JWE assertion: %w", err)
	}

	// Decrypt token using given key.
	payloadBytes, err := object.Decrypt(key)
	if err != nil {
		return "", fmt.Errorf("unable to decrypt JWE: %w", err)
	}

	// Decode payload
	var data interface{}
	if err := json.Unmarshal(payloadBytes, &data); err != nil {
		return "", fmt.Errorf("unable to decode payload: %w", err)
	}

	// No error
	return data, nil
}

// ToJWS returns a JWT token.
func ToJWS(payload, privkey interface{}) (string, error) {
	var alg jose.SignatureAlgorithm

	// Select appropriate algorithm
	switch k := privkey.(type) {
	case *rsa.PrivateKey:
		alg = jose.RS256
	case *ecdsa.PrivateKey:
		switch k.Curve {
		case elliptic.P256():
			alg = jose.ES256
		case elliptic.P384():
			alg = jose.ES384
		case elliptic.P521():
			alg = jose.ES512
		}
	case ed25519.PrivateKey:
		if fips.Enabled() {
			return "", errors.New("signature with Ed25519 key is disabled in FIPS Mode")
		}
		alg = jose.EdDSA
	default:
		return "", fmt.Errorf("this private key type is not supported '%T'", privkey)
	}

	// Create a signer
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: alg, Key: privkey}, nil)
	if err != nil {
		return "", fmt.Errorf("unable to initialize signer: %w", err)
	}

	// Encode payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("unable to marshal payload as json: %w", err)
	}

	// Sign the token
	object, err := signer.Sign(payloadBytes)
	if err != nil {
		return "", fmt.Errorf("unable to sign payload: %w", err)
	}

	// Serialize final token
	serialize, err := object.CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("unable to generate final token: %w", err)
	}

	// No error
	return serialize, nil
}

// ParseJWT unpack a JWT without signature verification.
func ParseJWT(token string) (interface{}, error) {
	// Parse token
	t, err := jwt.ParseSigned(token)
	if err != nil {
		return nil, fmt.Errorf("unable to parse input token: %w", err)
	}

	// Extract claims without verification
	var claims map[string]interface{}
	if err := t.UnsafeClaimsWithoutVerification(&claims); err != nil {
		return nil, fmt.Errorf("unable to extract claims from token: %w", err)
	}

	return struct {
		Headers []jose.Header
		Claims  map[string]interface{}
	}{
		Headers: t.Headers,
		Claims:  claims,
	}, nil
}

func VerifyJWT(token string, key interface{}) (interface{}, error) {
	// Parse token
	t, err := jwt.ParseSigned(token)
	if err != nil {
		return nil, fmt.Errorf("unable to parse input token: %w", err)
	}

	// Extract claims without verification
	var claims map[string]interface{}
	if err := t.Claims(key, &claims); err != nil {
		return nil, fmt.Errorf("unable to extract claims from token: %w", err)
	}

	return struct {
		Headers []jose.Header
		Claims  map[string]interface{}
	}{
		Headers: t.Headers,
		Claims:  claims,
	}, nil
}

// Bech32Decode decodes given bech32 encoded string.
func Bech32Decode(in string) (interface{}, error) {
	hrp, data, err := bech32.Decode(in)
	if err != nil {
		return nil, fmt.Errorf("unbale to decode Bech32 encoding string: %w", err)
	}

	return struct {
		Hrp  string
		Data []byte
	}{
		Hrp:  hrp,
		Data: data,
	}, nil
}
