// Code generated by protoc-gen-go. DO NOT EDIT.
// source: pb.proto

package pb

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

type ServiceTag struct {
	Key                  string   `protobuf:"bytes,1,opt,name=Key,proto3" json:"Key,omitempty"`
	Value                string   `protobuf:"bytes,2,opt,name=Value,proto3" json:"Value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ServiceTag) Reset()         { *m = ServiceTag{} }
func (m *ServiceTag) String() string { return proto.CompactTextString(m) }
func (*ServiceTag) ProtoMessage()    {}
func (*ServiceTag) Descriptor() ([]byte, []int) {
	return fileDescriptor_f80abaa17e25ccc8, []int{0}
}

func (m *ServiceTag) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ServiceTag.Unmarshal(m, b)
}
func (m *ServiceTag) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ServiceTag.Marshal(b, m, deterministic)
}
func (m *ServiceTag) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ServiceTag.Merge(m, src)
}
func (m *ServiceTag) XXX_Size() int {
	return xxx_messageInfo_ServiceTag.Size(m)
}
func (m *ServiceTag) XXX_DiscardUnknown() {
	xxx_messageInfo_ServiceTag.DiscardUnknown(m)
}

var xxx_messageInfo_ServiceTag proto.InternalMessageInfo

func (m *ServiceTag) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *ServiceTag) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

type NodeService struct {
	ID                   string        `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	Name                 string        `protobuf:"bytes,2,opt,name=Name,proto3" json:"Name,omitempty"`
	NetworkAddress       string        `protobuf:"bytes,3,opt,name=NetworkAddress,proto3" json:"NetworkAddress,omitempty"`
	Peer                 string        `protobuf:"bytes,4,opt,name=Peer,proto3" json:"Peer,omitempty"`
	Tags                 []*ServiceTag `protobuf:"bytes,5,rep,name=Tags,proto3" json:"Tags,omitempty"`
	Health               string        `protobuf:"bytes,6,opt,name=Health,proto3" json:"Health,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *NodeService) Reset()         { *m = NodeService{} }
func (m *NodeService) String() string { return proto.CompactTextString(m) }
func (*NodeService) ProtoMessage()    {}
func (*NodeService) Descriptor() ([]byte, []int) {
	return fileDescriptor_f80abaa17e25ccc8, []int{1}
}

func (m *NodeService) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NodeService.Unmarshal(m, b)
}
func (m *NodeService) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NodeService.Marshal(b, m, deterministic)
}
func (m *NodeService) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NodeService.Merge(m, src)
}
func (m *NodeService) XXX_Size() int {
	return xxx_messageInfo_NodeService.Size(m)
}
func (m *NodeService) XXX_DiscardUnknown() {
	xxx_messageInfo_NodeService.DiscardUnknown(m)
}

var xxx_messageInfo_NodeService proto.InternalMessageInfo

func (m *NodeService) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *NodeService) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *NodeService) GetNetworkAddress() string {
	if m != nil {
		return m.NetworkAddress
	}
	return ""
}

func (m *NodeService) GetPeer() string {
	if m != nil {
		return m.Peer
	}
	return ""
}

func (m *NodeService) GetTags() []*ServiceTag {
	if m != nil {
		return m.Tags
	}
	return nil
}

func (m *NodeService) GetHealth() string {
	if m != nil {
		return m.Health
	}
	return ""
}

type Peer struct {
	ID                   string         `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	Hostname             string         `protobuf:"bytes,3,opt,name=Hostname,proto3" json:"Hostname,omitempty"`
	MemoryUsage          *MemoryUsage   `protobuf:"bytes,7,opt,name=MemoryUsage,proto3" json:"MemoryUsage,omitempty"`
	ComputeUsage         *ComputeUsage  `protobuf:"bytes,8,opt,name=ComputeUsage,proto3" json:"ComputeUsage,omitempty"`
	Runtime              string         `protobuf:"bytes,9,opt,name=Runtime,proto3" json:"Runtime,omitempty"`
	HostedServices       []*NodeService `protobuf:"bytes,10,rep,name=HostedServices,proto3" json:"HostedServices,omitempty"`
	Services             []string       `protobuf:"bytes,11,rep,name=Services,proto3" json:"Services,omitempty"`
	Started              int64          `protobuf:"varint,12,opt,name=Started,proto3" json:"Started,omitempty"`
	LastContact          int64          `protobuf:"varint,13,opt,name=LastContact,proto3" json:"LastContact,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *Peer) Reset()         { *m = Peer{} }
func (m *Peer) String() string { return proto.CompactTextString(m) }
func (*Peer) ProtoMessage()    {}
func (*Peer) Descriptor() ([]byte, []int) {
	return fileDescriptor_f80abaa17e25ccc8, []int{2}
}

func (m *Peer) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Peer.Unmarshal(m, b)
}
func (m *Peer) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Peer.Marshal(b, m, deterministic)
}
func (m *Peer) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Peer.Merge(m, src)
}
func (m *Peer) XXX_Size() int {
	return xxx_messageInfo_Peer.Size(m)
}
func (m *Peer) XXX_DiscardUnknown() {
	xxx_messageInfo_Peer.DiscardUnknown(m)
}

var xxx_messageInfo_Peer proto.InternalMessageInfo

func (m *Peer) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *Peer) GetHostname() string {
	if m != nil {
		return m.Hostname
	}
	return ""
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

func (m *Peer) GetHostedServices() []*NodeService {
	if m != nil {
		return m.HostedServices
	}
	return nil
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

func (m *Peer) GetLastContact() int64 {
	if m != nil {
		return m.LastContact
	}
	return 0
}

type ComputeUsage struct {
	Cores                int64    `protobuf:"varint,1,opt,name=Cores,proto3" json:"Cores,omitempty"`
	Goroutines           int64    `protobuf:"varint,2,opt,name=Goroutines,proto3" json:"Goroutines,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ComputeUsage) Reset()         { *m = ComputeUsage{} }
func (m *ComputeUsage) String() string { return proto.CompactTextString(m) }
func (*ComputeUsage) ProtoMessage()    {}
func (*ComputeUsage) Descriptor() ([]byte, []int) {
	return fileDescriptor_f80abaa17e25ccc8, []int{3}
}

func (m *ComputeUsage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ComputeUsage.Unmarshal(m, b)
}
func (m *ComputeUsage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ComputeUsage.Marshal(b, m, deterministic)
}
func (m *ComputeUsage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ComputeUsage.Merge(m, src)
}
func (m *ComputeUsage) XXX_Size() int {
	return xxx_messageInfo_ComputeUsage.Size(m)
}
func (m *ComputeUsage) XXX_DiscardUnknown() {
	xxx_messageInfo_ComputeUsage.DiscardUnknown(m)
}

var xxx_messageInfo_ComputeUsage proto.InternalMessageInfo

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
	Alloc                uint64   `protobuf:"varint,1,opt,name=Alloc,proto3" json:"Alloc,omitempty"`
	TotalAlloc           uint64   `protobuf:"varint,2,opt,name=TotalAlloc,proto3" json:"TotalAlloc,omitempty"`
	Sys                  uint64   `protobuf:"varint,3,opt,name=Sys,proto3" json:"Sys,omitempty"`
	NumGC                uint32   `protobuf:"varint,4,opt,name=NumGC,proto3" json:"NumGC,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MemoryUsage) Reset()         { *m = MemoryUsage{} }
func (m *MemoryUsage) String() string { return proto.CompactTextString(m) }
func (*MemoryUsage) ProtoMessage()    {}
func (*MemoryUsage) Descriptor() ([]byte, []int) {
	return fileDescriptor_f80abaa17e25ccc8, []int{4}
}

func (m *MemoryUsage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MemoryUsage.Unmarshal(m, b)
}
func (m *MemoryUsage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MemoryUsage.Marshal(b, m, deterministic)
}
func (m *MemoryUsage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MemoryUsage.Merge(m, src)
}
func (m *MemoryUsage) XXX_Size() int {
	return xxx_messageInfo_MemoryUsage.Size(m)
}
func (m *MemoryUsage) XXX_DiscardUnknown() {
	xxx_messageInfo_MemoryUsage.DiscardUnknown(m)
}

var xxx_messageInfo_MemoryUsage proto.InternalMessageInfo

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
	Peers                []*Peer  `protobuf:"bytes,1,rep,name=Peers,proto3" json:"Peers,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PeerList) Reset()         { *m = PeerList{} }
func (m *PeerList) String() string { return proto.CompactTextString(m) }
func (*PeerList) ProtoMessage()    {}
func (*PeerList) Descriptor() ([]byte, []int) {
	return fileDescriptor_f80abaa17e25ccc8, []int{5}
}

func (m *PeerList) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PeerList.Unmarshal(m, b)
}
func (m *PeerList) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PeerList.Marshal(b, m, deterministic)
}
func (m *PeerList) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PeerList.Merge(m, src)
}
func (m *PeerList) XXX_Size() int {
	return xxx_messageInfo_PeerList.Size(m)
}
func (m *PeerList) XXX_DiscardUnknown() {
	xxx_messageInfo_PeerList.DiscardUnknown(m)
}

var xxx_messageInfo_PeerList proto.InternalMessageInfo

func (m *PeerList) GetPeers() []*Peer {
	if m != nil {
		return m.Peers
	}
	return nil
}

type GetEndpointsInput struct {
	ServiceName          string   `protobuf:"bytes,1,opt,name=ServiceName,proto3" json:"ServiceName,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetEndpointsInput) Reset()         { *m = GetEndpointsInput{} }
func (m *GetEndpointsInput) String() string { return proto.CompactTextString(m) }
func (*GetEndpointsInput) ProtoMessage()    {}
func (*GetEndpointsInput) Descriptor() ([]byte, []int) {
	return fileDescriptor_f80abaa17e25ccc8, []int{6}
}

func (m *GetEndpointsInput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetEndpointsInput.Unmarshal(m, b)
}
func (m *GetEndpointsInput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetEndpointsInput.Marshal(b, m, deterministic)
}
func (m *GetEndpointsInput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetEndpointsInput.Merge(m, src)
}
func (m *GetEndpointsInput) XXX_Size() int {
	return xxx_messageInfo_GetEndpointsInput.Size(m)
}
func (m *GetEndpointsInput) XXX_DiscardUnknown() {
	xxx_messageInfo_GetEndpointsInput.DiscardUnknown(m)
}

var xxx_messageInfo_GetEndpointsInput proto.InternalMessageInfo

func (m *GetEndpointsInput) GetServiceName() string {
	if m != nil {
		return m.ServiceName
	}
	return ""
}

type GetEndpointsOutput struct {
	NodeServices         []*NodeService `protobuf:"bytes,1,rep,name=NodeServices,proto3" json:"NodeServices,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *GetEndpointsOutput) Reset()         { *m = GetEndpointsOutput{} }
func (m *GetEndpointsOutput) String() string { return proto.CompactTextString(m) }
func (*GetEndpointsOutput) ProtoMessage()    {}
func (*GetEndpointsOutput) Descriptor() ([]byte, []int) {
	return fileDescriptor_f80abaa17e25ccc8, []int{7}
}

func (m *GetEndpointsOutput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetEndpointsOutput.Unmarshal(m, b)
}
func (m *GetEndpointsOutput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetEndpointsOutput.Marshal(b, m, deterministic)
}
func (m *GetEndpointsOutput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetEndpointsOutput.Merge(m, src)
}
func (m *GetEndpointsOutput) XXX_Size() int {
	return xxx_messageInfo_GetEndpointsOutput.Size(m)
}
func (m *GetEndpointsOutput) XXX_DiscardUnknown() {
	xxx_messageInfo_GetEndpointsOutput.DiscardUnknown(m)
}

var xxx_messageInfo_GetEndpointsOutput proto.InternalMessageInfo

func (m *GetEndpointsOutput) GetNodeServices() []*NodeService {
	if m != nil {
		return m.NodeServices
	}
	return nil
}

type RegisterServiceInput struct {
	ServiceID            string   `protobuf:"bytes,1,opt,name=ServiceID,proto3" json:"ServiceID,omitempty"`
	ServiceName          string   `protobuf:"bytes,2,opt,name=ServiceName,proto3" json:"ServiceName,omitempty"`
	NetworkAddress       string   `protobuf:"bytes,3,opt,name=NetworkAddress,proto3" json:"NetworkAddress,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RegisterServiceInput) Reset()         { *m = RegisterServiceInput{} }
func (m *RegisterServiceInput) String() string { return proto.CompactTextString(m) }
func (*RegisterServiceInput) ProtoMessage()    {}
func (*RegisterServiceInput) Descriptor() ([]byte, []int) {
	return fileDescriptor_f80abaa17e25ccc8, []int{8}
}

func (m *RegisterServiceInput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RegisterServiceInput.Unmarshal(m, b)
}
func (m *RegisterServiceInput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RegisterServiceInput.Marshal(b, m, deterministic)
}
func (m *RegisterServiceInput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RegisterServiceInput.Merge(m, src)
}
func (m *RegisterServiceInput) XXX_Size() int {
	return xxx_messageInfo_RegisterServiceInput.Size(m)
}
func (m *RegisterServiceInput) XXX_DiscardUnknown() {
	xxx_messageInfo_RegisterServiceInput.DiscardUnknown(m)
}

var xxx_messageInfo_RegisterServiceInput proto.InternalMessageInfo

func (m *RegisterServiceInput) GetServiceID() string {
	if m != nil {
		return m.ServiceID
	}
	return ""
}

func (m *RegisterServiceInput) GetServiceName() string {
	if m != nil {
		return m.ServiceName
	}
	return ""
}

func (m *RegisterServiceInput) GetNetworkAddress() string {
	if m != nil {
		return m.NetworkAddress
	}
	return ""
}

type RegisterServiceOutput struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RegisterServiceOutput) Reset()         { *m = RegisterServiceOutput{} }
func (m *RegisterServiceOutput) String() string { return proto.CompactTextString(m) }
func (*RegisterServiceOutput) ProtoMessage()    {}
func (*RegisterServiceOutput) Descriptor() ([]byte, []int) {
	return fileDescriptor_f80abaa17e25ccc8, []int{9}
}

func (m *RegisterServiceOutput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RegisterServiceOutput.Unmarshal(m, b)
}
func (m *RegisterServiceOutput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RegisterServiceOutput.Marshal(b, m, deterministic)
}
func (m *RegisterServiceOutput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RegisterServiceOutput.Merge(m, src)
}
func (m *RegisterServiceOutput) XXX_Size() int {
	return xxx_messageInfo_RegisterServiceOutput.Size(m)
}
func (m *RegisterServiceOutput) XXX_DiscardUnknown() {
	xxx_messageInfo_RegisterServiceOutput.DiscardUnknown(m)
}

var xxx_messageInfo_RegisterServiceOutput proto.InternalMessageInfo

type UnregisterServiceInput struct {
	ServiceID            string   `protobuf:"bytes,1,opt,name=ServiceID,proto3" json:"ServiceID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *UnregisterServiceInput) Reset()         { *m = UnregisterServiceInput{} }
func (m *UnregisterServiceInput) String() string { return proto.CompactTextString(m) }
func (*UnregisterServiceInput) ProtoMessage()    {}
func (*UnregisterServiceInput) Descriptor() ([]byte, []int) {
	return fileDescriptor_f80abaa17e25ccc8, []int{10}
}

func (m *UnregisterServiceInput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UnregisterServiceInput.Unmarshal(m, b)
}
func (m *UnregisterServiceInput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UnregisterServiceInput.Marshal(b, m, deterministic)
}
func (m *UnregisterServiceInput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UnregisterServiceInput.Merge(m, src)
}
func (m *UnregisterServiceInput) XXX_Size() int {
	return xxx_messageInfo_UnregisterServiceInput.Size(m)
}
func (m *UnregisterServiceInput) XXX_DiscardUnknown() {
	xxx_messageInfo_UnregisterServiceInput.DiscardUnknown(m)
}

var xxx_messageInfo_UnregisterServiceInput proto.InternalMessageInfo

func (m *UnregisterServiceInput) GetServiceID() string {
	if m != nil {
		return m.ServiceID
	}
	return ""
}

type UnregisterServiceOutput struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *UnregisterServiceOutput) Reset()         { *m = UnregisterServiceOutput{} }
func (m *UnregisterServiceOutput) String() string { return proto.CompactTextString(m) }
func (*UnregisterServiceOutput) ProtoMessage()    {}
func (*UnregisterServiceOutput) Descriptor() ([]byte, []int) {
	return fileDescriptor_f80abaa17e25ccc8, []int{11}
}

func (m *UnregisterServiceOutput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UnregisterServiceOutput.Unmarshal(m, b)
}
func (m *UnregisterServiceOutput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UnregisterServiceOutput.Marshal(b, m, deterministic)
}
func (m *UnregisterServiceOutput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UnregisterServiceOutput.Merge(m, src)
}
func (m *UnregisterServiceOutput) XXX_Size() int {
	return xxx_messageInfo_UnregisterServiceOutput.Size(m)
}
func (m *UnregisterServiceOutput) XXX_DiscardUnknown() {
	xxx_messageInfo_UnregisterServiceOutput.DiscardUnknown(m)
}

var xxx_messageInfo_UnregisterServiceOutput proto.InternalMessageInfo

type AddServiceTagInput struct {
	ServiceID            string   `protobuf:"bytes,1,opt,name=ServiceID,proto3" json:"ServiceID,omitempty"`
	TagKey               string   `protobuf:"bytes,2,opt,name=TagKey,proto3" json:"TagKey,omitempty"`
	TagValue             string   `protobuf:"bytes,3,opt,name=TagValue,proto3" json:"TagValue,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AddServiceTagInput) Reset()         { *m = AddServiceTagInput{} }
func (m *AddServiceTagInput) String() string { return proto.CompactTextString(m) }
func (*AddServiceTagInput) ProtoMessage()    {}
func (*AddServiceTagInput) Descriptor() ([]byte, []int) {
	return fileDescriptor_f80abaa17e25ccc8, []int{12}
}

func (m *AddServiceTagInput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AddServiceTagInput.Unmarshal(m, b)
}
func (m *AddServiceTagInput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AddServiceTagInput.Marshal(b, m, deterministic)
}
func (m *AddServiceTagInput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AddServiceTagInput.Merge(m, src)
}
func (m *AddServiceTagInput) XXX_Size() int {
	return xxx_messageInfo_AddServiceTagInput.Size(m)
}
func (m *AddServiceTagInput) XXX_DiscardUnknown() {
	xxx_messageInfo_AddServiceTagInput.DiscardUnknown(m)
}

var xxx_messageInfo_AddServiceTagInput proto.InternalMessageInfo

func (m *AddServiceTagInput) GetServiceID() string {
	if m != nil {
		return m.ServiceID
	}
	return ""
}

func (m *AddServiceTagInput) GetTagKey() string {
	if m != nil {
		return m.TagKey
	}
	return ""
}

func (m *AddServiceTagInput) GetTagValue() string {
	if m != nil {
		return m.TagValue
	}
	return ""
}

type AddServiceTagOutput struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AddServiceTagOutput) Reset()         { *m = AddServiceTagOutput{} }
func (m *AddServiceTagOutput) String() string { return proto.CompactTextString(m) }
func (*AddServiceTagOutput) ProtoMessage()    {}
func (*AddServiceTagOutput) Descriptor() ([]byte, []int) {
	return fileDescriptor_f80abaa17e25ccc8, []int{13}
}

func (m *AddServiceTagOutput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AddServiceTagOutput.Unmarshal(m, b)
}
func (m *AddServiceTagOutput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AddServiceTagOutput.Marshal(b, m, deterministic)
}
func (m *AddServiceTagOutput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AddServiceTagOutput.Merge(m, src)
}
func (m *AddServiceTagOutput) XXX_Size() int {
	return xxx_messageInfo_AddServiceTagOutput.Size(m)
}
func (m *AddServiceTagOutput) XXX_DiscardUnknown() {
	xxx_messageInfo_AddServiceTagOutput.DiscardUnknown(m)
}

var xxx_messageInfo_AddServiceTagOutput proto.InternalMessageInfo

type RemoveServiceTagInput struct {
	ServiceID            string   `protobuf:"bytes,1,opt,name=ServiceID,proto3" json:"ServiceID,omitempty"`
	TagKey               string   `protobuf:"bytes,2,opt,name=TagKey,proto3" json:"TagKey,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RemoveServiceTagInput) Reset()         { *m = RemoveServiceTagInput{} }
func (m *RemoveServiceTagInput) String() string { return proto.CompactTextString(m) }
func (*RemoveServiceTagInput) ProtoMessage()    {}
func (*RemoveServiceTagInput) Descriptor() ([]byte, []int) {
	return fileDescriptor_f80abaa17e25ccc8, []int{14}
}

func (m *RemoveServiceTagInput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RemoveServiceTagInput.Unmarshal(m, b)
}
func (m *RemoveServiceTagInput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RemoveServiceTagInput.Marshal(b, m, deterministic)
}
func (m *RemoveServiceTagInput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RemoveServiceTagInput.Merge(m, src)
}
func (m *RemoveServiceTagInput) XXX_Size() int {
	return xxx_messageInfo_RemoveServiceTagInput.Size(m)
}
func (m *RemoveServiceTagInput) XXX_DiscardUnknown() {
	xxx_messageInfo_RemoveServiceTagInput.DiscardUnknown(m)
}

var xxx_messageInfo_RemoveServiceTagInput proto.InternalMessageInfo

func (m *RemoveServiceTagInput) GetServiceID() string {
	if m != nil {
		return m.ServiceID
	}
	return ""
}

func (m *RemoveServiceTagInput) GetTagKey() string {
	if m != nil {
		return m.TagKey
	}
	return ""
}

type RemoveServiceTagOutput struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RemoveServiceTagOutput) Reset()         { *m = RemoveServiceTagOutput{} }
func (m *RemoveServiceTagOutput) String() string { return proto.CompactTextString(m) }
func (*RemoveServiceTagOutput) ProtoMessage()    {}
func (*RemoveServiceTagOutput) Descriptor() ([]byte, []int) {
	return fileDescriptor_f80abaa17e25ccc8, []int{15}
}

func (m *RemoveServiceTagOutput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RemoveServiceTagOutput.Unmarshal(m, b)
}
func (m *RemoveServiceTagOutput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RemoveServiceTagOutput.Marshal(b, m, deterministic)
}
func (m *RemoveServiceTagOutput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RemoveServiceTagOutput.Merge(m, src)
}
func (m *RemoveServiceTagOutput) XXX_Size() int {
	return xxx_messageInfo_RemoveServiceTagOutput.Size(m)
}
func (m *RemoveServiceTagOutput) XXX_DiscardUnknown() {
	xxx_messageInfo_RemoveServiceTagOutput.DiscardUnknown(m)
}

var xxx_messageInfo_RemoveServiceTagOutput proto.InternalMessageInfo

func init() {
	proto.RegisterType((*ServiceTag)(nil), "pb.ServiceTag")
	proto.RegisterType((*NodeService)(nil), "pb.NodeService")
	proto.RegisterType((*Peer)(nil), "pb.Peer")
	proto.RegisterType((*ComputeUsage)(nil), "pb.ComputeUsage")
	proto.RegisterType((*MemoryUsage)(nil), "pb.MemoryUsage")
	proto.RegisterType((*PeerList)(nil), "pb.PeerList")
	proto.RegisterType((*GetEndpointsInput)(nil), "pb.GetEndpointsInput")
	proto.RegisterType((*GetEndpointsOutput)(nil), "pb.GetEndpointsOutput")
	proto.RegisterType((*RegisterServiceInput)(nil), "pb.RegisterServiceInput")
	proto.RegisterType((*RegisterServiceOutput)(nil), "pb.RegisterServiceOutput")
	proto.RegisterType((*UnregisterServiceInput)(nil), "pb.UnregisterServiceInput")
	proto.RegisterType((*UnregisterServiceOutput)(nil), "pb.UnregisterServiceOutput")
	proto.RegisterType((*AddServiceTagInput)(nil), "pb.AddServiceTagInput")
	proto.RegisterType((*AddServiceTagOutput)(nil), "pb.AddServiceTagOutput")
	proto.RegisterType((*RemoveServiceTagInput)(nil), "pb.RemoveServiceTagInput")
	proto.RegisterType((*RemoveServiceTagOutput)(nil), "pb.RemoveServiceTagOutput")
}

func init() { proto.RegisterFile("pb.proto", fileDescriptor_f80abaa17e25ccc8) }

var fileDescriptor_f80abaa17e25ccc8 = []byte{
	// 682 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x55, 0x41, 0x6f, 0xd3, 0x4c,
	0x10, 0x55, 0xec, 0x34, 0x4d, 0x27, 0x69, 0x9a, 0x6f, 0xbf, 0x36, 0xdd, 0x46, 0xa8, 0x8a, 0xf6,
	0x80, 0x22, 0x0e, 0x15, 0xb4, 0x05, 0x4e, 0x48, 0x94, 0x04, 0xb5, 0x11, 0x6d, 0x40, 0x9b, 0x94,
	0xbb, 0x13, 0x2f, 0xc1, 0x6a, 0xec, 0xb5, 0xec, 0x75, 0x51, 0x2e, 0xfc, 0x18, 0x6e, 0xfc, 0x11,
	0x7e, 0x17, 0x9a, 0x5d, 0xbb, 0x75, 0xec, 0x1e, 0x0a, 0xe2, 0xb6, 0xf3, 0x66, 0xdf, 0xec, 0x9b,
	0xd9, 0xe7, 0x35, 0xd4, 0xc3, 0xd9, 0x51, 0x18, 0x49, 0x25, 0x89, 0x15, 0xce, 0xd8, 0x29, 0xc0,
	0x44, 0x44, 0xb7, 0xde, 0x5c, 0x4c, 0x9d, 0x05, 0x69, 0x83, 0xfd, 0x41, 0xac, 0x68, 0xa5, 0x57,
	0xe9, 0x6f, 0x71, 0x5c, 0x92, 0x5d, 0xd8, 0xf8, 0xec, 0x2c, 0x13, 0x41, 0x2d, 0x8d, 0x99, 0x80,
	0xfd, 0xac, 0x40, 0x63, 0x2c, 0x5d, 0x91, 0x52, 0x49, 0x0b, 0xac, 0xd1, 0x30, 0xa5, 0x59, 0xa3,
	0x21, 0x21, 0x50, 0x1d, 0x3b, 0x7e, 0x46, 0xd2, 0x6b, 0xf2, 0x14, 0x5a, 0x63, 0xa1, 0xbe, 0xc9,
	0xe8, 0xe6, 0xcc, 0x75, 0x23, 0x11, 0xc7, 0xd4, 0xd6, 0xd9, 0x02, 0x8a, 0xdc, 0x4f, 0x42, 0x44,
	0xb4, 0x6a, 0xb8, 0xb8, 0x26, 0x0c, 0xaa, 0x53, 0x67, 0x11, 0xd3, 0x8d, 0x9e, 0xdd, 0x6f, 0x1c,
	0xb7, 0x8e, 0xc2, 0xd9, 0xd1, 0xbd, 0x6a, 0xae, 0x73, 0xa4, 0x03, 0xb5, 0x0b, 0xe1, 0x2c, 0xd5,
	0x57, 0x5a, 0xd3, 0xcc, 0x34, 0x62, 0xbf, 0x2c, 0x53, 0xb0, 0x24, 0xb2, 0x0b, 0xf5, 0x0b, 0x19,
	0xab, 0x00, 0x85, 0x1a, 0x29, 0x77, 0x31, 0x79, 0x01, 0x8d, 0x2b, 0xe1, 0xcb, 0x68, 0x75, 0x1d,
	0x3b, 0x0b, 0x41, 0x37, 0x7b, 0x95, 0x7e, 0xe3, 0x78, 0x07, 0xcf, 0xcd, 0xc1, 0x3c, 0xbf, 0x87,
	0x9c, 0x42, 0x73, 0x20, 0xfd, 0x30, 0x51, 0xc2, 0x70, 0xea, 0x9a, 0xd3, 0x46, 0x4e, 0x1e, 0xe7,
	0x6b, 0xbb, 0x08, 0x85, 0x4d, 0x9e, 0x04, 0xca, 0xf3, 0x05, 0xdd, 0xd2, 0x1a, 0xb2, 0x90, 0xbc,
	0x86, 0x16, 0xca, 0x11, 0x6e, 0xda, 0x69, 0x4c, 0x41, 0x77, 0xaf, 0x55, 0xe4, 0x86, 0xcf, 0x0b,
	0xdb, 0xb0, 0xaf, 0x3b, 0x4a, 0xa3, 0x67, 0x63, 0x5f, 0x77, 0x39, 0x0a, 0x9b, 0x13, 0xe5, 0x44,
	0x4a, 0xb8, 0xb4, 0xd9, 0xab, 0xf4, 0x6d, 0x9e, 0x85, 0xa4, 0x07, 0x8d, 0x4b, 0x27, 0x56, 0x03,
	0x19, 0x28, 0x67, 0xae, 0xe8, 0xb6, 0xce, 0xe6, 0x21, 0x36, 0x5c, 0x6f, 0x10, 0xad, 0x31, 0x90,
	0x91, 0x88, 0xf5, 0x48, 0x6d, 0x6e, 0x02, 0x72, 0x08, 0x70, 0x2e, 0x23, 0x99, 0x28, 0x2f, 0x10,
	0xb1, 0x36, 0x80, 0xcd, 0x73, 0x08, 0xbb, 0x59, 0x9b, 0x2c, 0x16, 0x39, 0x5b, 0x2e, 0xe5, 0x5c,
	0x17, 0xa9, 0x72, 0x13, 0x60, 0x91, 0xa9, 0x54, 0xce, 0xd2, 0xa4, 0x2c, 0x9d, 0xca, 0x21, 0xe8,
	0xd3, 0xc9, 0xca, 0x18, 0xa8, 0xca, 0x71, 0x89, 0x75, 0xc6, 0x89, 0x7f, 0x3e, 0xd0, 0xb6, 0xd9,
	0xe6, 0x26, 0x60, 0xcf, 0xa0, 0x8e, 0x57, 0x7f, 0xe9, 0xc5, 0x8a, 0x1c, 0xc2, 0x06, 0xae, 0x51,
	0x2e, 0x8e, 0xb1, 0x8e, 0x63, 0x44, 0x80, 0x1b, 0x98, 0xbd, 0x84, 0xff, 0xce, 0x85, 0x7a, 0x1f,
	0xb8, 0xa1, 0xf4, 0x02, 0x15, 0x8f, 0x82, 0x30, 0x51, 0x38, 0x95, 0x74, 0x76, 0xda, 0xcf, 0xc6,
	0x3c, 0x79, 0x88, 0x8d, 0x80, 0xe4, 0x69, 0x1f, 0x13, 0x85, 0xbc, 0x13, 0x68, 0xe6, 0xae, 0x28,
	0x3b, 0xb3, 0x74, 0x75, 0x6b, 0x9b, 0xd8, 0x77, 0xd8, 0xe5, 0x62, 0xe1, 0xc5, 0x4a, 0x44, 0x29,
	0x66, 0x44, 0x3c, 0x81, 0xad, 0x2c, 0xce, 0xfc, 0x7b, 0x0f, 0x14, 0x25, 0x5a, 0x25, 0x89, 0x8f,
	0xfd, 0xf2, 0xd8, 0x3e, 0xec, 0x15, 0xce, 0x37, 0xdd, 0xb0, 0x57, 0xd0, 0xb9, 0x0e, 0xa2, 0x3f,
	0x96, 0xc6, 0x0e, 0x60, 0xbf, 0xc4, 0x4b, 0x4b, 0x7e, 0x01, 0x72, 0xe6, 0xba, 0xf7, 0x1f, 0xf1,
	0x63, 0x3a, 0xed, 0x40, 0x6d, 0xea, 0x2c, 0xf0, 0x81, 0x32, 0x4d, 0xa6, 0x11, 0x1a, 0x7e, 0xea,
	0x2c, 0xcc, 0x33, 0x95, 0x7e, 0xc8, 0x59, 0xcc, 0xf6, 0xe0, 0xff, 0xb5, 0x73, 0xd2, 0xe3, 0xaf,
	0xb0, 0x55, 0x5f, 0xde, 0x8a, 0x7f, 0xa2, 0x80, 0x51, 0xe8, 0x14, 0xcb, 0x99, 0x83, 0x8e, 0x7f,
	0x58, 0xd0, 0x1e, 0x7a, 0xf1, 0x5c, 0xde, 0x8a, 0x68, 0x95, 0x3d, 0x97, 0x6f, 0xa0, 0x99, 0xf7,
	0x0c, 0xd9, 0x43, 0x5f, 0x94, 0xcc, 0xd7, 0xed, 0x14, 0xe1, 0xd4, 0x5c, 0xef, 0x60, 0x67, 0xa2,
	0x22, 0xe1, 0xf8, 0x7f, 0x5b, 0xe1, 0x79, 0x85, 0xbc, 0x85, 0xed, 0xb5, 0xb9, 0x10, 0xbd, 0xb5,
	0x7c, 0x25, 0xdd, 0xfd, 0x12, 0x9e, 0xaa, 0x18, 0x41, 0xbb, 0xd8, 0x33, 0x39, 0xc0, 0xcd, 0x0f,
	0x0e, 0xb6, 0xdb, 0x7d, 0x28, 0x65, 0x4a, 0xcd, 0x6a, 0xfa, 0x7f, 0x74, 0xf2, 0x3b, 0x00, 0x00,
	0xff, 0xff, 0x4c, 0x9b, 0x45, 0xba, 0x9b, 0x06, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// DiscoveryServiceClient is the client API for DiscoveryService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type DiscoveryServiceClient interface {
	GetEndpoints(ctx context.Context, in *GetEndpointsInput, opts ...grpc.CallOption) (*GetEndpointsOutput, error)
	StreamEndpoints(ctx context.Context, in *GetEndpointsInput, opts ...grpc.CallOption) (DiscoveryService_StreamEndpointsClient, error)
	AddServiceTag(ctx context.Context, in *AddServiceTagInput, opts ...grpc.CallOption) (*AddServiceTagOutput, error)
	RemoveServiceTag(ctx context.Context, in *RemoveServiceTagInput, opts ...grpc.CallOption) (*RemoveServiceTagOutput, error)
}

type discoveryServiceClient struct {
	cc *grpc.ClientConn
}

func NewDiscoveryServiceClient(cc *grpc.ClientConn) DiscoveryServiceClient {
	return &discoveryServiceClient{cc}
}

func (c *discoveryServiceClient) GetEndpoints(ctx context.Context, in *GetEndpointsInput, opts ...grpc.CallOption) (*GetEndpointsOutput, error) {
	out := new(GetEndpointsOutput)
	err := c.cc.Invoke(ctx, "/pb.DiscoveryService/GetEndpoints", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *discoveryServiceClient) StreamEndpoints(ctx context.Context, in *GetEndpointsInput, opts ...grpc.CallOption) (DiscoveryService_StreamEndpointsClient, error) {
	stream, err := c.cc.NewStream(ctx, &_DiscoveryService_serviceDesc.Streams[0], "/pb.DiscoveryService/StreamEndpoints", opts...)
	if err != nil {
		return nil, err
	}
	x := &discoveryServiceStreamEndpointsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type DiscoveryService_StreamEndpointsClient interface {
	Recv() (*GetEndpointsOutput, error)
	grpc.ClientStream
}

type discoveryServiceStreamEndpointsClient struct {
	grpc.ClientStream
}

func (x *discoveryServiceStreamEndpointsClient) Recv() (*GetEndpointsOutput, error) {
	m := new(GetEndpointsOutput)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *discoveryServiceClient) AddServiceTag(ctx context.Context, in *AddServiceTagInput, opts ...grpc.CallOption) (*AddServiceTagOutput, error) {
	out := new(AddServiceTagOutput)
	err := c.cc.Invoke(ctx, "/pb.DiscoveryService/AddServiceTag", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *discoveryServiceClient) RemoveServiceTag(ctx context.Context, in *RemoveServiceTagInput, opts ...grpc.CallOption) (*RemoveServiceTagOutput, error) {
	out := new(RemoveServiceTagOutput)
	err := c.cc.Invoke(ctx, "/pb.DiscoveryService/RemoveServiceTag", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DiscoveryServiceServer is the server API for DiscoveryService service.
type DiscoveryServiceServer interface {
	GetEndpoints(context.Context, *GetEndpointsInput) (*GetEndpointsOutput, error)
	StreamEndpoints(*GetEndpointsInput, DiscoveryService_StreamEndpointsServer) error
	AddServiceTag(context.Context, *AddServiceTagInput) (*AddServiceTagOutput, error)
	RemoveServiceTag(context.Context, *RemoveServiceTagInput) (*RemoveServiceTagOutput, error)
}

// UnimplementedDiscoveryServiceServer can be embedded to have forward compatible implementations.
type UnimplementedDiscoveryServiceServer struct {
}

func (*UnimplementedDiscoveryServiceServer) GetEndpoints(ctx context.Context, req *GetEndpointsInput) (*GetEndpointsOutput, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetEndpoints not implemented")
}
func (*UnimplementedDiscoveryServiceServer) StreamEndpoints(req *GetEndpointsInput, srv DiscoveryService_StreamEndpointsServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamEndpoints not implemented")
}
func (*UnimplementedDiscoveryServiceServer) AddServiceTag(ctx context.Context, req *AddServiceTagInput) (*AddServiceTagOutput, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddServiceTag not implemented")
}
func (*UnimplementedDiscoveryServiceServer) RemoveServiceTag(ctx context.Context, req *RemoveServiceTagInput) (*RemoveServiceTagOutput, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveServiceTag not implemented")
}

func RegisterDiscoveryServiceServer(s *grpc.Server, srv DiscoveryServiceServer) {
	s.RegisterService(&_DiscoveryService_serviceDesc, srv)
}

func _DiscoveryService_GetEndpoints_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetEndpointsInput)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DiscoveryServiceServer).GetEndpoints(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.DiscoveryService/GetEndpoints",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DiscoveryServiceServer).GetEndpoints(ctx, req.(*GetEndpointsInput))
	}
	return interceptor(ctx, in, info, handler)
}

func _DiscoveryService_StreamEndpoints_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetEndpointsInput)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(DiscoveryServiceServer).StreamEndpoints(m, &discoveryServiceStreamEndpointsServer{stream})
}

type DiscoveryService_StreamEndpointsServer interface {
	Send(*GetEndpointsOutput) error
	grpc.ServerStream
}

type discoveryServiceStreamEndpointsServer struct {
	grpc.ServerStream
}

func (x *discoveryServiceStreamEndpointsServer) Send(m *GetEndpointsOutput) error {
	return x.ServerStream.SendMsg(m)
}

func _DiscoveryService_AddServiceTag_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddServiceTagInput)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DiscoveryServiceServer).AddServiceTag(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.DiscoveryService/AddServiceTag",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DiscoveryServiceServer).AddServiceTag(ctx, req.(*AddServiceTagInput))
	}
	return interceptor(ctx, in, info, handler)
}

func _DiscoveryService_RemoveServiceTag_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveServiceTagInput)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DiscoveryServiceServer).RemoveServiceTag(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.DiscoveryService/RemoveServiceTag",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DiscoveryServiceServer).RemoveServiceTag(ctx, req.(*RemoveServiceTagInput))
	}
	return interceptor(ctx, in, info, handler)
}

var _DiscoveryService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "pb.DiscoveryService",
	HandlerType: (*DiscoveryServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetEndpoints",
			Handler:    _DiscoveryService_GetEndpoints_Handler,
		},
		{
			MethodName: "AddServiceTag",
			Handler:    _DiscoveryService_AddServiceTag_Handler,
		},
		{
			MethodName: "RemoveServiceTag",
			Handler:    _DiscoveryService_RemoveServiceTag_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamEndpoints",
			Handler:       _DiscoveryService_StreamEndpoints_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "pb.proto",
}
