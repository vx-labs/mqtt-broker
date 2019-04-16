// Code generated by protoc-gen-go. DO NOT EDIT.
// source: types.proto

/*
Package peers is a generated protocol buffer package.

It is generated from these files:
	types.proto

It has these top-level messages:
	Peer
	ComputeUsage
	MemoryUsage
	PeerList
*/
package peers

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

type Peer struct {
	ID           string        `protobuf:"bytes,1,opt,name=ID" json:"ID,omitempty"`
	MeshID       string        `protobuf:"bytes,2,opt,name=MeshID" json:"MeshID,omitempty"`
	Hostname     string        `protobuf:"bytes,3,opt,name=Hostname" json:"Hostname,omitempty"`
	Address      string        `protobuf:"bytes,4,opt,name=Address" json:"Address,omitempty"`
	LastAdded    int64         `protobuf:"varint,5,opt,name=LastAdded" json:"LastAdded,omitempty"`
	LastDeleted  int64         `protobuf:"varint,6,opt,name=LastDeleted" json:"LastDeleted,omitempty"`
	MemoryUsage  *MemoryUsage  `protobuf:"bytes,7,opt,name=MemoryUsage" json:"MemoryUsage,omitempty"`
	ComputeUsage *ComputeUsage `protobuf:"bytes,8,opt,name=ComputeUsage" json:"ComputeUsage,omitempty"`
	Runtime      string        `protobuf:"bytes,9,opt,name=Runtime" json:"Runtime,omitempty"`
	Services     []string      `protobuf:"bytes,10,rep,name=Services" json:"Services,omitempty"`
	Started      int64         `protobuf:"varint,11,opt,name=Started" json:"Started,omitempty"`
}

func (m *Peer) Reset()                    { *m = Peer{} }
func (m *Peer) String() string            { return proto.CompactTextString(m) }
func (*Peer) ProtoMessage()               {}
func (*Peer) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Peer) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *Peer) GetMeshID() string {
	if m != nil {
		return m.MeshID
	}
	return ""
}

func (m *Peer) GetHostname() string {
	if m != nil {
		return m.Hostname
	}
	return ""
}

func (m *Peer) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *Peer) GetLastAdded() int64 {
	if m != nil {
		return m.LastAdded
	}
	return 0
}

func (m *Peer) GetLastDeleted() int64 {
	if m != nil {
		return m.LastDeleted
	}
	return 0
}

func (m *Peer) GetMemoryUsage() *MemoryUsage {
	if m != nil {
		return m.MemoryUsage
	}
	return nil
}

func (m *Peer) GetComputeUsage() *ComputeUsage {
	if m != nil {
		return m.ComputeUsage
	}
	return nil
}

func (m *Peer) GetRuntime() string {
	if m != nil {
		return m.Runtime
	}
	return ""
}

func (m *Peer) GetServices() []string {
	if m != nil {
		return m.Services
	}
	return nil
}

func (m *Peer) GetStarted() int64 {
	if m != nil {
		return m.Started
	}
	return 0
}

type ComputeUsage struct {
	Cores      int64 `protobuf:"varint,1,opt,name=Cores" json:"Cores,omitempty"`
	Goroutines int64 `protobuf:"varint,2,opt,name=Goroutines" json:"Goroutines,omitempty"`
}

func (m *ComputeUsage) Reset()                    { *m = ComputeUsage{} }
func (m *ComputeUsage) String() string            { return proto.CompactTextString(m) }
func (*ComputeUsage) ProtoMessage()               {}
func (*ComputeUsage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *ComputeUsage) GetCores() int64 {
	if m != nil {
		return m.Cores
	}
	return 0
}

func (m *ComputeUsage) GetGoroutines() int64 {
	if m != nil {
		return m.Goroutines
	}
	return 0
}

type MemoryUsage struct {
	Alloc      uint64 `protobuf:"varint,1,opt,name=Alloc" json:"Alloc,omitempty"`
	TotalAlloc uint64 `protobuf:"varint,2,opt,name=TotalAlloc" json:"TotalAlloc,omitempty"`
	Sys        uint64 `protobuf:"varint,3,opt,name=Sys" json:"Sys,omitempty"`
	NumGC      uint32 `protobuf:"varint,4,opt,name=NumGC" json:"NumGC,omitempty"`
}

func (m *MemoryUsage) Reset()                    { *m = MemoryUsage{} }
func (m *MemoryUsage) String() string            { return proto.CompactTextString(m) }
func (*MemoryUsage) ProtoMessage()               {}
func (*MemoryUsage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *MemoryUsage) GetAlloc() uint64 {
	if m != nil {
		return m.Alloc
	}
	return 0
}

func (m *MemoryUsage) GetTotalAlloc() uint64 {
	if m != nil {
		return m.TotalAlloc
	}
	return 0
}

func (m *MemoryUsage) GetSys() uint64 {
	if m != nil {
		return m.Sys
	}
	return 0
}

func (m *MemoryUsage) GetNumGC() uint32 {
	if m != nil {
		return m.NumGC
	}
	return 0
}

type PeerList struct {
	Peers []*Peer `protobuf:"bytes,1,rep,name=Peers" json:"Peers,omitempty"`
}

func (m *PeerList) Reset()                    { *m = PeerList{} }
func (m *PeerList) String() string            { return proto.CompactTextString(m) }
func (*PeerList) ProtoMessage()               {}
func (*PeerList) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *PeerList) GetPeers() []*Peer {
	if m != nil {
		return m.Peers
	}
	return nil
}

func init() {
	proto.RegisterType((*Peer)(nil), "peers.Peer")
	proto.RegisterType((*ComputeUsage)(nil), "peers.ComputeUsage")
	proto.RegisterType((*MemoryUsage)(nil), "peers.MemoryUsage")
	proto.RegisterType((*PeerList)(nil), "peers.PeerList")
}

func init() { proto.RegisterFile("types.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 360 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x92, 0xdd, 0x4a, 0xfb, 0x40,
	0x10, 0xc5, 0xc9, 0x57, 0x3f, 0x26, 0xff, 0xbf, 0xc8, 0x2a, 0xb2, 0x88, 0x48, 0xcc, 0x55, 0x6e,
	0xec, 0x45, 0x15, 0xbc, 0x2e, 0x0d, 0xd4, 0x42, 0x2b, 0xb2, 0xd5, 0x07, 0x88, 0xcd, 0xa0, 0xc1,
	0xa4, 0x1b, 0x76, 0x37, 0x42, 0x9f, 0xd7, 0x17, 0x91, 0x9d, 0xa4, 0x9a, 0xde, 0xed, 0xef, 0x9c,
	0x9d, 0x61, 0xcf, 0xec, 0x40, 0x68, 0xf6, 0x35, 0xea, 0x49, 0xad, 0xa4, 0x91, 0x2c, 0xa8, 0x11,
	0x95, 0x8e, 0xbf, 0x5d, 0xf0, 0x9f, 0x11, 0x15, 0x3b, 0x01, 0x77, 0x99, 0x72, 0x27, 0x72, 0x92,
	0xb1, 0x70, 0x97, 0x29, 0xbb, 0x80, 0xc1, 0x1a, 0xf5, 0xc7, 0x32, 0xe5, 0x2e, 0x69, 0x1d, 0xb1,
	0x4b, 0x18, 0x3d, 0x4a, 0x6d, 0x76, 0x59, 0x85, 0xdc, 0x23, 0xe7, 0x97, 0x19, 0x87, 0xe1, 0x2c,
	0xcf, 0x15, 0x6a, 0xcd, 0x7d, 0xb2, 0x0e, 0xc8, 0xae, 0x60, 0xbc, 0xca, 0xb4, 0x99, 0xe5, 0x39,
	0xe6, 0x3c, 0x88, 0x9c, 0xc4, 0x13, 0x7f, 0x02, 0x8b, 0x20, 0xb4, 0x90, 0x62, 0x89, 0x06, 0x73,
	0x3e, 0x20, 0xbf, 0x2f, 0xb1, 0x7b, 0x08, 0xd7, 0x58, 0x49, 0xb5, 0x7f, 0xd5, 0xd9, 0x3b, 0xf2,
	0x61, 0xe4, 0x24, 0xe1, 0x94, 0x4d, 0x28, 0xc3, 0xa4, 0xe7, 0x88, 0xfe, 0x35, 0xf6, 0x00, 0xff,
	0xe6, 0xb2, 0xaa, 0x1b, 0x83, 0x6d, 0xd9, 0x88, 0xca, 0xce, 0xba, 0xb2, 0xbe, 0x25, 0x8e, 0x2e,
	0xda, 0x20, 0xa2, 0xd9, 0x99, 0xa2, 0x42, 0x3e, 0x6e, 0x83, 0x74, 0x68, 0xe3, 0x6f, 0x50, 0x7d,
	0x15, 0x5b, 0xd4, 0x1c, 0x22, 0xcf, 0xc6, 0x3f, 0xb0, 0xad, 0xda, 0x98, 0x4c, 0xd9, 0x08, 0x21,
	0x45, 0x38, 0x60, 0x9c, 0x1e, 0x3f, 0x84, 0x9d, 0x43, 0x30, 0x97, 0x0a, 0x35, 0xcd, 0xdb, 0x13,
	0x2d, 0xb0, 0x6b, 0x80, 0x85, 0x54, 0xb2, 0x31, 0xc5, 0x0e, 0x35, 0x8d, 0xdd, 0x13, 0x3d, 0x25,
	0xfe, 0x3c, 0x1a, 0x82, 0x6d, 0x32, 0x2b, 0x4b, 0xb9, 0xa5, 0x26, 0xbe, 0x68, 0xc1, 0x36, 0x79,
	0x91, 0x26, 0x2b, 0x5b, 0xcb, 0x25, 0xab, 0xa7, 0xb0, 0x53, 0xf0, 0x36, 0x7b, 0x4d, 0x5f, 0xe7,
	0x0b, 0x7b, 0xb4, 0x7d, 0x9e, 0x9a, 0x6a, 0x31, 0xa7, 0x3f, 0xfb, 0x2f, 0x5a, 0x88, 0x6f, 0x61,
	0x64, 0xf7, 0x62, 0x55, 0x68, 0xc3, 0x6e, 0x20, 0xb0, 0x67, 0xfb, 0x5c, 0x2f, 0x09, 0xa7, 0x61,
	0x37, 0x40, 0xab, 0x89, 0xd6, 0x79, 0x1b, 0xd0, 0x56, 0xdd, 0xfd, 0x04, 0x00, 0x00, 0xff, 0xff,
	0x59, 0xe7, 0xf6, 0xa1, 0x64, 0x02, 0x00, 0x00,
}
