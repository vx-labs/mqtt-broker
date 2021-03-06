// Code generated by protoc-gen-go. DO NOT EDIT.
// source: events.proto

package events

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

type SessionCreated struct {
	ID                   string   `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	Tenant               string   `protobuf:"bytes,2,opt,name=Tenant,proto3" json:"Tenant,omitempty"`
	Peer                 string   `protobuf:"bytes,3,opt,name=Peer,proto3" json:"Peer,omitempty"`
	WillTopic            []byte   `protobuf:"bytes,4,opt,name=WillTopic,proto3" json:"WillTopic,omitempty"`
	WillQoS              int32    `protobuf:"varint,5,opt,name=WillQoS,proto3" json:"WillQoS,omitempty"`
	WillPayload          []byte   `protobuf:"bytes,6,opt,name=WillPayload,proto3" json:"WillPayload,omitempty"`
	WillRetain           bool     `protobuf:"varint,7,opt,name=WillRetain,proto3" json:"WillRetain,omitempty"`
	ClientID             string   `protobuf:"bytes,9,opt,name=ClientID,proto3" json:"ClientID,omitempty"`
	Transport            string   `protobuf:"bytes,10,opt,name=Transport,proto3" json:"Transport,omitempty"`
	RemoteAddress        string   `protobuf:"bytes,11,opt,name=RemoteAddress,proto3" json:"RemoteAddress,omitempty"`
	KeepaliveInterval    int32    `protobuf:"varint,12,opt,name=KeepaliveInterval,proto3" json:"KeepaliveInterval,omitempty"`
	Timestamp            int64    `protobuf:"varint,13,opt,name=Timestamp,proto3" json:"Timestamp,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SessionCreated) Reset()         { *m = SessionCreated{} }
func (m *SessionCreated) String() string { return proto.CompactTextString(m) }
func (*SessionCreated) ProtoMessage()    {}
func (*SessionCreated) Descriptor() ([]byte, []int) {
	return fileDescriptor_8f22242cb04491f9, []int{0}
}

func (m *SessionCreated) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SessionCreated.Unmarshal(m, b)
}
func (m *SessionCreated) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SessionCreated.Marshal(b, m, deterministic)
}
func (m *SessionCreated) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SessionCreated.Merge(m, src)
}
func (m *SessionCreated) XXX_Size() int {
	return xxx_messageInfo_SessionCreated.Size(m)
}
func (m *SessionCreated) XXX_DiscardUnknown() {
	xxx_messageInfo_SessionCreated.DiscardUnknown(m)
}

var xxx_messageInfo_SessionCreated proto.InternalMessageInfo

func (m *SessionCreated) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *SessionCreated) GetTenant() string {
	if m != nil {
		return m.Tenant
	}
	return ""
}

func (m *SessionCreated) GetPeer() string {
	if m != nil {
		return m.Peer
	}
	return ""
}

func (m *SessionCreated) GetWillTopic() []byte {
	if m != nil {
		return m.WillTopic
	}
	return nil
}

func (m *SessionCreated) GetWillQoS() int32 {
	if m != nil {
		return m.WillQoS
	}
	return 0
}

func (m *SessionCreated) GetWillPayload() []byte {
	if m != nil {
		return m.WillPayload
	}
	return nil
}

func (m *SessionCreated) GetWillRetain() bool {
	if m != nil {
		return m.WillRetain
	}
	return false
}

func (m *SessionCreated) GetClientID() string {
	if m != nil {
		return m.ClientID
	}
	return ""
}

func (m *SessionCreated) GetTransport() string {
	if m != nil {
		return m.Transport
	}
	return ""
}

func (m *SessionCreated) GetRemoteAddress() string {
	if m != nil {
		return m.RemoteAddress
	}
	return ""
}

func (m *SessionCreated) GetKeepaliveInterval() int32 {
	if m != nil {
		return m.KeepaliveInterval
	}
	return 0
}

func (m *SessionCreated) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

type SessionClosed struct {
	ID                   string   `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	Tenant               string   `protobuf:"bytes,2,opt,name=Tenant,proto3" json:"Tenant,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SessionClosed) Reset()         { *m = SessionClosed{} }
func (m *SessionClosed) String() string { return proto.CompactTextString(m) }
func (*SessionClosed) ProtoMessage()    {}
func (*SessionClosed) Descriptor() ([]byte, []int) {
	return fileDescriptor_8f22242cb04491f9, []int{1}
}

func (m *SessionClosed) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SessionClosed.Unmarshal(m, b)
}
func (m *SessionClosed) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SessionClosed.Marshal(b, m, deterministic)
}
func (m *SessionClosed) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SessionClosed.Merge(m, src)
}
func (m *SessionClosed) XXX_Size() int {
	return xxx_messageInfo_SessionClosed.Size(m)
}
func (m *SessionClosed) XXX_DiscardUnknown() {
	xxx_messageInfo_SessionClosed.DiscardUnknown(m)
}

var xxx_messageInfo_SessionClosed proto.InternalMessageInfo

func (m *SessionClosed) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *SessionClosed) GetTenant() string {
	if m != nil {
		return m.Tenant
	}
	return ""
}

type SessionLost struct {
	ID                   string   `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	Tenant               string   `protobuf:"bytes,2,opt,name=Tenant,proto3" json:"Tenant,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SessionLost) Reset()         { *m = SessionLost{} }
func (m *SessionLost) String() string { return proto.CompactTextString(m) }
func (*SessionLost) ProtoMessage()    {}
func (*SessionLost) Descriptor() ([]byte, []int) {
	return fileDescriptor_8f22242cb04491f9, []int{2}
}

func (m *SessionLost) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SessionLost.Unmarshal(m, b)
}
func (m *SessionLost) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SessionLost.Marshal(b, m, deterministic)
}
func (m *SessionLost) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SessionLost.Merge(m, src)
}
func (m *SessionLost) XXX_Size() int {
	return xxx_messageInfo_SessionLost.Size(m)
}
func (m *SessionLost) XXX_DiscardUnknown() {
	xxx_messageInfo_SessionLost.DiscardUnknown(m)
}

var xxx_messageInfo_SessionLost proto.InternalMessageInfo

func (m *SessionLost) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *SessionLost) GetTenant() string {
	if m != nil {
		return m.Tenant
	}
	return ""
}

type PeerLost struct {
	ID                   string   `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PeerLost) Reset()         { *m = PeerLost{} }
func (m *PeerLost) String() string { return proto.CompactTextString(m) }
func (*PeerLost) ProtoMessage()    {}
func (*PeerLost) Descriptor() ([]byte, []int) {
	return fileDescriptor_8f22242cb04491f9, []int{3}
}

func (m *PeerLost) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PeerLost.Unmarshal(m, b)
}
func (m *PeerLost) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PeerLost.Marshal(b, m, deterministic)
}
func (m *PeerLost) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PeerLost.Merge(m, src)
}
func (m *PeerLost) XXX_Size() int {
	return xxx_messageInfo_PeerLost.Size(m)
}
func (m *PeerLost) XXX_DiscardUnknown() {
	xxx_messageInfo_PeerLost.DiscardUnknown(m)
}

var xxx_messageInfo_PeerLost proto.InternalMessageInfo

func (m *PeerLost) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

type SessionSubscribed struct {
	SessionID            string   `protobuf:"bytes,1,opt,name=SessionID,proto3" json:"SessionID,omitempty"`
	Tenant               string   `protobuf:"bytes,2,opt,name=Tenant,proto3" json:"Tenant,omitempty"`
	Pattern              []byte   `protobuf:"bytes,3,opt,name=Pattern,proto3" json:"Pattern,omitempty"`
	Qos                  int32    `protobuf:"varint,4,opt,name=Qos,proto3" json:"Qos,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SessionSubscribed) Reset()         { *m = SessionSubscribed{} }
func (m *SessionSubscribed) String() string { return proto.CompactTextString(m) }
func (*SessionSubscribed) ProtoMessage()    {}
func (*SessionSubscribed) Descriptor() ([]byte, []int) {
	return fileDescriptor_8f22242cb04491f9, []int{4}
}

func (m *SessionSubscribed) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SessionSubscribed.Unmarshal(m, b)
}
func (m *SessionSubscribed) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SessionSubscribed.Marshal(b, m, deterministic)
}
func (m *SessionSubscribed) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SessionSubscribed.Merge(m, src)
}
func (m *SessionSubscribed) XXX_Size() int {
	return xxx_messageInfo_SessionSubscribed.Size(m)
}
func (m *SessionSubscribed) XXX_DiscardUnknown() {
	xxx_messageInfo_SessionSubscribed.DiscardUnknown(m)
}

var xxx_messageInfo_SessionSubscribed proto.InternalMessageInfo

func (m *SessionSubscribed) GetSessionID() string {
	if m != nil {
		return m.SessionID
	}
	return ""
}

func (m *SessionSubscribed) GetTenant() string {
	if m != nil {
		return m.Tenant
	}
	return ""
}

func (m *SessionSubscribed) GetPattern() []byte {
	if m != nil {
		return m.Pattern
	}
	return nil
}

func (m *SessionSubscribed) GetQos() int32 {
	if m != nil {
		return m.Qos
	}
	return 0
}

type SessionUnsubscribed struct {
	SessionID            string   `protobuf:"bytes,1,opt,name=SessionID,proto3" json:"SessionID,omitempty"`
	Tenant               string   `protobuf:"bytes,2,opt,name=Tenant,proto3" json:"Tenant,omitempty"`
	Pattern              []byte   `protobuf:"bytes,3,opt,name=Pattern,proto3" json:"Pattern,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SessionUnsubscribed) Reset()         { *m = SessionUnsubscribed{} }
func (m *SessionUnsubscribed) String() string { return proto.CompactTextString(m) }
func (*SessionUnsubscribed) ProtoMessage()    {}
func (*SessionUnsubscribed) Descriptor() ([]byte, []int) {
	return fileDescriptor_8f22242cb04491f9, []int{5}
}

func (m *SessionUnsubscribed) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SessionUnsubscribed.Unmarshal(m, b)
}
func (m *SessionUnsubscribed) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SessionUnsubscribed.Marshal(b, m, deterministic)
}
func (m *SessionUnsubscribed) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SessionUnsubscribed.Merge(m, src)
}
func (m *SessionUnsubscribed) XXX_Size() int {
	return xxx_messageInfo_SessionUnsubscribed.Size(m)
}
func (m *SessionUnsubscribed) XXX_DiscardUnknown() {
	xxx_messageInfo_SessionUnsubscribed.DiscardUnknown(m)
}

var xxx_messageInfo_SessionUnsubscribed proto.InternalMessageInfo

func (m *SessionUnsubscribed) GetSessionID() string {
	if m != nil {
		return m.SessionID
	}
	return ""
}

func (m *SessionUnsubscribed) GetTenant() string {
	if m != nil {
		return m.Tenant
	}
	return ""
}

func (m *SessionUnsubscribed) GetPattern() []byte {
	if m != nil {
		return m.Pattern
	}
	return nil
}

type SessionKeepalived struct {
	SessionID            string   `protobuf:"bytes,1,opt,name=SessionID,proto3" json:"SessionID,omitempty"`
	Tenant               string   `protobuf:"bytes,2,opt,name=Tenant,proto3" json:"Tenant,omitempty"`
	Timestamp            int64    `protobuf:"varint,3,opt,name=Timestamp,proto3" json:"Timestamp,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SessionKeepalived) Reset()         { *m = SessionKeepalived{} }
func (m *SessionKeepalived) String() string { return proto.CompactTextString(m) }
func (*SessionKeepalived) ProtoMessage()    {}
func (*SessionKeepalived) Descriptor() ([]byte, []int) {
	return fileDescriptor_8f22242cb04491f9, []int{6}
}

func (m *SessionKeepalived) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SessionKeepalived.Unmarshal(m, b)
}
func (m *SessionKeepalived) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SessionKeepalived.Marshal(b, m, deterministic)
}
func (m *SessionKeepalived) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SessionKeepalived.Merge(m, src)
}
func (m *SessionKeepalived) XXX_Size() int {
	return xxx_messageInfo_SessionKeepalived.Size(m)
}
func (m *SessionKeepalived) XXX_DiscardUnknown() {
	xxx_messageInfo_SessionKeepalived.DiscardUnknown(m)
}

var xxx_messageInfo_SessionKeepalived proto.InternalMessageInfo

func (m *SessionKeepalived) GetSessionID() string {
	if m != nil {
		return m.SessionID
	}
	return ""
}

func (m *SessionKeepalived) GetTenant() string {
	if m != nil {
		return m.Tenant
	}
	return ""
}

func (m *SessionKeepalived) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

type StateTransitionSet struct {
	Events               []*StateTransition `protobuf:"bytes,1,rep,name=events,proto3" json:"events,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *StateTransitionSet) Reset()         { *m = StateTransitionSet{} }
func (m *StateTransitionSet) String() string { return proto.CompactTextString(m) }
func (*StateTransitionSet) ProtoMessage()    {}
func (*StateTransitionSet) Descriptor() ([]byte, []int) {
	return fileDescriptor_8f22242cb04491f9, []int{7}
}

func (m *StateTransitionSet) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StateTransitionSet.Unmarshal(m, b)
}
func (m *StateTransitionSet) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StateTransitionSet.Marshal(b, m, deterministic)
}
func (m *StateTransitionSet) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StateTransitionSet.Merge(m, src)
}
func (m *StateTransitionSet) XXX_Size() int {
	return xxx_messageInfo_StateTransitionSet.Size(m)
}
func (m *StateTransitionSet) XXX_DiscardUnknown() {
	xxx_messageInfo_StateTransitionSet.DiscardUnknown(m)
}

var xxx_messageInfo_StateTransitionSet proto.InternalMessageInfo

func (m *StateTransitionSet) GetEvents() []*StateTransition {
	if m != nil {
		return m.Events
	}
	return nil
}

type StateTransition struct {
	// Types that are valid to be assigned to Event:
	//	*StateTransition_SessionCreated
	//	*StateTransition_SessionClosed
	//	*StateTransition_SessionLost
	//	*StateTransition_SessionSubscribed
	//	*StateTransition_SessionUnsubscribed
	//	*StateTransition_SessionKeepalived
	Event                isStateTransition_Event `protobuf_oneof:"Event"`
	XXX_NoUnkeyedLiteral struct{}                `json:"-"`
	XXX_unrecognized     []byte                  `json:"-"`
	XXX_sizecache        int32                   `json:"-"`
}

func (m *StateTransition) Reset()         { *m = StateTransition{} }
func (m *StateTransition) String() string { return proto.CompactTextString(m) }
func (*StateTransition) ProtoMessage()    {}
func (*StateTransition) Descriptor() ([]byte, []int) {
	return fileDescriptor_8f22242cb04491f9, []int{8}
}

func (m *StateTransition) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StateTransition.Unmarshal(m, b)
}
func (m *StateTransition) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StateTransition.Marshal(b, m, deterministic)
}
func (m *StateTransition) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StateTransition.Merge(m, src)
}
func (m *StateTransition) XXX_Size() int {
	return xxx_messageInfo_StateTransition.Size(m)
}
func (m *StateTransition) XXX_DiscardUnknown() {
	xxx_messageInfo_StateTransition.DiscardUnknown(m)
}

var xxx_messageInfo_StateTransition proto.InternalMessageInfo

type isStateTransition_Event interface {
	isStateTransition_Event()
}

type StateTransition_SessionCreated struct {
	SessionCreated *SessionCreated `protobuf:"bytes,2,opt,name=SessionCreated,proto3,oneof"`
}

type StateTransition_SessionClosed struct {
	SessionClosed *SessionClosed `protobuf:"bytes,3,opt,name=SessionClosed,proto3,oneof"`
}

type StateTransition_SessionLost struct {
	SessionLost *SessionLost `protobuf:"bytes,4,opt,name=SessionLost,proto3,oneof"`
}

type StateTransition_SessionSubscribed struct {
	SessionSubscribed *SessionSubscribed `protobuf:"bytes,5,opt,name=SessionSubscribed,proto3,oneof"`
}

type StateTransition_SessionUnsubscribed struct {
	SessionUnsubscribed *SessionUnsubscribed `protobuf:"bytes,6,opt,name=SessionUnsubscribed,proto3,oneof"`
}

type StateTransition_SessionKeepalived struct {
	SessionKeepalived *SessionKeepalived `protobuf:"bytes,7,opt,name=SessionKeepalived,proto3,oneof"`
}

func (*StateTransition_SessionCreated) isStateTransition_Event() {}

func (*StateTransition_SessionClosed) isStateTransition_Event() {}

func (*StateTransition_SessionLost) isStateTransition_Event() {}

func (*StateTransition_SessionSubscribed) isStateTransition_Event() {}

func (*StateTransition_SessionUnsubscribed) isStateTransition_Event() {}

func (*StateTransition_SessionKeepalived) isStateTransition_Event() {}

func (m *StateTransition) GetEvent() isStateTransition_Event {
	if m != nil {
		return m.Event
	}
	return nil
}

func (m *StateTransition) GetSessionCreated() *SessionCreated {
	if x, ok := m.GetEvent().(*StateTransition_SessionCreated); ok {
		return x.SessionCreated
	}
	return nil
}

func (m *StateTransition) GetSessionClosed() *SessionClosed {
	if x, ok := m.GetEvent().(*StateTransition_SessionClosed); ok {
		return x.SessionClosed
	}
	return nil
}

func (m *StateTransition) GetSessionLost() *SessionLost {
	if x, ok := m.GetEvent().(*StateTransition_SessionLost); ok {
		return x.SessionLost
	}
	return nil
}

func (m *StateTransition) GetSessionSubscribed() *SessionSubscribed {
	if x, ok := m.GetEvent().(*StateTransition_SessionSubscribed); ok {
		return x.SessionSubscribed
	}
	return nil
}

func (m *StateTransition) GetSessionUnsubscribed() *SessionUnsubscribed {
	if x, ok := m.GetEvent().(*StateTransition_SessionUnsubscribed); ok {
		return x.SessionUnsubscribed
	}
	return nil
}

func (m *StateTransition) GetSessionKeepalived() *SessionKeepalived {
	if x, ok := m.GetEvent().(*StateTransition_SessionKeepalived); ok {
		return x.SessionKeepalived
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*StateTransition) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*StateTransition_SessionCreated)(nil),
		(*StateTransition_SessionClosed)(nil),
		(*StateTransition_SessionLost)(nil),
		(*StateTransition_SessionSubscribed)(nil),
		(*StateTransition_SessionUnsubscribed)(nil),
		(*StateTransition_SessionKeepalived)(nil),
	}
}

func init() {
	proto.RegisterType((*SessionCreated)(nil), "events.SessionCreated")
	proto.RegisterType((*SessionClosed)(nil), "events.SessionClosed")
	proto.RegisterType((*SessionLost)(nil), "events.SessionLost")
	proto.RegisterType((*PeerLost)(nil), "events.PeerLost")
	proto.RegisterType((*SessionSubscribed)(nil), "events.SessionSubscribed")
	proto.RegisterType((*SessionUnsubscribed)(nil), "events.SessionUnsubscribed")
	proto.RegisterType((*SessionKeepalived)(nil), "events.SessionKeepalived")
	proto.RegisterType((*StateTransitionSet)(nil), "events.StateTransitionSet")
	proto.RegisterType((*StateTransition)(nil), "events.StateTransition")
}

func init() { proto.RegisterFile("events.proto", fileDescriptor_8f22242cb04491f9) }

var fileDescriptor_8f22242cb04491f9 = []byte{
	// 528 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x54, 0x5d, 0x6b, 0x13, 0x4d,
	0x14, 0xce, 0x66, 0x9b, 0xaf, 0xb3, 0x49, 0xdf, 0xb7, 0x53, 0xac, 0x63, 0x15, 0x59, 0x16, 0x2f,
	0xf6, 0x42, 0x2a, 0xac, 0x48, 0xaf, 0x04, 0xb5, 0x29, 0x64, 0x51, 0x30, 0x9d, 0x44, 0xbc, 0x9e,
	0x34, 0x07, 0x59, 0xd8, 0xec, 0x84, 0x9d, 0x69, 0xc0, 0xdf, 0xe6, 0xbf, 0xf2, 0x17, 0xc8, 0x1c,
	0x37, 0xd9, 0x64, 0xba, 0x42, 0x11, 0xbc, 0x9b, 0xf3, 0x9c, 0x8f, 0x79, 0x66, 0xce, 0x73, 0x0e,
	0x0c, 0x71, 0x83, 0x85, 0xd1, 0x17, 0xeb, 0x52, 0x19, 0xc5, 0xba, 0xbf, 0xad, 0xe8, 0x67, 0x1b,
	0x8e, 0x67, 0xa8, 0x75, 0xa6, 0x8a, 0xab, 0x12, 0xa5, 0xc1, 0x25, 0x3b, 0x86, 0x76, 0x3a, 0xe6,
	0x5e, 0xe8, 0xc5, 0x03, 0xd1, 0x4e, 0xc7, 0xec, 0x0c, 0xba, 0x73, 0x2c, 0x64, 0x61, 0x78, 0x9b,
	0xb0, 0xca, 0x62, 0x0c, 0x8e, 0xa6, 0x88, 0x25, 0xf7, 0x09, 0xa5, 0x33, 0x7b, 0x06, 0x83, 0xaf,
	0x59, 0x9e, 0xcf, 0xd5, 0x3a, 0xbb, 0xe5, 0x47, 0xa1, 0x17, 0x0f, 0x45, 0x0d, 0x30, 0x0e, 0x3d,
	0x6b, 0xdc, 0xa8, 0x19, 0xef, 0x84, 0x5e, 0xdc, 0x11, 0x5b, 0x93, 0x85, 0x10, 0xd8, 0xe3, 0x54,
	0x7e, 0xcf, 0x95, 0x5c, 0xf2, 0x2e, 0x65, 0xee, 0x43, 0xec, 0x39, 0x80, 0x35, 0x05, 0x1a, 0x99,
	0x15, 0xbc, 0x17, 0x7a, 0x71, 0x5f, 0xec, 0x21, 0xec, 0x1c, 0xfa, 0x57, 0x79, 0x86, 0x85, 0x49,
	0xc7, 0x7c, 0x40, 0x8c, 0x76, 0xb6, 0x65, 0x35, 0x2f, 0x65, 0xa1, 0xd7, 0xaa, 0x34, 0x1c, 0xc8,
	0x59, 0x03, 0xec, 0x05, 0x8c, 0x04, 0xae, 0x94, 0xc1, 0xf7, 0xcb, 0x65, 0x89, 0x5a, 0xf3, 0x80,
	0x22, 0x0e, 0x41, 0xf6, 0x12, 0x4e, 0x3e, 0x22, 0xae, 0x65, 0x9e, 0x6d, 0x30, 0x2d, 0x0c, 0x96,
	0x1b, 0x99, 0xf3, 0x21, 0xbd, 0xe2, 0xbe, 0x83, 0x6e, 0xcc, 0x56, 0xa8, 0x8d, 0x5c, 0xad, 0xf9,
	0x28, 0xf4, 0x62, 0x5f, 0xd4, 0x40, 0x74, 0x09, 0xa3, 0xed, 0x9f, 0xe7, 0x4a, 0x3f, 0xfc, 0xcb,
	0xa3, 0x37, 0x10, 0x54, 0x89, 0x9f, 0x94, 0x36, 0x0f, 0x4e, 0x3b, 0x87, 0xbe, 0xed, 0x4e, 0x53,
	0x4e, 0x74, 0x07, 0x27, 0x55, 0xc9, 0xd9, 0xdd, 0x42, 0xdf, 0x96, 0xd9, 0x02, 0x97, 0x96, 0x7e,
	0x05, 0xee, 0x62, 0x6b, 0xe0, 0x8f, 0x82, 0xe0, 0xd0, 0x9b, 0x4a, 0x63, 0xb0, 0x2c, 0x48, 0x13,
	0x43, 0xb1, 0x35, 0xd9, 0xff, 0xe0, 0xdf, 0x28, 0x4d, 0x82, 0xe8, 0x08, 0x7b, 0x8c, 0x10, 0x4e,
	0xab, 0x82, 0x5f, 0x0a, 0xfd, 0xcf, 0x2e, 0x8e, 0xbe, 0xed, 0x5e, 0xb7, 0xeb, 0xd1, 0xdf, 0x5e,
	0x72, 0xd0, 0x52, 0xdf, 0x6d, 0xe9, 0x35, 0xb0, 0x99, 0x91, 0x06, 0x49, 0x56, 0x99, 0xb1, 0xdf,
	0x89, 0x86, 0xbd, 0x82, 0x6a, 0xce, 0xb8, 0x17, 0xfa, 0x71, 0x90, 0x3c, 0xbe, 0xa8, 0x86, 0xd0,
	0x89, 0x15, 0xdb, 0x71, 0xfc, 0xe1, 0xc3, 0x7f, 0x8e, 0x8f, 0xbd, 0x73, 0x27, 0x94, 0x88, 0x05,
	0xc9, 0xd9, 0xae, 0xd8, 0x81, 0x77, 0xd2, 0x12, 0xee, 0x44, 0xbf, 0x75, 0xf4, 0x46, 0xf4, 0x83,
	0xe4, 0x91, 0x5b, 0x80, 0x9c, 0x93, 0x96, 0x70, 0xd4, 0x79, 0x79, 0xa0, 0x3a, 0xea, 0x62, 0x90,
	0x9c, 0x3a, 0xc9, 0xd6, 0x35, 0x69, 0x89, 0x03, 0x7d, 0xa6, 0x0d, 0xda, 0xa2, 0xc9, 0x0f, 0x92,
	0x27, 0x4e, 0x7a, 0x1d, 0x30, 0x69, 0x89, 0x06, 0x45, 0x7e, 0x6e, 0xd4, 0x0b, 0x2d, 0x8a, 0x20,
	0x79, 0xea, 0x14, 0xdb, 0x0f, 0x99, 0xb4, 0x44, 0xa3, 0xd2, 0xd2, 0x06, 0x65, 0xd0, 0x5a, 0xb9,
	0xcf, 0xad, 0x0e, 0xd8, 0xe3, 0x56, 0x83, 0x1f, 0x7a, 0xd0, 0xb9, 0xb6, 0x09, 0x8b, 0x2e, 0xed,
	0xd6, 0xd7, 0xbf, 0x02, 0x00, 0x00, 0xff, 0xff, 0x0f, 0x73, 0x70, 0x98, 0x6b, 0x05, 0x00, 0x00,
}
