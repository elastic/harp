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

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.4
// source: harp/container/v1/container.proto

package containerv1

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Header describes container headers.
type Header struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Content encoding describes the content encoding used for raw.
	// Unspecified means no encoding.
	ContentEncoding string `protobuf:"bytes,1,opt,name=content_encoding,json=contentEncoding,proto3" json:"content_encoding,omitempty"`
	// Content type is the serialization method used to serialize 'raw'.
	// Unspecified means "application/vnd.harp.protobuf".
	ContentType string `protobuf:"bytes,2,opt,name=content_type,json=contentType,proto3" json:"content_type,omitempty"`
	// Ephemeral public key used for encryption.
	EncryptionPublicKey []byte `protobuf:"bytes,3,opt,name=encryption_public_key,json=encryptionPublicKey,proto3" json:"encryption_public_key,omitempty"`
	// Container box contains public signing key encrypted with payload key.
	ContainerBox []byte `protobuf:"bytes,4,opt,name=container_box,json=containerBox,proto3" json:"container_box,omitempty"`
	// Recipient list for identity bound secret container.
	Recipients []*Recipient `protobuf:"bytes,6,rep,name=recipients,proto3" json:"recipients,omitempty"`
	// Seal strategy
	SealVersion uint32 `protobuf:"varint,7,opt,name=seal_version,json=sealVersion,proto3" json:"seal_version,omitempty"`
}

func (x *Header) Reset() {
	*x = Header{}
	if protoimpl.UnsafeEnabled {
		mi := &file_harp_container_v1_container_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Header) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Header) ProtoMessage() {}

func (x *Header) ProtoReflect() protoreflect.Message {
	mi := &file_harp_container_v1_container_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Header.ProtoReflect.Descriptor instead.
func (*Header) Descriptor() ([]byte, []int) {
	return file_harp_container_v1_container_proto_rawDescGZIP(), []int{0}
}

func (x *Header) GetContentEncoding() string {
	if x != nil {
		return x.ContentEncoding
	}
	return ""
}

func (x *Header) GetContentType() string {
	if x != nil {
		return x.ContentType
	}
	return ""
}

func (x *Header) GetEncryptionPublicKey() []byte {
	if x != nil {
		return x.EncryptionPublicKey
	}
	return nil
}

func (x *Header) GetContainerBox() []byte {
	if x != nil {
		return x.ContainerBox
	}
	return nil
}

func (x *Header) GetRecipients() []*Recipient {
	if x != nil {
		return x.Recipients
	}
	return nil
}

func (x *Header) GetSealVersion() uint32 {
	if x != nil {
		return x.SealVersion
	}
	return 0
}

// Recipient describes container recipient informations.
type Recipient struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Recipient identifier
	Identifier []byte `protobuf:"bytes,1,opt,name=identifier,proto3" json:"identifier,omitempty"`
	// Encrypted copy of the payload key for recipient.
	Key []byte `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`
}

func (x *Recipient) Reset() {
	*x = Recipient{}
	if protoimpl.UnsafeEnabled {
		mi := &file_harp_container_v1_container_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Recipient) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Recipient) ProtoMessage() {}

func (x *Recipient) ProtoReflect() protoreflect.Message {
	mi := &file_harp_container_v1_container_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Recipient.ProtoReflect.Descriptor instead.
func (*Recipient) Descriptor() ([]byte, []int) {
	return file_harp_container_v1_container_proto_rawDescGZIP(), []int{1}
}

func (x *Recipient) GetIdentifier() []byte {
	if x != nil {
		return x.Identifier
	}
	return nil
}

func (x *Recipient) GetKey() []byte {
	if x != nil {
		return x.Key
	}
	return nil
}

// Container describes the container attributes.
type Container struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Container headers.
	Headers *Header `protobuf:"bytes,1,opt,name=headers,proto3" json:"headers,omitempty"`
	// Raw hold the complete serialized object in protobuf.
	Raw []byte `protobuf:"bytes,2,opt,name=raw,proto3" json:"raw,omitempty"`
}

func (x *Container) Reset() {
	*x = Container{}
	if protoimpl.UnsafeEnabled {
		mi := &file_harp_container_v1_container_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Container) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Container) ProtoMessage() {}

func (x *Container) ProtoReflect() protoreflect.Message {
	mi := &file_harp_container_v1_container_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Container.ProtoReflect.Descriptor instead.
func (*Container) Descriptor() ([]byte, []int) {
	return file_harp_container_v1_container_proto_rawDescGZIP(), []int{2}
}

func (x *Container) GetHeaders() *Header {
	if x != nil {
		return x.Headers
	}
	return nil
}

func (x *Container) GetRaw() []byte {
	if x != nil {
		return x.Raw
	}
	return nil
}

var File_harp_container_v1_container_proto protoreflect.FileDescriptor

var file_harp_container_v1_container_proto_rawDesc = []byte{
	0x0a, 0x21, 0x68, 0x61, 0x72, 0x70, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72,
	0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x11, 0x68, 0x61, 0x72, 0x70, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69,
	0x6e, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x22, 0x90, 0x02, 0x0a, 0x06, 0x48, 0x65, 0x61, 0x64, 0x65,
	0x72, 0x12, 0x29, 0x0a, 0x10, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x5f, 0x65, 0x6e, 0x63,
	0x6f, 0x64, 0x69, 0x6e, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x63, 0x6f, 0x6e,
	0x74, 0x65, 0x6e, 0x74, 0x45, 0x6e, 0x63, 0x6f, 0x64, 0x69, 0x6e, 0x67, 0x12, 0x21, 0x0a, 0x0c,
	0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0b, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12,
	0x32, 0x0a, 0x15, 0x65, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x70, 0x75,
	0x62, 0x6c, 0x69, 0x63, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x13,
	0x65, 0x6e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63,
	0x4b, 0x65, 0x79, 0x12, 0x23, 0x0a, 0x0d, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72,
	0x5f, 0x62, 0x6f, 0x78, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0c, 0x63, 0x6f, 0x6e, 0x74,
	0x61, 0x69, 0x6e, 0x65, 0x72, 0x42, 0x6f, 0x78, 0x12, 0x3c, 0x0a, 0x0a, 0x72, 0x65, 0x63, 0x69,
	0x70, 0x69, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x68,
	0x61, 0x72, 0x70, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x2e, 0x76, 0x31,
	0x2e, 0x52, 0x65, 0x63, 0x69, 0x70, 0x69, 0x65, 0x6e, 0x74, 0x52, 0x0a, 0x72, 0x65, 0x63, 0x69,
	0x70, 0x69, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x21, 0x0a, 0x0c, 0x73, 0x65, 0x61, 0x6c, 0x5f, 0x76,
	0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0b, 0x73, 0x65,
	0x61, 0x6c, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0x3d, 0x0a, 0x09, 0x52, 0x65, 0x63,
	0x69, 0x70, 0x69, 0x65, 0x6e, 0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69,
	0x66, 0x69, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0a, 0x69, 0x64, 0x65, 0x6e,
	0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x22, 0x52, 0x0a, 0x09, 0x43, 0x6f, 0x6e, 0x74,
	0x61, 0x69, 0x6e, 0x65, 0x72, 0x12, 0x33, 0x0a, 0x07, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x68, 0x61, 0x72, 0x70, 0x2e, 0x63, 0x6f,
	0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x48, 0x65, 0x61, 0x64, 0x65,
	0x72, 0x52, 0x07, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x12, 0x10, 0x0a, 0x03, 0x72, 0x61,
	0x77, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x72, 0x61, 0x77, 0x42, 0xb1, 0x01, 0x0a,
	0x2d, 0x63, 0x6f, 0x6d, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x65, 0x6c, 0x61, 0x73,
	0x74, 0x69, 0x63, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x73, 0x65, 0x63, 0x2e, 0x68, 0x61, 0x72,
	0x70, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x42, 0x0e,
	0x43, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01,
	0x5a, 0x40, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x65, 0x6c, 0x61,
	0x73, 0x74, 0x69, 0x63, 0x2f, 0x68, 0x61, 0x72, 0x70, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x67, 0x65,
	0x6e, 0x2f, 0x67, 0x6f, 0x2f, 0x68, 0x61, 0x72, 0x70, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69,
	0x6e, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x3b, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72,
	0x76, 0x31, 0xa2, 0x02, 0x03, 0x53, 0x43, 0x58, 0xaa, 0x02, 0x11, 0x68, 0x61, 0x72, 0x70, 0x2e,
	0x43, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x11, 0x68,
	0x61, 0x72, 0x70, 0x5c, 0x43, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x5c, 0x56, 0x31,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_harp_container_v1_container_proto_rawDescOnce sync.Once
	file_harp_container_v1_container_proto_rawDescData = file_harp_container_v1_container_proto_rawDesc
)

func file_harp_container_v1_container_proto_rawDescGZIP() []byte {
	file_harp_container_v1_container_proto_rawDescOnce.Do(func() {
		file_harp_container_v1_container_proto_rawDescData = protoimpl.X.CompressGZIP(file_harp_container_v1_container_proto_rawDescData)
	})
	return file_harp_container_v1_container_proto_rawDescData
}

var file_harp_container_v1_container_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_harp_container_v1_container_proto_goTypes = []interface{}{
	(*Header)(nil),    // 0: harp.container.v1.Header
	(*Recipient)(nil), // 1: harp.container.v1.Recipient
	(*Container)(nil), // 2: harp.container.v1.Container
}
var file_harp_container_v1_container_proto_depIdxs = []int32{
	1, // 0: harp.container.v1.Header.recipients:type_name -> harp.container.v1.Recipient
	0, // 1: harp.container.v1.Container.headers:type_name -> harp.container.v1.Header
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_harp_container_v1_container_proto_init() }
func file_harp_container_v1_container_proto_init() {
	if File_harp_container_v1_container_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_harp_container_v1_container_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Header); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_harp_container_v1_container_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Recipient); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_harp_container_v1_container_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Container); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_harp_container_v1_container_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_harp_container_v1_container_proto_goTypes,
		DependencyIndexes: file_harp_container_v1_container_proto_depIdxs,
		MessageInfos:      file_harp_container_v1_container_proto_msgTypes,
	}.Build()
	File_harp_container_v1_container_proto = out.File
	file_harp_container_v1_container_proto_rawDesc = nil
	file_harp_container_v1_container_proto_goTypes = nil
	file_harp_container_v1_container_proto_depIdxs = nil
}
