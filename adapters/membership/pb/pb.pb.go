// Code generated by protoc-gen-go. DO NOT EDIT.
// source: pb.proto

package pb

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

type Part struct {
	Key                  string   `protobuf:"bytes,1,opt,name=Key,proto3" json:"Key,omitempty"`
	Data                 []byte   `protobuf:"bytes,2,opt,name=Data,proto3" json:"Data,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Part) Reset()         { *m = Part{} }
func (m *Part) String() string { return proto.CompactTextString(m) }
func (*Part) ProtoMessage()    {}
func (*Part) Descriptor() ([]byte, []int) {
	return fileDescriptor_f80abaa17e25ccc8, []int{0}
}

func (m *Part) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Part.Unmarshal(m, b)
}
func (m *Part) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Part.Marshal(b, m, deterministic)
}
func (m *Part) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Part.Merge(m, src)
}
func (m *Part) XXX_Size() int {
	return xxx_messageInfo_Part.Size(m)
}
func (m *Part) XXX_DiscardUnknown() {
	xxx_messageInfo_Part.DiscardUnknown(m)
}

var xxx_messageInfo_Part proto.InternalMessageInfo

func (m *Part) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *Part) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

type FullState struct {
	Parts                []*Part  `protobuf:"bytes,1,rep,name=parts,proto3" json:"parts,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *FullState) Reset()         { *m = FullState{} }
func (m *FullState) String() string { return proto.CompactTextString(m) }
func (*FullState) ProtoMessage()    {}
func (*FullState) Descriptor() ([]byte, []int) {
	return fileDescriptor_f80abaa17e25ccc8, []int{1}
}

func (m *FullState) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FullState.Unmarshal(m, b)
}
func (m *FullState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FullState.Marshal(b, m, deterministic)
}
func (m *FullState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FullState.Merge(m, src)
}
func (m *FullState) XXX_Size() int {
	return xxx_messageInfo_FullState.Size(m)
}
func (m *FullState) XXX_DiscardUnknown() {
	xxx_messageInfo_FullState.DiscardUnknown(m)
}

var xxx_messageInfo_FullState proto.InternalMessageInfo

func (m *FullState) GetParts() []*Part {
	if m != nil {
		return m.Parts
	}
	return nil
}

func init() {
	proto.RegisterType((*Part)(nil), "pb.Part")
	proto.RegisterType((*FullState)(nil), "pb.FullState")
}

func init() { proto.RegisterFile("pb.proto", fileDescriptor_f80abaa17e25ccc8) }

var fileDescriptor_f80abaa17e25ccc8 = []byte{
	// 123 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x28, 0x48, 0xd2, 0x2b,
	0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x2a, 0x48, 0x52, 0xd2, 0xe1, 0x62, 0x09, 0x48, 0x2c, 0x2a,
	0x11, 0x12, 0xe0, 0x62, 0xf6, 0x4e, 0xad, 0x94, 0x60, 0x54, 0x60, 0xd4, 0xe0, 0x0c, 0x02, 0x31,
	0x85, 0x84, 0xb8, 0x58, 0x5c, 0x12, 0x4b, 0x12, 0x25, 0x98, 0x14, 0x18, 0x35, 0x78, 0x82, 0xc0,
	0x6c, 0x25, 0x6d, 0x2e, 0x4e, 0xb7, 0xd2, 0x9c, 0x9c, 0xe0, 0x92, 0xc4, 0x92, 0x54, 0x21, 0x39,
	0x2e, 0xd6, 0x82, 0xc4, 0xa2, 0x92, 0x62, 0x09, 0x46, 0x05, 0x66, 0x0d, 0x6e, 0x23, 0x0e, 0xbd,
	0x82, 0x24, 0x3d, 0x90, 0x59, 0x41, 0x10, 0xe1, 0x24, 0x36, 0xb0, 0x2d, 0xc6, 0x80, 0x00, 0x00,
	0x00, 0xff, 0xff, 0xcb, 0xd2, 0x76, 0x9a, 0x71, 0x00, 0x00, 0x00,
}
