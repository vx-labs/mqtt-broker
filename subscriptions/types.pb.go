// Code generated by protoc-gen-go. DO NOT EDIT.
// source: types.proto

package subscriptions

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Metadata struct {
	ID                   string   `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	SessionID            string   `protobuf:"bytes,2,opt,name=SessionID,proto3" json:"SessionID,omitempty"`
	Tenant               string   `protobuf:"bytes,3,opt,name=Tenant,proto3" json:"Tenant,omitempty"`
	Pattern              []byte   `protobuf:"bytes,4,opt,name=Pattern,proto3" json:"Pattern,omitempty"`
	Qos                  int32    `protobuf:"varint,5,opt,name=Qos,proto3" json:"Qos,omitempty"`
	Peer                 string   `protobuf:"bytes,6,opt,name=Peer,proto3" json:"Peer,omitempty"`
	LastAdded            int64    `protobuf:"varint,7,opt,name=LastAdded,proto3" json:"LastAdded,omitempty"`
	LastDeleted          int64    `protobuf:"varint,8,opt,name=LastDeleted,proto3" json:"LastDeleted,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Metadata) Reset()         { *m = Metadata{} }
func (m *Metadata) String() string { return proto.CompactTextString(m) }
func (*Metadata) ProtoMessage()    {}
func (*Metadata) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{0}
}

func (m *Metadata) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Metadata.Unmarshal(m, b)
}
func (m *Metadata) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Metadata.Marshal(b, m, deterministic)
}
func (m *Metadata) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Metadata.Merge(m, src)
}
func (m *Metadata) XXX_Size() int {
	return xxx_messageInfo_Metadata.Size(m)
}
func (m *Metadata) XXX_DiscardUnknown() {
	xxx_messageInfo_Metadata.DiscardUnknown(m)
}

var xxx_messageInfo_Metadata proto.InternalMessageInfo

func (m *Metadata) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *Metadata) GetSessionID() string {
	if m != nil {
		return m.SessionID
	}
	return ""
}

func (m *Metadata) GetTenant() string {
	if m != nil {
		return m.Tenant
	}
	return ""
}

func (m *Metadata) GetPattern() []byte {
	if m != nil {
		return m.Pattern
	}
	return nil
}

func (m *Metadata) GetQos() int32 {
	if m != nil {
		return m.Qos
	}
	return 0
}

func (m *Metadata) GetPeer() string {
	if m != nil {
		return m.Peer
	}
	return ""
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

type SubscriptionMetadataList struct {
	Metadatas            []*Metadata `protobuf:"bytes,1,rep,name=Metadatas,proto3" json:"Metadatas,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *SubscriptionMetadataList) Reset()         { *m = SubscriptionMetadataList{} }
func (m *SubscriptionMetadataList) String() string { return proto.CompactTextString(m) }
func (*SubscriptionMetadataList) ProtoMessage()    {}
func (*SubscriptionMetadataList) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{1}
}

func (m *SubscriptionMetadataList) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SubscriptionMetadataList.Unmarshal(m, b)
}
func (m *SubscriptionMetadataList) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SubscriptionMetadataList.Marshal(b, m, deterministic)
}
func (m *SubscriptionMetadataList) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SubscriptionMetadataList.Merge(m, src)
}
func (m *SubscriptionMetadataList) XXX_Size() int {
	return xxx_messageInfo_SubscriptionMetadataList.Size(m)
}
func (m *SubscriptionMetadataList) XXX_DiscardUnknown() {
	xxx_messageInfo_SubscriptionMetadataList.DiscardUnknown(m)
}

var xxx_messageInfo_SubscriptionMetadataList proto.InternalMessageInfo

func (m *SubscriptionMetadataList) GetMetadatas() []*Metadata {
	if m != nil {
		return m.Metadatas
	}
	return nil
}

func init() {
	proto.RegisterType((*Metadata)(nil), "subscriptions.Metadata")
	proto.RegisterType((*SubscriptionMetadataList)(nil), "subscriptions.SubscriptionMetadataList")
}

func init() { proto.RegisterFile("types.proto", fileDescriptor_d938547f84707355) }

var fileDescriptor_d938547f84707355 = []byte{
	// 236 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x90, 0xc1, 0x4a, 0xc3, 0x40,
	0x10, 0x86, 0xd9, 0xa4, 0x4d, 0x9b, 0x89, 0x8a, 0xcc, 0x41, 0xe7, 0xe0, 0x61, 0xe9, 0x29, 0xa7,
	0x1c, 0x14, 0x1f, 0x40, 0xc8, 0x25, 0x50, 0xa1, 0x4d, 0x7d, 0x81, 0xad, 0x99, 0x43, 0x40, 0x76,
	0x43, 0x66, 0x3c, 0xf8, 0x96, 0x3e, 0x92, 0x64, 0x31, 0xa4, 0xbd, 0xcd, 0xf7, 0xfd, 0xcb, 0x0e,
	0xf3, 0x43, 0xa1, 0x3f, 0x03, 0x4b, 0x35, 0x8c, 0x41, 0x03, 0xde, 0xca, 0xf7, 0x59, 0x3e, 0xc7,
	0x7e, 0xd0, 0x3e, 0x78, 0xd9, 0xfd, 0x1a, 0xd8, 0xbe, 0xb3, 0xba, 0xce, 0xa9, 0xc3, 0x3b, 0x48,
	0x9a, 0x9a, 0x8c, 0x35, 0x65, 0xde, 0x26, 0x4d, 0x8d, 0x4f, 0x90, 0x9f, 0x58, 0xa4, 0x0f, 0xbe,
	0xa9, 0x29, 0x89, 0x7a, 0x11, 0xf8, 0x00, 0xd9, 0x07, 0x7b, 0xe7, 0x95, 0xd2, 0x18, 0xfd, 0x13,
	0x12, 0x6c, 0x0e, 0x4e, 0x95, 0x47, 0x4f, 0x2b, 0x6b, 0xca, 0x9b, 0x76, 0x46, 0xbc, 0x87, 0xf4,
	0x18, 0x84, 0xd6, 0xd6, 0x94, 0xeb, 0x76, 0x1a, 0x11, 0x61, 0x75, 0x60, 0x1e, 0x29, 0x8b, 0x3f,
	0xc4, 0x79, 0xda, 0xba, 0x77, 0xa2, 0x6f, 0x5d, 0xc7, 0x1d, 0x6d, 0xac, 0x29, 0xd3, 0x76, 0x11,
	0x68, 0xa1, 0x98, 0xa0, 0xe6, 0x2f, 0x56, 0xee, 0x68, 0x1b, 0xf3, 0x4b, 0xb5, 0x3b, 0x02, 0x9d,
	0x2e, 0x6e, 0x9c, 0xaf, 0xdb, 0xf7, 0xa2, 0xf8, 0x0a, 0xf9, 0xcc, 0x42, 0xc6, 0xa6, 0x65, 0xf1,
	0xfc, 0x58, 0x5d, 0x35, 0x52, 0xcd, 0x79, 0xbb, 0xbc, 0x3c, 0x67, 0xb1, 0xbb, 0x97, 0xbf, 0x00,
	0x00, 0x00, 0xff, 0xff, 0x9d, 0xe1, 0x8e, 0x30, 0x4a, 0x01, 0x00, 0x00,
}
