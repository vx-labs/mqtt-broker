// source: types.proto

/*
Package topics is a generated protocol buffer package.

It is generated from these files:
	types.proto

It has these top-level messages:
	Metadata
	RetainedMessageMetadataList
*/
package topics

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Metadata struct {
	ID          string `protobuf:"bytes,1,opt,name=ID" json:"ID,omitempty"`
	Tenant      string `protobuf:"bytes,2,opt,name=Tenant" json:"Tenant,omitempty"`
	Topic       []byte `protobuf:"bytes,3,opt,name=Topic,proto3" json:"Topic,omitempty"`
	Payload     []byte `protobuf:"bytes,4,opt,name=Payload,proto3" json:"Payload,omitempty"`
	Qos         int32  `protobuf:"varint,5,opt,name=Qos" json:"Qos,omitempty"`
	LastAdded   int64  `protobuf:"varint,6,opt,name=LastAdded" json:"LastAdded,omitempty"`
	LastDeleted int64  `protobuf:"varint,7,opt,name=LastDeleted" json:"LastDeleted,omitempty"`
}

func (m *Metadata) Reset()                    { *m = Metadata{} }
func (m *Metadata) String() string            { return proto.CompactTextString(m) }
func (*Metadata) ProtoMessage()               {}
func (*Metadata) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Metadata) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *Metadata) GetTenant() string {
	if m != nil {
		return m.Tenant
	}
	return ""
}

func (m *Metadata) GetTopic() []byte {
	if m != nil {
		return m.Topic
	}
	return nil
}

func (m *Metadata) GetPayload() []byte {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (m *Metadata) GetQos() int32 {
	if m != nil {
		return m.Qos
	}
	return 0
}

func (m *Metadata) GetLastAdded() int64 {
	if m != nil {
		return m.LastAdded
	}
	return 0
}

func (m *Metadata) GetLastDeleted() int64 {
	if m != nil {
		return m.LastDeleted
	}
	return 0
}

type RetainedMessageMetadataList struct {
	Metadatas []*Metadata `protobuf:"bytes,1,rep,name=RetainedMessages" json:"RetainedMessages,omitempty"`
}

func (m *RetainedMessageMetadataList) Reset()                    { *m = RetainedMessageMetadataList{} }
func (m *RetainedMessageMetadataList) String() string            { return proto.CompactTextString(m) }
func (*RetainedMessageMetadataList) ProtoMessage()               {}
func (*RetainedMessageMetadataList) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *RetainedMessageMetadataList) GetRetainedMessages() []*Metadata {
	if m != nil {
		return m.Metadatas
	}
	return nil
}

func init() {
	proto.RegisterType((*Metadata)(nil), "topics.Metadata")
	proto.RegisterType((*RetainedMessageMetadataList)(nil), "topics.RetainedMessageMetadataList")
}

func init() { proto.RegisterFile("types.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 225 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x90, 0x31, 0x4b, 0xc4, 0x40,
	0x10, 0x85, 0xd9, 0xc4, 0xe4, 0xbc, 0x89, 0x48, 0x18, 0x44, 0x06, 0xb4, 0x58, 0xae, 0xda, 0x2a,
	0x85, 0xb6, 0x36, 0x42, 0x9a, 0x83, 0x3b, 0xd0, 0xe5, 0x3a, 0xab, 0xd1, 0x1d, 0x24, 0x70, 0x64,
	0x83, 0x3b, 0xcd, 0xfd, 0x2f, 0x7f, 0xa0, 0x24, 0x31, 0x28, 0xda, 0xed, 0xf7, 0xbd, 0xb7, 0xc5,
	0x1b, 0xa8, 0xf4, 0x34, 0x48, 0x6a, 0x86, 0x8f, 0xa8, 0x11, 0x4b, 0x8d, 0x43, 0xf7, 0x96, 0x36,
	0x9f, 0x06, 0xce, 0xf7, 0xa2, 0x1c, 0x58, 0x19, 0x2f, 0x21, 0xdb, 0xb6, 0x64, 0xac, 0x71, 0x6b,
	0x9f, 0x6d, 0x5b, 0xbc, 0x86, 0xf2, 0x20, 0x3d, 0xf7, 0x4a, 0xd9, 0xe4, 0xbe, 0x09, 0xaf, 0xa0,
	0x38, 0x8c, 0xdf, 0x29, 0xb7, 0xc6, 0x5d, 0xf8, 0x19, 0x90, 0x60, 0xf5, 0xc4, 0xa7, 0x63, 0xe4,
	0x40, 0x67, 0x93, 0x5f, 0x10, 0x6b, 0xc8, 0x9f, 0x63, 0xa2, 0xc2, 0x1a, 0x57, 0xf8, 0xf1, 0x89,
	0xb7, 0xb0, 0xde, 0x71, 0xd2, 0xc7, 0x10, 0x24, 0x50, 0x69, 0x8d, 0xcb, 0xfd, 0x8f, 0x40, 0x0b,
	0xd5, 0x08, 0xad, 0x1c, 0x45, 0x25, 0xd0, 0x6a, 0xca, 0x7f, 0xab, 0xcd, 0x0b, 0xdc, 0x78, 0x51,
	0xee, 0x7a, 0x09, 0x7b, 0x49, 0x89, 0xdf, 0x65, 0x19, 0xb1, 0xeb, 0x92, 0xe2, 0x03, 0xd4, 0x7f,
	0xe2, 0x44, 0xc6, 0xe6, 0xae, 0xba, 0xab, 0x9b, 0x79, 0x78, 0xb3, 0xf4, 0xfd, 0xbf, 0xe6, 0x6b,
	0x39, 0x9d, 0xe8, 0xfe, 0x2b, 0x00, 0x00, 0xff, 0xff, 0x12, 0xe0, 0xa0, 0x09, 0x31, 0x01, 0x00,
	0x00,
}
