// Code generated by protoc-gen-go. DO NOT EDIT.
// source: types.proto

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

type Metadata struct {
	ID                   string   `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	SessionID            string   `protobuf:"bytes,2,opt,name=SessionID,proto3" json:"SessionID,omitempty"`
	Tenant               string   `protobuf:"bytes,3,opt,name=Tenant,proto3" json:"Tenant,omitempty"`
	Pattern              []byte   `protobuf:"bytes,4,opt,name=Pattern,proto3" json:"Pattern,omitempty"`
	Qos                  int32    `protobuf:"varint,5,opt,name=Qos,proto3" json:"Qos,omitempty"`
	Peer                 string   `protobuf:"bytes,6,opt,name=Peer,proto3" json:"Peer,omitempty"`
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

type SubscriptionCreateInput struct {
	ID                   string   `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	SessionID            string   `protobuf:"bytes,2,opt,name=SessionID,proto3" json:"SessionID,omitempty"`
	Tenant               string   `protobuf:"bytes,3,opt,name=Tenant,proto3" json:"Tenant,omitempty"`
	Pattern              []byte   `protobuf:"bytes,4,opt,name=Pattern,proto3" json:"Pattern,omitempty"`
	Qos                  int32    `protobuf:"varint,5,opt,name=Qos,proto3" json:"Qos,omitempty"`
	Peer                 string   `protobuf:"bytes,6,opt,name=Peer,proto3" json:"Peer,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SubscriptionCreateInput) Reset()         { *m = SubscriptionCreateInput{} }
func (m *SubscriptionCreateInput) String() string { return proto.CompactTextString(m) }
func (*SubscriptionCreateInput) ProtoMessage()    {}
func (*SubscriptionCreateInput) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{2}
}

func (m *SubscriptionCreateInput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SubscriptionCreateInput.Unmarshal(m, b)
}
func (m *SubscriptionCreateInput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SubscriptionCreateInput.Marshal(b, m, deterministic)
}
func (m *SubscriptionCreateInput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SubscriptionCreateInput.Merge(m, src)
}
func (m *SubscriptionCreateInput) XXX_Size() int {
	return xxx_messageInfo_SubscriptionCreateInput.Size(m)
}
func (m *SubscriptionCreateInput) XXX_DiscardUnknown() {
	xxx_messageInfo_SubscriptionCreateInput.DiscardUnknown(m)
}

var xxx_messageInfo_SubscriptionCreateInput proto.InternalMessageInfo

func (m *SubscriptionCreateInput) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *SubscriptionCreateInput) GetSessionID() string {
	if m != nil {
		return m.SessionID
	}
	return ""
}

func (m *SubscriptionCreateInput) GetTenant() string {
	if m != nil {
		return m.Tenant
	}
	return ""
}

func (m *SubscriptionCreateInput) GetPattern() []byte {
	if m != nil {
		return m.Pattern
	}
	return nil
}

func (m *SubscriptionCreateInput) GetQos() int32 {
	if m != nil {
		return m.Qos
	}
	return 0
}

func (m *SubscriptionCreateInput) GetPeer() string {
	if m != nil {
		return m.Peer
	}
	return ""
}

type SubscriptionCreateOutput struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SubscriptionCreateOutput) Reset()         { *m = SubscriptionCreateOutput{} }
func (m *SubscriptionCreateOutput) String() string { return proto.CompactTextString(m) }
func (*SubscriptionCreateOutput) ProtoMessage()    {}
func (*SubscriptionCreateOutput) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{3}
}

func (m *SubscriptionCreateOutput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SubscriptionCreateOutput.Unmarshal(m, b)
}
func (m *SubscriptionCreateOutput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SubscriptionCreateOutput.Marshal(b, m, deterministic)
}
func (m *SubscriptionCreateOutput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SubscriptionCreateOutput.Merge(m, src)
}
func (m *SubscriptionCreateOutput) XXX_Size() int {
	return xxx_messageInfo_SubscriptionCreateOutput.Size(m)
}
func (m *SubscriptionCreateOutput) XXX_DiscardUnknown() {
	xxx_messageInfo_SubscriptionCreateOutput.DiscardUnknown(m)
}

var xxx_messageInfo_SubscriptionCreateOutput proto.InternalMessageInfo

type SubscriptionByIDInput struct {
	ID                   string   `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SubscriptionByIDInput) Reset()         { *m = SubscriptionByIDInput{} }
func (m *SubscriptionByIDInput) String() string { return proto.CompactTextString(m) }
func (*SubscriptionByIDInput) ProtoMessage()    {}
func (*SubscriptionByIDInput) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{4}
}

func (m *SubscriptionByIDInput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SubscriptionByIDInput.Unmarshal(m, b)
}
func (m *SubscriptionByIDInput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SubscriptionByIDInput.Marshal(b, m, deterministic)
}
func (m *SubscriptionByIDInput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SubscriptionByIDInput.Merge(m, src)
}
func (m *SubscriptionByIDInput) XXX_Size() int {
	return xxx_messageInfo_SubscriptionByIDInput.Size(m)
}
func (m *SubscriptionByIDInput) XXX_DiscardUnknown() {
	xxx_messageInfo_SubscriptionByIDInput.DiscardUnknown(m)
}

var xxx_messageInfo_SubscriptionByIDInput proto.InternalMessageInfo

func (m *SubscriptionByIDInput) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

type SubscriptionBySessionInput struct {
	SessionID            string   `protobuf:"bytes,1,opt,name=SessionID,proto3" json:"SessionID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SubscriptionBySessionInput) Reset()         { *m = SubscriptionBySessionInput{} }
func (m *SubscriptionBySessionInput) String() string { return proto.CompactTextString(m) }
func (*SubscriptionBySessionInput) ProtoMessage()    {}
func (*SubscriptionBySessionInput) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{5}
}

func (m *SubscriptionBySessionInput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SubscriptionBySessionInput.Unmarshal(m, b)
}
func (m *SubscriptionBySessionInput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SubscriptionBySessionInput.Marshal(b, m, deterministic)
}
func (m *SubscriptionBySessionInput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SubscriptionBySessionInput.Merge(m, src)
}
func (m *SubscriptionBySessionInput) XXX_Size() int {
	return xxx_messageInfo_SubscriptionBySessionInput.Size(m)
}
func (m *SubscriptionBySessionInput) XXX_DiscardUnknown() {
	xxx_messageInfo_SubscriptionBySessionInput.DiscardUnknown(m)
}

var xxx_messageInfo_SubscriptionBySessionInput proto.InternalMessageInfo

func (m *SubscriptionBySessionInput) GetSessionID() string {
	if m != nil {
		return m.SessionID
	}
	return ""
}

type SubscriptionByTopicInput struct {
	Topic                []byte   `protobuf:"bytes,1,opt,name=Topic,proto3" json:"Topic,omitempty"`
	Tenant               string   `protobuf:"bytes,2,opt,name=Tenant,proto3" json:"Tenant,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SubscriptionByTopicInput) Reset()         { *m = SubscriptionByTopicInput{} }
func (m *SubscriptionByTopicInput) String() string { return proto.CompactTextString(m) }
func (*SubscriptionByTopicInput) ProtoMessage()    {}
func (*SubscriptionByTopicInput) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{6}
}

func (m *SubscriptionByTopicInput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SubscriptionByTopicInput.Unmarshal(m, b)
}
func (m *SubscriptionByTopicInput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SubscriptionByTopicInput.Marshal(b, m, deterministic)
}
func (m *SubscriptionByTopicInput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SubscriptionByTopicInput.Merge(m, src)
}
func (m *SubscriptionByTopicInput) XXX_Size() int {
	return xxx_messageInfo_SubscriptionByTopicInput.Size(m)
}
func (m *SubscriptionByTopicInput) XXX_DiscardUnknown() {
	xxx_messageInfo_SubscriptionByTopicInput.DiscardUnknown(m)
}

var xxx_messageInfo_SubscriptionByTopicInput proto.InternalMessageInfo

func (m *SubscriptionByTopicInput) GetTopic() []byte {
	if m != nil {
		return m.Topic
	}
	return nil
}

func (m *SubscriptionByTopicInput) GetTenant() string {
	if m != nil {
		return m.Tenant
	}
	return ""
}

type SubscriptionFilterInput struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SubscriptionFilterInput) Reset()         { *m = SubscriptionFilterInput{} }
func (m *SubscriptionFilterInput) String() string { return proto.CompactTextString(m) }
func (*SubscriptionFilterInput) ProtoMessage()    {}
func (*SubscriptionFilterInput) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{7}
}

func (m *SubscriptionFilterInput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SubscriptionFilterInput.Unmarshal(m, b)
}
func (m *SubscriptionFilterInput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SubscriptionFilterInput.Marshal(b, m, deterministic)
}
func (m *SubscriptionFilterInput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SubscriptionFilterInput.Merge(m, src)
}
func (m *SubscriptionFilterInput) XXX_Size() int {
	return xxx_messageInfo_SubscriptionFilterInput.Size(m)
}
func (m *SubscriptionFilterInput) XXX_DiscardUnknown() {
	xxx_messageInfo_SubscriptionFilterInput.DiscardUnknown(m)
}

var xxx_messageInfo_SubscriptionFilterInput proto.InternalMessageInfo

type SubscriptionDeleteInput struct {
	ID                   string   `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SubscriptionDeleteInput) Reset()         { *m = SubscriptionDeleteInput{} }
func (m *SubscriptionDeleteInput) String() string { return proto.CompactTextString(m) }
func (*SubscriptionDeleteInput) ProtoMessage()    {}
func (*SubscriptionDeleteInput) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{8}
}

func (m *SubscriptionDeleteInput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SubscriptionDeleteInput.Unmarshal(m, b)
}
func (m *SubscriptionDeleteInput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SubscriptionDeleteInput.Marshal(b, m, deterministic)
}
func (m *SubscriptionDeleteInput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SubscriptionDeleteInput.Merge(m, src)
}
func (m *SubscriptionDeleteInput) XXX_Size() int {
	return xxx_messageInfo_SubscriptionDeleteInput.Size(m)
}
func (m *SubscriptionDeleteInput) XXX_DiscardUnknown() {
	xxx_messageInfo_SubscriptionDeleteInput.DiscardUnknown(m)
}

var xxx_messageInfo_SubscriptionDeleteInput proto.InternalMessageInfo

func (m *SubscriptionDeleteInput) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

type SubscriptionDeleteOutput struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SubscriptionDeleteOutput) Reset()         { *m = SubscriptionDeleteOutput{} }
func (m *SubscriptionDeleteOutput) String() string { return proto.CompactTextString(m) }
func (*SubscriptionDeleteOutput) ProtoMessage()    {}
func (*SubscriptionDeleteOutput) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{9}
}

func (m *SubscriptionDeleteOutput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SubscriptionDeleteOutput.Unmarshal(m, b)
}
func (m *SubscriptionDeleteOutput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SubscriptionDeleteOutput.Marshal(b, m, deterministic)
}
func (m *SubscriptionDeleteOutput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SubscriptionDeleteOutput.Merge(m, src)
}
func (m *SubscriptionDeleteOutput) XXX_Size() int {
	return xxx_messageInfo_SubscriptionDeleteOutput.Size(m)
}
func (m *SubscriptionDeleteOutput) XXX_DiscardUnknown() {
	xxx_messageInfo_SubscriptionDeleteOutput.DiscardUnknown(m)
}

var xxx_messageInfo_SubscriptionDeleteOutput proto.InternalMessageInfo

type SubscriptionStateTransition struct {
	Kind                 string                                          `protobuf:"bytes,1,opt,name=Kind,proto3" json:"Kind,omitempty"`
	SubscriptionCreated  *SubscriptionStateTransitionSubscriptionCreated `protobuf:"bytes,2,opt,name=SubscriptionCreated,proto3" json:"SubscriptionCreated,omitempty"`
	SubscriptionDeleted  *SubscriptionStateTransitionSubscriptionDeleted `protobuf:"bytes,3,opt,name=SubscriptionDeleted,proto3" json:"SubscriptionDeleted,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                                        `json:"-"`
	XXX_unrecognized     []byte                                          `json:"-"`
	XXX_sizecache        int32                                           `json:"-"`
}

func (m *SubscriptionStateTransition) Reset()         { *m = SubscriptionStateTransition{} }
func (m *SubscriptionStateTransition) String() string { return proto.CompactTextString(m) }
func (*SubscriptionStateTransition) ProtoMessage()    {}
func (*SubscriptionStateTransition) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{10}
}

func (m *SubscriptionStateTransition) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SubscriptionStateTransition.Unmarshal(m, b)
}
func (m *SubscriptionStateTransition) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SubscriptionStateTransition.Marshal(b, m, deterministic)
}
func (m *SubscriptionStateTransition) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SubscriptionStateTransition.Merge(m, src)
}
func (m *SubscriptionStateTransition) XXX_Size() int {
	return xxx_messageInfo_SubscriptionStateTransition.Size(m)
}
func (m *SubscriptionStateTransition) XXX_DiscardUnknown() {
	xxx_messageInfo_SubscriptionStateTransition.DiscardUnknown(m)
}

var xxx_messageInfo_SubscriptionStateTransition proto.InternalMessageInfo

func (m *SubscriptionStateTransition) GetKind() string {
	if m != nil {
		return m.Kind
	}
	return ""
}

func (m *SubscriptionStateTransition) GetSubscriptionCreated() *SubscriptionStateTransitionSubscriptionCreated {
	if m != nil {
		return m.SubscriptionCreated
	}
	return nil
}

func (m *SubscriptionStateTransition) GetSubscriptionDeleted() *SubscriptionStateTransitionSubscriptionDeleted {
	if m != nil {
		return m.SubscriptionDeleted
	}
	return nil
}

type SubscriptionStateTransitionSubscriptionCreated struct {
	Input                *SubscriptionCreateInput `protobuf:"bytes,1,opt,name=Input,proto3" json:"Input,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                 `json:"-"`
	XXX_unrecognized     []byte                   `json:"-"`
	XXX_sizecache        int32                    `json:"-"`
}

func (m *SubscriptionStateTransitionSubscriptionCreated) Reset() {
	*m = SubscriptionStateTransitionSubscriptionCreated{}
}
func (m *SubscriptionStateTransitionSubscriptionCreated) String() string {
	return proto.CompactTextString(m)
}
func (*SubscriptionStateTransitionSubscriptionCreated) ProtoMessage() {}
func (*SubscriptionStateTransitionSubscriptionCreated) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{11}
}

func (m *SubscriptionStateTransitionSubscriptionCreated) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SubscriptionStateTransitionSubscriptionCreated.Unmarshal(m, b)
}
func (m *SubscriptionStateTransitionSubscriptionCreated) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SubscriptionStateTransitionSubscriptionCreated.Marshal(b, m, deterministic)
}
func (m *SubscriptionStateTransitionSubscriptionCreated) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SubscriptionStateTransitionSubscriptionCreated.Merge(m, src)
}
func (m *SubscriptionStateTransitionSubscriptionCreated) XXX_Size() int {
	return xxx_messageInfo_SubscriptionStateTransitionSubscriptionCreated.Size(m)
}
func (m *SubscriptionStateTransitionSubscriptionCreated) XXX_DiscardUnknown() {
	xxx_messageInfo_SubscriptionStateTransitionSubscriptionCreated.DiscardUnknown(m)
}

var xxx_messageInfo_SubscriptionStateTransitionSubscriptionCreated proto.InternalMessageInfo

func (m *SubscriptionStateTransitionSubscriptionCreated) GetInput() *SubscriptionCreateInput {
	if m != nil {
		return m.Input
	}
	return nil
}

type SubscriptionStateTransitionSubscriptionDeleted struct {
	ID                   string   `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SubscriptionStateTransitionSubscriptionDeleted) Reset() {
	*m = SubscriptionStateTransitionSubscriptionDeleted{}
}
func (m *SubscriptionStateTransitionSubscriptionDeleted) String() string {
	return proto.CompactTextString(m)
}
func (*SubscriptionStateTransitionSubscriptionDeleted) ProtoMessage() {}
func (*SubscriptionStateTransitionSubscriptionDeleted) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{12}
}

func (m *SubscriptionStateTransitionSubscriptionDeleted) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SubscriptionStateTransitionSubscriptionDeleted.Unmarshal(m, b)
}
func (m *SubscriptionStateTransitionSubscriptionDeleted) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SubscriptionStateTransitionSubscriptionDeleted.Marshal(b, m, deterministic)
}
func (m *SubscriptionStateTransitionSubscriptionDeleted) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SubscriptionStateTransitionSubscriptionDeleted.Merge(m, src)
}
func (m *SubscriptionStateTransitionSubscriptionDeleted) XXX_Size() int {
	return xxx_messageInfo_SubscriptionStateTransitionSubscriptionDeleted.Size(m)
}
func (m *SubscriptionStateTransitionSubscriptionDeleted) XXX_DiscardUnknown() {
	xxx_messageInfo_SubscriptionStateTransitionSubscriptionDeleted.DiscardUnknown(m)
}

var xxx_messageInfo_SubscriptionStateTransitionSubscriptionDeleted proto.InternalMessageInfo

func (m *SubscriptionStateTransitionSubscriptionDeleted) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func init() {
	proto.RegisterType((*Metadata)(nil), "pb.Metadata")
	proto.RegisterType((*SubscriptionMetadataList)(nil), "pb.SubscriptionMetadataList")
	proto.RegisterType((*SubscriptionCreateInput)(nil), "pb.SubscriptionCreateInput")
	proto.RegisterType((*SubscriptionCreateOutput)(nil), "pb.SubscriptionCreateOutput")
	proto.RegisterType((*SubscriptionByIDInput)(nil), "pb.SubscriptionByIDInput")
	proto.RegisterType((*SubscriptionBySessionInput)(nil), "pb.SubscriptionBySessionInput")
	proto.RegisterType((*SubscriptionByTopicInput)(nil), "pb.SubscriptionByTopicInput")
	proto.RegisterType((*SubscriptionFilterInput)(nil), "pb.SubscriptionFilterInput")
	proto.RegisterType((*SubscriptionDeleteInput)(nil), "pb.SubscriptionDeleteInput")
	proto.RegisterType((*SubscriptionDeleteOutput)(nil), "pb.SubscriptionDeleteOutput")
	proto.RegisterType((*SubscriptionStateTransition)(nil), "pb.SubscriptionStateTransition")
	proto.RegisterType((*SubscriptionStateTransitionSubscriptionCreated)(nil), "pb.SubscriptionStateTransitionSubscriptionCreated")
	proto.RegisterType((*SubscriptionStateTransitionSubscriptionDeleted)(nil), "pb.SubscriptionStateTransitionSubscriptionDeleted")
}

func init() { proto.RegisterFile("types.proto", fileDescriptor_d938547f84707355) }

var fileDescriptor_d938547f84707355 = []byte{
	// 505 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xcc, 0x55, 0xcd, 0x8e, 0xd3, 0x30,
	0x10, 0x96, 0x93, 0xb6, 0x4b, 0xa7, 0x15, 0x42, 0xc3, 0x02, 0xde, 0x6e, 0x85, 0x2a, 0x5f, 0x28,
	0x1c, 0x8a, 0x08, 0x37, 0x4e, 0xcb, 0x6e, 0x58, 0x11, 0x01, 0x62, 0x49, 0xfb, 0x02, 0x69, 0xe3,
	0x83, 0xa5, 0x2a, 0x89, 0x62, 0x17, 0xa9, 0x6f, 0xc1, 0x1b, 0x70, 0xe2, 0x31, 0x91, 0x50, 0xec,
	0x64, 0xeb, 0xfc, 0x50, 0xd4, 0x1b, 0x37, 0x8f, 0x67, 0xe6, 0x9b, 0x6f, 0xbe, 0xcc, 0x38, 0x30,
	0x52, 0xfb, 0x8c, 0xcb, 0x45, 0x96, 0xa7, 0x2a, 0x45, 0x27, 0x5b, 0xb3, 0x1f, 0x04, 0x1e, 0x7c,
	0xe1, 0x2a, 0x8a, 0x23, 0x15, 0xe1, 0x43, 0x70, 0x02, 0x9f, 0x92, 0x19, 0x99, 0x0f, 0x43, 0x27,
	0xf0, 0x71, 0x0a, 0xc3, 0x25, 0x97, 0x52, 0xa4, 0x49, 0xe0, 0x53, 0x47, 0x5f, 0x1f, 0x2e, 0xf0,
	0x29, 0x0c, 0x56, 0x3c, 0x89, 0x12, 0x45, 0x5d, 0xed, 0x2a, 0x2d, 0xa4, 0x70, 0x76, 0x17, 0x29,
	0xc5, 0xf3, 0x84, 0xf6, 0x66, 0x64, 0x3e, 0x0e, 0x2b, 0x13, 0x1f, 0x81, 0xfb, 0x2d, 0x95, 0xb4,
	0x3f, 0x23, 0xf3, 0x7e, 0x58, 0x1c, 0x11, 0xa1, 0x77, 0xc7, 0x79, 0x4e, 0x07, 0x1a, 0x41, 0x9f,
	0xd9, 0x2d, 0xd0, 0xe5, 0x6e, 0x2d, 0x37, 0xb9, 0xc8, 0x94, 0x48, 0x93, 0x8a, 0xdd, 0x67, 0x21,
	0x15, 0xbe, 0x82, 0x61, 0x65, 0x4b, 0x4a, 0x66, 0xee, 0x7c, 0xe4, 0x8d, 0x17, 0xd9, 0x7a, 0x51,
	0x5d, 0x86, 0x07, 0x37, 0xfb, 0x49, 0xe0, 0x99, 0x0d, 0x74, 0x93, 0xf3, 0x48, 0xf1, 0x20, 0xc9,
	0x76, 0xea, 0x3f, 0xe9, 0x74, 0x52, 0xef, 0xd4, 0x10, 0xfc, 0xba, 0x53, 0xd9, 0x4e, 0xb1, 0x17,
	0xf0, 0xc4, 0xf6, 0x5d, 0xef, 0x03, 0xbf, 0x93, 0x3a, 0x7b, 0x07, 0x93, 0x7a, 0x60, 0xc5, 0x5b,
	0x47, 0xd7, 0x1a, 0x23, 0x8d, 0xc6, 0xd8, 0xc7, 0x3a, 0x81, 0xeb, 0xfd, 0x2a, 0xcd, 0xc4, 0xc6,
	0x64, 0x9e, 0x43, 0x5f, 0x5b, 0x3a, 0x6b, 0x1c, 0x1a, 0xc3, 0x92, 0xc2, 0xb1, 0xa5, 0x60, 0x17,
	0x75, 0xad, 0x6f, 0xc5, 0x56, 0xf1, 0x5c, 0x03, 0xb1, 0x97, 0x75, 0x97, 0xcf, 0xb7, 0xfc, 0x2f,
	0x9f, 0xa1, 0x29, 0x88, 0x09, 0x2d, 0x05, 0xf9, 0x4d, 0xe0, 0xd2, 0x76, 0x2e, 0x55, 0xa4, 0xf8,
	0x2a, 0x8f, 0x12, 0x29, 0x0a, 0xb3, 0x10, 0xf8, 0x93, 0x48, 0xe2, 0x12, 0x4d, 0x9f, 0x31, 0x86,
	0xc7, 0x6d, 0x81, 0x63, 0x4d, 0x7d, 0xe4, 0x79, 0xc5, 0xe0, 0x1c, 0x41, 0xec, 0xc8, 0x0c, 0xbb,
	0xe0, 0x9a, 0x55, 0x0c, 0xeb, 0x58, 0xcf, 0xca, 0x69, 0x55, 0xca, 0xcc, 0xb0, 0x0b, 0x8e, 0x6d,
	0xe0, 0x44, 0xb2, 0xf8, 0x06, 0xfa, 0x5a, 0x66, 0x2d, 0xc9, 0xc8, 0xbb, 0x6c, 0x32, 0xb1, 0x16,
	0x22, 0x34, 0x91, 0xec, 0x0a, 0x4e, 0xe4, 0xda, 0xfc, 0x84, 0xde, 0x2f, 0x17, 0xce, 0xed, 0x38,
	0xb9, 0xe4, 0xf9, 0x77, 0xb1, 0xe1, 0x78, 0x03, 0x03, 0x53, 0x10, 0x8f, 0x11, 0x99, 0x4c, 0xbb,
	0x9d, 0x66, 0x08, 0x0a, 0x10, 0x53, 0xb8, 0x0d, 0x62, 0xcd, 0x55, 0x1b, 0xc4, 0x9e, 0x24, 0x7c,
	0x0d, 0xbd, 0x62, 0x9d, 0xf0, 0xa2, 0x19, 0x75, 0xbf, 0x64, 0x93, 0xda, 0xa3, 0x82, 0x01, 0x0c,
	0xef, 0xd7, 0x0a, 0x9f, 0xb7, 0xb3, 0xec, 0x8d, 0x6b, 0xd7, 0xae, 0x3d, 0x60, 0x1f, 0xe0, 0xac,
	0xdc, 0x32, 0x9c, 0xb6, 0x81, 0x0e, 0xeb, 0xf7, 0x0f, 0x98, 0x2b, 0x70, 0xdf, 0x6f, 0xb7, 0x6d,
	0x11, 0xac, 0xbd, 0x3b, 0x8e, 0xb0, 0x1e, 0xe8, 0x7f, 0xc0, 0xdb, 0x3f, 0x01, 0x00, 0x00, 0xff,
	0xff, 0xfc, 0xdc, 0x8e, 0xcf, 0x12, 0x06, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// SubscriptionsServiceClient is the client API for SubscriptionsService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type SubscriptionsServiceClient interface {
	Create(ctx context.Context, in *SubscriptionCreateInput, opts ...grpc.CallOption) (*SubscriptionCreateOutput, error)
	Delete(ctx context.Context, in *SubscriptionDeleteInput, opts ...grpc.CallOption) (*SubscriptionDeleteOutput, error)
	ByID(ctx context.Context, in *SubscriptionByIDInput, opts ...grpc.CallOption) (*Metadata, error)
	BySession(ctx context.Context, in *SubscriptionBySessionInput, opts ...grpc.CallOption) (*SubscriptionMetadataList, error)
	ByTopic(ctx context.Context, in *SubscriptionByTopicInput, opts ...grpc.CallOption) (*SubscriptionMetadataList, error)
	All(ctx context.Context, in *SubscriptionFilterInput, opts ...grpc.CallOption) (*SubscriptionMetadataList, error)
}

type subscriptionsServiceClient struct {
	cc *grpc.ClientConn
}

func NewSubscriptionsServiceClient(cc *grpc.ClientConn) SubscriptionsServiceClient {
	return &subscriptionsServiceClient{cc}
}

func (c *subscriptionsServiceClient) Create(ctx context.Context, in *SubscriptionCreateInput, opts ...grpc.CallOption) (*SubscriptionCreateOutput, error) {
	out := new(SubscriptionCreateOutput)
	err := c.cc.Invoke(ctx, "/pb.SubscriptionsService/Create", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *subscriptionsServiceClient) Delete(ctx context.Context, in *SubscriptionDeleteInput, opts ...grpc.CallOption) (*SubscriptionDeleteOutput, error) {
	out := new(SubscriptionDeleteOutput)
	err := c.cc.Invoke(ctx, "/pb.SubscriptionsService/Delete", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *subscriptionsServiceClient) ByID(ctx context.Context, in *SubscriptionByIDInput, opts ...grpc.CallOption) (*Metadata, error) {
	out := new(Metadata)
	err := c.cc.Invoke(ctx, "/pb.SubscriptionsService/ByID", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *subscriptionsServiceClient) BySession(ctx context.Context, in *SubscriptionBySessionInput, opts ...grpc.CallOption) (*SubscriptionMetadataList, error) {
	out := new(SubscriptionMetadataList)
	err := c.cc.Invoke(ctx, "/pb.SubscriptionsService/BySession", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *subscriptionsServiceClient) ByTopic(ctx context.Context, in *SubscriptionByTopicInput, opts ...grpc.CallOption) (*SubscriptionMetadataList, error) {
	out := new(SubscriptionMetadataList)
	err := c.cc.Invoke(ctx, "/pb.SubscriptionsService/ByTopic", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *subscriptionsServiceClient) All(ctx context.Context, in *SubscriptionFilterInput, opts ...grpc.CallOption) (*SubscriptionMetadataList, error) {
	out := new(SubscriptionMetadataList)
	err := c.cc.Invoke(ctx, "/pb.SubscriptionsService/All", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SubscriptionsServiceServer is the server API for SubscriptionsService service.
type SubscriptionsServiceServer interface {
	Create(context.Context, *SubscriptionCreateInput) (*SubscriptionCreateOutput, error)
	Delete(context.Context, *SubscriptionDeleteInput) (*SubscriptionDeleteOutput, error)
	ByID(context.Context, *SubscriptionByIDInput) (*Metadata, error)
	BySession(context.Context, *SubscriptionBySessionInput) (*SubscriptionMetadataList, error)
	ByTopic(context.Context, *SubscriptionByTopicInput) (*SubscriptionMetadataList, error)
	All(context.Context, *SubscriptionFilterInput) (*SubscriptionMetadataList, error)
}

// UnimplementedSubscriptionsServiceServer can be embedded to have forward compatible implementations.
type UnimplementedSubscriptionsServiceServer struct {
}

func (*UnimplementedSubscriptionsServiceServer) Create(ctx context.Context, req *SubscriptionCreateInput) (*SubscriptionCreateOutput, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}
func (*UnimplementedSubscriptionsServiceServer) Delete(ctx context.Context, req *SubscriptionDeleteInput) (*SubscriptionDeleteOutput, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}
func (*UnimplementedSubscriptionsServiceServer) ByID(ctx context.Context, req *SubscriptionByIDInput) (*Metadata, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ByID not implemented")
}
func (*UnimplementedSubscriptionsServiceServer) BySession(ctx context.Context, req *SubscriptionBySessionInput) (*SubscriptionMetadataList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BySession not implemented")
}
func (*UnimplementedSubscriptionsServiceServer) ByTopic(ctx context.Context, req *SubscriptionByTopicInput) (*SubscriptionMetadataList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ByTopic not implemented")
}
func (*UnimplementedSubscriptionsServiceServer) All(ctx context.Context, req *SubscriptionFilterInput) (*SubscriptionMetadataList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method All not implemented")
}

func RegisterSubscriptionsServiceServer(s *grpc.Server, srv SubscriptionsServiceServer) {
	s.RegisterService(&_SubscriptionsService_serviceDesc, srv)
}

func _SubscriptionsService_Create_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubscriptionCreateInput)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SubscriptionsServiceServer).Create(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.SubscriptionsService/Create",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SubscriptionsServiceServer).Create(ctx, req.(*SubscriptionCreateInput))
	}
	return interceptor(ctx, in, info, handler)
}

func _SubscriptionsService_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubscriptionDeleteInput)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SubscriptionsServiceServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.SubscriptionsService/Delete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SubscriptionsServiceServer).Delete(ctx, req.(*SubscriptionDeleteInput))
	}
	return interceptor(ctx, in, info, handler)
}

func _SubscriptionsService_ByID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubscriptionByIDInput)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SubscriptionsServiceServer).ByID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.SubscriptionsService/ByID",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SubscriptionsServiceServer).ByID(ctx, req.(*SubscriptionByIDInput))
	}
	return interceptor(ctx, in, info, handler)
}

func _SubscriptionsService_BySession_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubscriptionBySessionInput)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SubscriptionsServiceServer).BySession(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.SubscriptionsService/BySession",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SubscriptionsServiceServer).BySession(ctx, req.(*SubscriptionBySessionInput))
	}
	return interceptor(ctx, in, info, handler)
}

func _SubscriptionsService_ByTopic_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubscriptionByTopicInput)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SubscriptionsServiceServer).ByTopic(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.SubscriptionsService/ByTopic",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SubscriptionsServiceServer).ByTopic(ctx, req.(*SubscriptionByTopicInput))
	}
	return interceptor(ctx, in, info, handler)
}

func _SubscriptionsService_All_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubscriptionFilterInput)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SubscriptionsServiceServer).All(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.SubscriptionsService/All",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SubscriptionsServiceServer).All(ctx, req.(*SubscriptionFilterInput))
	}
	return interceptor(ctx, in, info, handler)
}

var _SubscriptionsService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "pb.SubscriptionsService",
	HandlerType: (*SubscriptionsServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Create",
			Handler:    _SubscriptionsService_Create_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _SubscriptionsService_Delete_Handler,
		},
		{
			MethodName: "ByID",
			Handler:    _SubscriptionsService_ByID_Handler,
		},
		{
			MethodName: "BySession",
			Handler:    _SubscriptionsService_BySession_Handler,
		},
		{
			MethodName: "ByTopic",
			Handler:    _SubscriptionsService_ByTopic_Handler,
		},
		{
			MethodName: "All",
			Handler:    _SubscriptionsService_All_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "types.proto",
}
