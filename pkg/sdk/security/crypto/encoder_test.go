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
	"testing"

	_ "golang.org/x/crypto/blake2b"
)

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
