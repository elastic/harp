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

package pkcs8

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"hash"
	"io"

	"github.com/pkg/errors"
	"golang.org/x/crypto/pbkdf2"
)

// PBKDF2SaltSize is the default size of the salt for PBKDF2, 128-bit salt.
const PBKDF2SaltSize = 16

// PBKDF2Iterations is the default number of iterations for PBKDF2, 100k
// iterations. Nist recommends at least 10k, 1Passsword uses 100k.
const PBKDF2Iterations = 100000

// Encrypted pkcs8
// Based on https://github.com/youmark/pkcs8
// MIT license
type prfParam struct {
	Algo      asn1.ObjectIdentifier
	NullParam asn1.RawValue
}

type pbkdf2Params struct {
	Salt           []byte
	IterationCount int
	PrfParam       prfParam `asn1:"optional"`
}

type pbkdf2Algorithms struct {
	Algo         asn1.ObjectIdentifier
	PBKDF2Params pbkdf2Params
}

type pbkdf2Encs struct {
	EncryAlgo asn1.ObjectIdentifier
	IV        []byte
}

type pbes2Params struct {
	KeyDerivationFunc pbkdf2Algorithms
	EncryptionScheme  pbkdf2Encs
}

type encryptedlAlgorithmIdentifier struct {
	Algorithm  asn1.ObjectIdentifier
	Parameters pbes2Params
}

type encryptedPrivateKeyInfo struct {
	Algo       encryptedlAlgorithmIdentifier
	PrivateKey []byte
}

var (
	// key derivation functions
	oidPKCS5PBKDF2    = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 5, 12}
	oidPBES2          = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 5, 13}
	oidHMACWithSHA256 = asn1.ObjectIdentifier{1, 2, 840, 113549, 2, 9}

	// encryption
	oidAES128CBC = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 1, 2}
	oidAES192CBC = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 1, 22}
	oidAES256CBC = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 1, 42}
)

// rfc1423Algo holds a method for enciphering a PEM block.
//nolint:structcheck // name is used for documentation
type rfc1423Algo struct {
	cipher     x509.PEMCipher
	name       string
	cipherFunc func(key []byte) (cipher.Block, error)
	keySize    int
	blockSize  int
	identifier asn1.ObjectIdentifier
}

// rfc1423Algos holds a slice of the possible ways to encrypt a PEM
// block. The ivSize numbers were taken from the OpenSSL source.
var rfc1423Algos = []rfc1423Algo{
	{
		cipher:     x509.PEMCipherAES128,
		name:       "AES-128-CBC",
		cipherFunc: aes.NewCipher,
		keySize:    16,
		blockSize:  aes.BlockSize,
		identifier: oidAES128CBC,
	}, {
		cipher:     x509.PEMCipherAES192,
		name:       "AES-192-CBC",
		cipherFunc: aes.NewCipher,
		keySize:    24,
		blockSize:  aes.BlockSize,
		identifier: oidAES192CBC,
	}, {
		cipher:     x509.PEMCipherAES256,
		name:       "AES-256-CBC",
		cipherFunc: aes.NewCipher,
		keySize:    32,
		blockSize:  aes.BlockSize,
		identifier: oidAES256CBC,
	},
}

func cipherByKey(key x509.PEMCipher) *rfc1423Algo {
	for i := range rfc1423Algos {
		alg := &rfc1423Algos[i]
		if alg.cipher == key {
			return alg
		}
	}
	return nil
}

// deriveKey uses a key derivation function to stretch the password into a key
// with the number of bits our cipher requires. This algorithm was derived from
// the OpenSSL source.
func (c rfc1423Algo) deriveKey(password, salt []byte, h func() hash.Hash) []byte {
	return pbkdf2.Key(password, salt, PBKDF2Iterations, c.keySize, h)
}

// DecryptPEMBlock takes a password encrypted PEM block and the password used
// to encrypt it and returns a slice of decrypted DER encoded bytes.
//
// If the PEM blocks has the Proc-Type header set to "4,ENCRYPTED" it uses
// x509.DecryptPEMBlock to decrypt the block. If not it tries to decrypt the
// block using AES-128-CBC, AES-192-CBC, AES-256-CBC using the
// key derived using PBKDF2 over the given password.
func DecryptPEMBlock(block *pem.Block, password []byte) ([]byte, error) {
	if block.Headers["Proc-Type"] == "4,ENCRYPTED" {
		return x509.DecryptPEMBlock(block, password)
	}

	// PKCS#8 header defined in RFC7468 section 11
	if block.Type == "ENCRYPTED PRIVATE KEY" {
		return DecryptPKCS8PrivateKey(block.Bytes, password)
	}

	return nil, errors.New("unsupported encrypted PEM")
}

// DecryptPKCS8PrivateKey takes a password encrypted private key using the
// PKCS#8 encoding and returns the decrypted data in PKCS#8 form.
//
// It supports AES-128-CBC, AES-192-CBC, AES-256-CBC encrypted
// data using the key derived with PBKDF2 over the given password.
func DecryptPKCS8PrivateKey(data, password []byte) ([]byte, error) {
	var pki encryptedPrivateKeyInfo
	if _, err := asn1.Unmarshal(data, &pki); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal private key")
	}

	if !pki.Algo.Algorithm.Equal(oidPBES2) {
		return nil, errors.New("unsupported encrypted PEM: only PBES2 is supported")
	}

	if !pki.Algo.Parameters.KeyDerivationFunc.Algo.Equal(oidPKCS5PBKDF2) {
		return nil, errors.New("unsupported encrypted PEM: only PBKDF2 is supported")
	}

	encParam := pki.Algo.Parameters.EncryptionScheme
	kdfParam := pki.Algo.Parameters.KeyDerivationFunc.PBKDF2Params

	iv := encParam.IV
	salt := kdfParam.Salt
	iter := kdfParam.IterationCount

	// pbkdf2 hash function
	keyHash := sha256.New

	encryptedKey := pki.PrivateKey
	var symkey []byte
	var block cipher.Block
	var err error
	switch {
	// AES-128-CBC, AES-192-CBC, AES-256-CBC
	case encParam.EncryAlgo.Equal(oidAES128CBC):
		symkey = pbkdf2.Key(password, salt, iter, 16, keyHash)
		block, err = aes.NewCipher(symkey)
	case encParam.EncryAlgo.Equal(oidAES192CBC):
		symkey = pbkdf2.Key(password, salt, iter, 24, keyHash)
		block, err = aes.NewCipher(symkey)
	case encParam.EncryAlgo.Equal(oidAES256CBC):
		symkey = pbkdf2.Key(password, salt, iter, 32, keyHash)
		block, err = aes.NewCipher(symkey)
	default:
		return nil, errors.Errorf("unsupported encrypted PEM: unknown algorithm %v", encParam.EncryAlgo)
	}
	if err != nil {
		return nil, err
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(encryptedKey, encryptedKey)

	return encryptedKey, nil
}

// EncryptPKCS8PrivateKey returns a PEM block holding the given PKCS#8 encroded
// private key, encrypted with the specified algorithm and a PBKDF2 derived key
// from the given password.
func EncryptPKCS8PrivateKey(rand io.Reader, data, password []byte, alg x509.PEMCipher) (*pem.Block, error) {
	ciph := cipherByKey(alg)
	if ciph == nil {
		return nil, errors.Errorf("failed to encrypt PEM: unknown algorithm %v", alg)
	}

	salt := make([]byte, PBKDF2SaltSize)
	if _, err := io.ReadFull(rand, salt); err != nil {
		return nil, errors.Wrap(err, "failed to generate salt")
	}
	iv := make([]byte, ciph.blockSize)
	if _, err := io.ReadFull(rand, iv); err != nil {
		return nil, errors.Wrap(err, "failed to generate IV")
	}

	key := ciph.deriveKey(password, salt, sha256.New)
	block, err := ciph.cipherFunc(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cipher")
	}
	enc := cipher.NewCBCEncrypter(block, iv)
	pad := ciph.blockSize - len(data)%ciph.blockSize
	encrypted := make([]byte, len(data), len(data)+pad)
	// We could save this copy by encrypting all the whole blocks in
	// the data separately, but it doesn't seem worth the additional
	// code.
	copy(encrypted, data)
	// See RFC 1423, section 1.1
	for i := 0; i < pad; i++ {
		encrypted = append(encrypted, byte(pad))
	}
	enc.CryptBlocks(encrypted, encrypted)

	// Build encrypted ans1 data
	pki := encryptedPrivateKeyInfo{
		Algo: encryptedlAlgorithmIdentifier{
			Algorithm: oidPBES2,
			Parameters: pbes2Params{
				KeyDerivationFunc: pbkdf2Algorithms{
					Algo: oidPKCS5PBKDF2,
					PBKDF2Params: pbkdf2Params{
						Salt:           salt,
						IterationCount: PBKDF2Iterations,
						PrfParam: prfParam{
							Algo: oidHMACWithSHA256,
						},
					},
				},
				EncryptionScheme: pbkdf2Encs{
					EncryAlgo: ciph.identifier,
					IV:        iv,
				},
			},
		},
		PrivateKey: encrypted,
	}

	b, err := asn1.Marshal(pki)
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling encrypted key")
	}
	return &pem.Block{
		Type:  "ENCRYPTED PRIVATE KEY",
		Bytes: b,
	}, nil
}
