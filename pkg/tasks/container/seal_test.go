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

package container

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/awnumar/memguard"
	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/tasks"
	fuzz "github.com/google/gofuzz"
)

func TestSealTask_Run(t *testing.T) {
	type fields struct {
		ContainerReader          tasks.ReaderProvider
		SealedContainerWriter    tasks.WriterProvider
		OutputWriter             tasks.WriterProvider
		PeerPublicKeys           []*[32]byte
		DCKDMasterKey            *memguard.LockedBuffer
		DCKDTarget               string
		JSONOutput               bool
		DisableContainerIdentity bool
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "nil",
			wantErr: true,
		},
		{
			name: "nil containerReader",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
			},
			wantErr: true,
		},
		{
			name: "nil sealedContainerWriter",
			fields: fields{
				ContainerReader:       cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				SealedContainerWriter: nil,
			},
			wantErr: true,
		},
		{
			name: "nil outputWriter",
			fields: fields{
				ContainerReader:       cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				SealedContainerWriter: cmdutil.DiscardWriter(),
				OutputWriter:          nil,
			},
			wantErr: true,
		},
		{
			name: "no public keys",
			fields: fields{
				ContainerReader:       cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				SealedContainerWriter: cmdutil.DiscardWriter(),
				OutputWriter:          cmdutil.DiscardWriter(),
				PeerPublicKeys:        []*[32]byte{},
			},
			wantErr: true,
		},
		{
			name: "no public keys",
			fields: fields{
				ContainerReader:       cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				SealedContainerWriter: cmdutil.DiscardWriter(),
				OutputWriter:          cmdutil.DiscardWriter(),
				PeerPublicKeys:        []*[32]byte{},
			},
			wantErr: true,
		},
		{
			name: "containerReader error",
			fields: fields{
				ContainerReader:       cmdutil.FileReader("non-existent.bundle"),
				SealedContainerWriter: cmdutil.DiscardWriter(),
				OutputWriter:          cmdutil.DiscardWriter(),
				PeerPublicKeys: []*[32]byte{
					{
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "containerReader not a bundle",
			fields: fields{
				ContainerReader:       cmdutil.FileReader("../../../test/fixtures/bundles/complete.json"),
				SealedContainerWriter: cmdutil.DiscardWriter(),
				OutputWriter:          cmdutil.DiscardWriter(),
				PeerPublicKeys: []*[32]byte{
					{
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "low-order public keys",
			fields: fields{
				ContainerReader:       cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				SealedContainerWriter: cmdutil.DiscardWriter(),
				OutputWriter:          cmdutil.DiscardWriter(),
				PeerPublicKeys: []*[32]byte{
					{
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "sealedContainerWriter error",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				SealedContainerWriter: func(ctx context.Context) (io.Writer, error) {
					return nil, errors.New("test")
				},
				OutputWriter: cmdutil.DiscardWriter(),
				PeerPublicKeys: []*[32]byte{
					{
						0x97, 0x75, 0x9e, 0x17, 0x35, 0x8a, 0x5b, 0xae, 0x6b, 0x5a, 0xfc, 0xde, 0x97, 0x40, 0x84, 0x7f,
						0xad, 0x59, 0xe6, 0x0a, 0x25, 0x81, 0xbe, 0xcd, 0xc6, 0xa0, 0x37, 0x0e, 0x0b, 0x66, 0x1d, 0x49,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "sealedContainerWriter closed",
			fields: fields{
				ContainerReader: cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				SealedContainerWriter: func(ctx context.Context) (io.Writer, error) {
					return cmdutil.NewClosedWriter(), nil
				},
				OutputWriter: cmdutil.DiscardWriter(),
				PeerPublicKeys: []*[32]byte{
					{
						0x97, 0x75, 0x9e, 0x17, 0x35, 0x8a, 0x5b, 0xae, 0x6b, 0x5a, 0xfc, 0xde, 0x97, 0x40, 0x84, 0x7f,
						0xad, 0x59, 0xe6, 0x0a, 0x25, 0x81, 0xbe, 0xcd, 0xc6, 0xa0, 0x37, 0x0e, 0x0b, 0x66, 0x1d, 0x49,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "outputWriter error",
			fields: fields{
				ContainerReader:       cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				SealedContainerWriter: cmdutil.DiscardWriter(),
				OutputWriter: func(ctx context.Context) (io.Writer, error) {
					return nil, errors.New("test")
				},
				PeerPublicKeys: []*[32]byte{
					{
						0x97, 0x75, 0x9e, 0x17, 0x35, 0x8a, 0x5b, 0xae, 0x6b, 0x5a, 0xfc, 0xde, 0x97, 0x40, 0x84, 0x7f,
						0xad, 0x59, 0xe6, 0x0a, 0x25, 0x81, 0xbe, 0xcd, 0xc6, 0xa0, 0x37, 0x0e, 0x0b, 0x66, 0x1d, 0x49,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "outputWriter closed",
			fields: fields{
				ContainerReader:       cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				SealedContainerWriter: cmdutil.DiscardWriter(),
				OutputWriter: func(ctx context.Context) (io.Writer, error) {
					return cmdutil.NewClosedWriter(), nil
				},
				PeerPublicKeys: []*[32]byte{
					{
						0x97, 0x75, 0x9e, 0x17, 0x35, 0x8a, 0x5b, 0xae, 0x6b, 0x5a, 0xfc, 0xde, 0x97, 0x40, 0x84, 0x7f,
						0xad, 0x59, 0xe6, 0x0a, 0x25, 0x81, 0xbe, 0xcd, 0xc6, 0xa0, 0x37, 0x0e, 0x0b, 0x66, 0x1d, 0x49,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "outputWriter closed - json",
			fields: fields{
				ContainerReader:       cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				SealedContainerWriter: cmdutil.DiscardWriter(),
				OutputWriter: func(ctx context.Context) (io.Writer, error) {
					return cmdutil.NewClosedWriter(), nil
				},
				PeerPublicKeys: []*[32]byte{
					{
						0x97, 0x75, 0x9e, 0x17, 0x35, 0x8a, 0x5b, 0xae, 0x6b, 0x5a, 0xfc, 0xde, 0x97, 0x40, 0x84, 0x7f,
						0xad, 0x59, 0xe6, 0x0a, 0x25, 0x81, 0xbe, 0xcd, 0xc6, 0xa0, 0x37, 0x0e, 0x0b, 0x66, 0x1d, 0x49,
					},
				},
				JSONOutput: true,
			},
			wantErr: true,
		},
		// ---------------------------------------------------------------------
		{
			name: "valid",
			fields: fields{
				ContainerReader:       cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				SealedContainerWriter: cmdutil.DiscardWriter(),
				OutputWriter:          cmdutil.DiscardWriter(),
				PeerPublicKeys: []*[32]byte{
					{
						0x97, 0x75, 0x9e, 0x17, 0x35, 0x8a, 0x5b, 0xae, 0x6b, 0x5a, 0xfc, 0xde, 0x97, 0x40, 0x84, 0x7f,
						0xad, 0x59, 0xe6, 0x0a, 0x25, 0x81, 0xbe, 0xcd, 0xc6, 0xa0, 0x37, 0x0e, 0x0b, 0x66, 0x1d, 0x49,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid - no container identity",
			fields: fields{
				ContainerReader:       cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				SealedContainerWriter: cmdutil.DiscardWriter(),
				OutputWriter:          cmdutil.DiscardWriter(),
				PeerPublicKeys: []*[32]byte{
					{
						0x97, 0x75, 0x9e, 0x17, 0x35, 0x8a, 0x5b, 0xae, 0x6b, 0x5a, 0xfc, 0xde, 0x97, 0x40, 0x84, 0x7f,
						0xad, 0x59, 0xe6, 0x0a, 0x25, 0x81, 0xbe, 0xcd, 0xc6, 0xa0, 0x37, 0x0e, 0x0b, 0x66, 0x1d, 0x49,
					},
				},
				DisableContainerIdentity: true,
			},
			wantErr: false,
		},
		{
			name: "valid - json output",
			fields: fields{
				ContainerReader:       cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				SealedContainerWriter: cmdutil.DiscardWriter(),
				OutputWriter:          cmdutil.DiscardWriter(),
				PeerPublicKeys: []*[32]byte{
					{
						0x97, 0x75, 0x9e, 0x17, 0x35, 0x8a, 0x5b, 0xae, 0x6b, 0x5a, 0xfc, 0xde, 0x97, 0x40, 0x84, 0x7f,
						0xad, 0x59, 0xe6, 0x0a, 0x25, 0x81, 0xbe, 0xcd, 0xc6, 0xa0, 0x37, 0x0e, 0x0b, 0x66, 0x1d, 0x49,
					},
				},
				JSONOutput: true,
			},
			wantErr: false,
		},
		{
			name: "valid - dckd",
			fields: fields{
				ContainerReader:       cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
				SealedContainerWriter: cmdutil.DiscardWriter(),
				OutputWriter:          cmdutil.DiscardWriter(),
				DCKDMasterKey:         memguard.NewBuffer(32),
				DCKDTarget:            "test",
				PeerPublicKeys: []*[32]byte{
					{
						0x97, 0x75, 0x9e, 0x17, 0x35, 0x8a, 0x5b, 0xae, 0x6b, 0x5a, 0xfc, 0xde, 0x97, 0x40, 0x84, 0x7f,
						0xad, 0x59, 0xe6, 0x0a, 0x25, 0x81, 0xbe, 0xcd, 0xc6, 0xa0, 0x37, 0x0e, 0x0b, 0x66, 0x1d, 0x49,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &SealTask{
				ContainerReader:          tt.fields.ContainerReader,
				SealedContainerWriter:    tt.fields.SealedContainerWriter,
				OutputWriter:             tt.fields.OutputWriter,
				PeerPublicKeys:           tt.fields.PeerPublicKeys,
				DCKDMasterKey:            tt.fields.DCKDMasterKey,
				DCKDTarget:               tt.fields.DCKDTarget,
				JSONOutput:               tt.fields.JSONOutput,
				DisableContainerIdentity: tt.fields.DisableContainerIdentity,
			}
			if err := tr.Run(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("SealTask.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSealTask_Fuzz(t *testing.T) {
	tsk := &SealTask{
		ContainerReader:          cmdutil.FileReader("../../../test/fixtures/bundles/complete.bundle"),
		SealedContainerWriter:    cmdutil.DiscardWriter(),
		OutputWriter:             cmdutil.DiscardWriter(),
		PeerPublicKeys:           []*[32]byte{},
		DisableContainerIdentity: true,
	}

	// Making sure the function never panics
	for i := 0; i < 50; i++ {
		f := fuzz.New()

		// Prepare arguments
		f.Fuzz(&tsk.PeerPublicKeys)

		// Execute
		tsk.Run(context.Background())
	}
}
