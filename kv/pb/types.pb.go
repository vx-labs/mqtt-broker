// Code generated by protoc-gen-go. DO NOT EDIT.
// source: types.proto

package pb

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	_ "github.com/vx-labs/mqtt-protocol/packet"
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

type KVMetadata struct {
	Version              uint64   `protobuf:"varint,1,opt,name=Version,proto3" json:"Version,omitempty"`
	Deadline             uint64   `protobuf:"varint,2,opt,name=Deadline,proto3" json:"Deadline,omitempty"`
	Key                  []byte   `protobuf:"bytes,3,opt,name=Key,proto3" json:"Key,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KVMetadata) Reset()         { *m = KVMetadata{} }
func (m *KVMetadata) String() string { return proto.CompactTextString(m) }
func (*KVMetadata) ProtoMessage()    {}
func (*KVMetadata) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{0}
}

func (m *KVMetadata) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KVMetadata.Unmarshal(m, b)
}
func (m *KVMetadata) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KVMetadata.Marshal(b, m, deterministic)
}
func (m *KVMetadata) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KVMetadata.Merge(m, src)
}
func (m *KVMetadata) XXX_Size() int {
	return xxx_messageInfo_KVMetadata.Size(m)
}
func (m *KVMetadata) XXX_DiscardUnknown() {
	xxx_messageInfo_KVMetadata.DiscardUnknown(m)
}

var xxx_messageInfo_KVMetadata proto.InternalMessageInfo

func (m *KVMetadata) GetVersion() uint64 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *KVMetadata) GetDeadline() uint64 {
	if m != nil {
		return m.Deadline
	}
	return 0
}

func (m *KVMetadata) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

type KVGetInput struct {
	Key                  []byte   `protobuf:"bytes,1,opt,name=Key,proto3" json:"Key,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KVGetInput) Reset()         { *m = KVGetInput{} }
func (m *KVGetInput) String() string { return proto.CompactTextString(m) }
func (*KVGetInput) ProtoMessage()    {}
func (*KVGetInput) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{1}
}

func (m *KVGetInput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KVGetInput.Unmarshal(m, b)
}
func (m *KVGetInput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KVGetInput.Marshal(b, m, deterministic)
}
func (m *KVGetInput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KVGetInput.Merge(m, src)
}
func (m *KVGetInput) XXX_Size() int {
	return xxx_messageInfo_KVGetInput.Size(m)
}
func (m *KVGetInput) XXX_DiscardUnknown() {
	xxx_messageInfo_KVGetInput.DiscardUnknown(m)
}

var xxx_messageInfo_KVGetInput proto.InternalMessageInfo

func (m *KVGetInput) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

type KVGetOutput struct {
	Key                  []byte   `protobuf:"bytes,1,opt,name=Key,proto3" json:"Key,omitempty"`
	Value                []byte   `protobuf:"bytes,2,opt,name=Value,proto3" json:"Value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KVGetOutput) Reset()         { *m = KVGetOutput{} }
func (m *KVGetOutput) String() string { return proto.CompactTextString(m) }
func (*KVGetOutput) ProtoMessage()    {}
func (*KVGetOutput) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{2}
}

func (m *KVGetOutput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KVGetOutput.Unmarshal(m, b)
}
func (m *KVGetOutput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KVGetOutput.Marshal(b, m, deterministic)
}
func (m *KVGetOutput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KVGetOutput.Merge(m, src)
}
func (m *KVGetOutput) XXX_Size() int {
	return xxx_messageInfo_KVGetOutput.Size(m)
}
func (m *KVGetOutput) XXX_DiscardUnknown() {
	xxx_messageInfo_KVGetOutput.DiscardUnknown(m)
}

var xxx_messageInfo_KVGetOutput proto.InternalMessageInfo

func (m *KVGetOutput) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *KVGetOutput) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

type KVGetWithMetadataInput struct {
	Key                  []byte   `protobuf:"bytes,1,opt,name=Key,proto3" json:"Key,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KVGetWithMetadataInput) Reset()         { *m = KVGetWithMetadataInput{} }
func (m *KVGetWithMetadataInput) String() string { return proto.CompactTextString(m) }
func (*KVGetWithMetadataInput) ProtoMessage()    {}
func (*KVGetWithMetadataInput) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{3}
}

func (m *KVGetWithMetadataInput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KVGetWithMetadataInput.Unmarshal(m, b)
}
func (m *KVGetWithMetadataInput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KVGetWithMetadataInput.Marshal(b, m, deterministic)
}
func (m *KVGetWithMetadataInput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KVGetWithMetadataInput.Merge(m, src)
}
func (m *KVGetWithMetadataInput) XXX_Size() int {
	return xxx_messageInfo_KVGetWithMetadataInput.Size(m)
}
func (m *KVGetWithMetadataInput) XXX_DiscardUnknown() {
	xxx_messageInfo_KVGetWithMetadataInput.DiscardUnknown(m)
}

var xxx_messageInfo_KVGetWithMetadataInput proto.InternalMessageInfo

func (m *KVGetWithMetadataInput) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

type KVGetWithMetadataOutput struct {
	Key                  []byte      `protobuf:"bytes,1,opt,name=Key,proto3" json:"Key,omitempty"`
	Value                []byte      `protobuf:"bytes,2,opt,name=Value,proto3" json:"Value,omitempty"`
	Metadata             *KVMetadata `protobuf:"bytes,3,opt,name=Metadata,proto3" json:"Metadata,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *KVGetWithMetadataOutput) Reset()         { *m = KVGetWithMetadataOutput{} }
func (m *KVGetWithMetadataOutput) String() string { return proto.CompactTextString(m) }
func (*KVGetWithMetadataOutput) ProtoMessage()    {}
func (*KVGetWithMetadataOutput) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{4}
}

func (m *KVGetWithMetadataOutput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KVGetWithMetadataOutput.Unmarshal(m, b)
}
func (m *KVGetWithMetadataOutput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KVGetWithMetadataOutput.Marshal(b, m, deterministic)
}
func (m *KVGetWithMetadataOutput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KVGetWithMetadataOutput.Merge(m, src)
}
func (m *KVGetWithMetadataOutput) XXX_Size() int {
	return xxx_messageInfo_KVGetWithMetadataOutput.Size(m)
}
func (m *KVGetWithMetadataOutput) XXX_DiscardUnknown() {
	xxx_messageInfo_KVGetWithMetadataOutput.DiscardUnknown(m)
}

var xxx_messageInfo_KVGetWithMetadataOutput proto.InternalMessageInfo

func (m *KVGetWithMetadataOutput) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *KVGetWithMetadataOutput) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *KVGetWithMetadataOutput) GetMetadata() *KVMetadata {
	if m != nil {
		return m.Metadata
	}
	return nil
}

type KVGetMetadataInput struct {
	Key                  []byte   `protobuf:"bytes,1,opt,name=Key,proto3" json:"Key,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KVGetMetadataInput) Reset()         { *m = KVGetMetadataInput{} }
func (m *KVGetMetadataInput) String() string { return proto.CompactTextString(m) }
func (*KVGetMetadataInput) ProtoMessage()    {}
func (*KVGetMetadataInput) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{5}
}

func (m *KVGetMetadataInput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KVGetMetadataInput.Unmarshal(m, b)
}
func (m *KVGetMetadataInput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KVGetMetadataInput.Marshal(b, m, deterministic)
}
func (m *KVGetMetadataInput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KVGetMetadataInput.Merge(m, src)
}
func (m *KVGetMetadataInput) XXX_Size() int {
	return xxx_messageInfo_KVGetMetadataInput.Size(m)
}
func (m *KVGetMetadataInput) XXX_DiscardUnknown() {
	xxx_messageInfo_KVGetMetadataInput.DiscardUnknown(m)
}

var xxx_messageInfo_KVGetMetadataInput proto.InternalMessageInfo

func (m *KVGetMetadataInput) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

type KVGetMetadataOutput struct {
	Metadata             *KVMetadata `protobuf:"bytes,1,opt,name=Metadata,proto3" json:"Metadata,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *KVGetMetadataOutput) Reset()         { *m = KVGetMetadataOutput{} }
func (m *KVGetMetadataOutput) String() string { return proto.CompactTextString(m) }
func (*KVGetMetadataOutput) ProtoMessage()    {}
func (*KVGetMetadataOutput) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{6}
}

func (m *KVGetMetadataOutput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KVGetMetadataOutput.Unmarshal(m, b)
}
func (m *KVGetMetadataOutput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KVGetMetadataOutput.Marshal(b, m, deterministic)
}
func (m *KVGetMetadataOutput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KVGetMetadataOutput.Merge(m, src)
}
func (m *KVGetMetadataOutput) XXX_Size() int {
	return xxx_messageInfo_KVGetMetadataOutput.Size(m)
}
func (m *KVGetMetadataOutput) XXX_DiscardUnknown() {
	xxx_messageInfo_KVGetMetadataOutput.DiscardUnknown(m)
}

var xxx_messageInfo_KVGetMetadataOutput proto.InternalMessageInfo

func (m *KVGetMetadataOutput) GetMetadata() *KVMetadata {
	if m != nil {
		return m.Metadata
	}
	return nil
}

type KVSetInput struct {
	Key                  []byte   `protobuf:"bytes,1,opt,name=Key,proto3" json:"Key,omitempty"`
	Value                []byte   `protobuf:"bytes,2,opt,name=Value,proto3" json:"Value,omitempty"`
	TimeToLive           uint64   `protobuf:"varint,3,opt,name=TimeToLive,proto3" json:"TimeToLive,omitempty"`
	Version              uint64   `protobuf:"varint,4,opt,name=Version,proto3" json:"Version,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KVSetInput) Reset()         { *m = KVSetInput{} }
func (m *KVSetInput) String() string { return proto.CompactTextString(m) }
func (*KVSetInput) ProtoMessage()    {}
func (*KVSetInput) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{7}
}

func (m *KVSetInput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KVSetInput.Unmarshal(m, b)
}
func (m *KVSetInput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KVSetInput.Marshal(b, m, deterministic)
}
func (m *KVSetInput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KVSetInput.Merge(m, src)
}
func (m *KVSetInput) XXX_Size() int {
	return xxx_messageInfo_KVSetInput.Size(m)
}
func (m *KVSetInput) XXX_DiscardUnknown() {
	xxx_messageInfo_KVSetInput.DiscardUnknown(m)
}

var xxx_messageInfo_KVSetInput proto.InternalMessageInfo

func (m *KVSetInput) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *KVSetInput) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *KVSetInput) GetTimeToLive() uint64 {
	if m != nil {
		return m.TimeToLive
	}
	return 0
}

func (m *KVSetInput) GetVersion() uint64 {
	if m != nil {
		return m.Version
	}
	return 0
}

type KVSetOutput struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KVSetOutput) Reset()         { *m = KVSetOutput{} }
func (m *KVSetOutput) String() string { return proto.CompactTextString(m) }
func (*KVSetOutput) ProtoMessage()    {}
func (*KVSetOutput) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{8}
}

func (m *KVSetOutput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KVSetOutput.Unmarshal(m, b)
}
func (m *KVSetOutput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KVSetOutput.Marshal(b, m, deterministic)
}
func (m *KVSetOutput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KVSetOutput.Merge(m, src)
}
func (m *KVSetOutput) XXX_Size() int {
	return xxx_messageInfo_KVSetOutput.Size(m)
}
func (m *KVSetOutput) XXX_DiscardUnknown() {
	xxx_messageInfo_KVSetOutput.DiscardUnknown(m)
}

var xxx_messageInfo_KVSetOutput proto.InternalMessageInfo

type KVDeleteInput struct {
	Key                  []byte   `protobuf:"bytes,1,opt,name=Key,proto3" json:"Key,omitempty"`
	Version              uint64   `protobuf:"varint,2,opt,name=Version,proto3" json:"Version,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KVDeleteInput) Reset()         { *m = KVDeleteInput{} }
func (m *KVDeleteInput) String() string { return proto.CompactTextString(m) }
func (*KVDeleteInput) ProtoMessage()    {}
func (*KVDeleteInput) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{9}
}

func (m *KVDeleteInput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KVDeleteInput.Unmarshal(m, b)
}
func (m *KVDeleteInput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KVDeleteInput.Marshal(b, m, deterministic)
}
func (m *KVDeleteInput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KVDeleteInput.Merge(m, src)
}
func (m *KVDeleteInput) XXX_Size() int {
	return xxx_messageInfo_KVDeleteInput.Size(m)
}
func (m *KVDeleteInput) XXX_DiscardUnknown() {
	xxx_messageInfo_KVDeleteInput.DiscardUnknown(m)
}

var xxx_messageInfo_KVDeleteInput proto.InternalMessageInfo

func (m *KVDeleteInput) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *KVDeleteInput) GetVersion() uint64 {
	if m != nil {
		return m.Version
	}
	return 0
}

type KVDeleteOutput struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KVDeleteOutput) Reset()         { *m = KVDeleteOutput{} }
func (m *KVDeleteOutput) String() string { return proto.CompactTextString(m) }
func (*KVDeleteOutput) ProtoMessage()    {}
func (*KVDeleteOutput) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{10}
}

func (m *KVDeleteOutput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KVDeleteOutput.Unmarshal(m, b)
}
func (m *KVDeleteOutput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KVDeleteOutput.Marshal(b, m, deterministic)
}
func (m *KVDeleteOutput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KVDeleteOutput.Merge(m, src)
}
func (m *KVDeleteOutput) XXX_Size() int {
	return xxx_messageInfo_KVDeleteOutput.Size(m)
}
func (m *KVDeleteOutput) XXX_DiscardUnknown() {
	xxx_messageInfo_KVDeleteOutput.DiscardUnknown(m)
}

var xxx_messageInfo_KVDeleteOutput proto.InternalMessageInfo

type KVStateTransitionSet struct {
	Events               []*KVStateTransition `protobuf:"bytes,1,rep,name=events,proto3" json:"events,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *KVStateTransitionSet) Reset()         { *m = KVStateTransitionSet{} }
func (m *KVStateTransitionSet) String() string { return proto.CompactTextString(m) }
func (*KVStateTransitionSet) ProtoMessage()    {}
func (*KVStateTransitionSet) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{11}
}

func (m *KVStateTransitionSet) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KVStateTransitionSet.Unmarshal(m, b)
}
func (m *KVStateTransitionSet) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KVStateTransitionSet.Marshal(b, m, deterministic)
}
func (m *KVStateTransitionSet) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KVStateTransitionSet.Merge(m, src)
}
func (m *KVStateTransitionSet) XXX_Size() int {
	return xxx_messageInfo_KVStateTransitionSet.Size(m)
}
func (m *KVStateTransitionSet) XXX_DiscardUnknown() {
	xxx_messageInfo_KVStateTransitionSet.DiscardUnknown(m)
}

var xxx_messageInfo_KVStateTransitionSet proto.InternalMessageInfo

func (m *KVStateTransitionSet) GetEvents() []*KVStateTransition {
	if m != nil {
		return m.Events
	}
	return nil
}

type KVStateTransition struct {
	// Types that are valid to be assigned to Event:
	//	*KVStateTransition_Set
	//	*KVStateTransition_Delete
	//	*KVStateTransition_DeleteBatch
	Event                isKVStateTransition_Event `protobuf_oneof:"Event"`
	XXX_NoUnkeyedLiteral struct{}                  `json:"-"`
	XXX_unrecognized     []byte                    `json:"-"`
	XXX_sizecache        int32                     `json:"-"`
}

func (m *KVStateTransition) Reset()         { *m = KVStateTransition{} }
func (m *KVStateTransition) String() string { return proto.CompactTextString(m) }
func (*KVStateTransition) ProtoMessage()    {}
func (*KVStateTransition) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{12}
}

func (m *KVStateTransition) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KVStateTransition.Unmarshal(m, b)
}
func (m *KVStateTransition) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KVStateTransition.Marshal(b, m, deterministic)
}
func (m *KVStateTransition) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KVStateTransition.Merge(m, src)
}
func (m *KVStateTransition) XXX_Size() int {
	return xxx_messageInfo_KVStateTransition.Size(m)
}
func (m *KVStateTransition) XXX_DiscardUnknown() {
	xxx_messageInfo_KVStateTransition.DiscardUnknown(m)
}

var xxx_messageInfo_KVStateTransition proto.InternalMessageInfo

type isKVStateTransition_Event interface {
	isKVStateTransition_Event()
}

type KVStateTransition_Set struct {
	Set *KVStateTransitionValueSet `protobuf:"bytes,1,opt,name=Set,proto3,oneof"`
}

type KVStateTransition_Delete struct {
	Delete *KVStateTransitionValueDeleted `protobuf:"bytes,2,opt,name=Delete,proto3,oneof"`
}

type KVStateTransition_DeleteBatch struct {
	DeleteBatch *KVStateTransitionValueBatchDeleted `protobuf:"bytes,3,opt,name=DeleteBatch,proto3,oneof"`
}

func (*KVStateTransition_Set) isKVStateTransition_Event() {}

func (*KVStateTransition_Delete) isKVStateTransition_Event() {}

func (*KVStateTransition_DeleteBatch) isKVStateTransition_Event() {}

func (m *KVStateTransition) GetEvent() isKVStateTransition_Event {
	if m != nil {
		return m.Event
	}
	return nil
}

func (m *KVStateTransition) GetSet() *KVStateTransitionValueSet {
	if x, ok := m.GetEvent().(*KVStateTransition_Set); ok {
		return x.Set
	}
	return nil
}

func (m *KVStateTransition) GetDelete() *KVStateTransitionValueDeleted {
	if x, ok := m.GetEvent().(*KVStateTransition_Delete); ok {
		return x.Delete
	}
	return nil
}

func (m *KVStateTransition) GetDeleteBatch() *KVStateTransitionValueBatchDeleted {
	if x, ok := m.GetEvent().(*KVStateTransition_DeleteBatch); ok {
		return x.DeleteBatch
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*KVStateTransition) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*KVStateTransition_Set)(nil),
		(*KVStateTransition_Delete)(nil),
		(*KVStateTransition_DeleteBatch)(nil),
	}
}

type KVStateTransitionValueSet struct {
	Key                  []byte   `protobuf:"bytes,1,opt,name=Key,proto3" json:"Key,omitempty"`
	Value                []byte   `protobuf:"bytes,2,opt,name=Value,proto3" json:"Value,omitempty"`
	Deadline             uint64   `protobuf:"varint,3,opt,name=Deadline,proto3" json:"Deadline,omitempty"`
	Version              uint64   `protobuf:"varint,4,opt,name=Version,proto3" json:"Version,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KVStateTransitionValueSet) Reset()         { *m = KVStateTransitionValueSet{} }
func (m *KVStateTransitionValueSet) String() string { return proto.CompactTextString(m) }
func (*KVStateTransitionValueSet) ProtoMessage()    {}
func (*KVStateTransitionValueSet) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{13}
}

func (m *KVStateTransitionValueSet) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KVStateTransitionValueSet.Unmarshal(m, b)
}
func (m *KVStateTransitionValueSet) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KVStateTransitionValueSet.Marshal(b, m, deterministic)
}
func (m *KVStateTransitionValueSet) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KVStateTransitionValueSet.Merge(m, src)
}
func (m *KVStateTransitionValueSet) XXX_Size() int {
	return xxx_messageInfo_KVStateTransitionValueSet.Size(m)
}
func (m *KVStateTransitionValueSet) XXX_DiscardUnknown() {
	xxx_messageInfo_KVStateTransitionValueSet.DiscardUnknown(m)
}

var xxx_messageInfo_KVStateTransitionValueSet proto.InternalMessageInfo

func (m *KVStateTransitionValueSet) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *KVStateTransitionValueSet) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *KVStateTransitionValueSet) GetDeadline() uint64 {
	if m != nil {
		return m.Deadline
	}
	return 0
}

func (m *KVStateTransitionValueSet) GetVersion() uint64 {
	if m != nil {
		return m.Version
	}
	return 0
}

type KVStateTransitionValueDeleted struct {
	Key                  []byte   `protobuf:"bytes,1,opt,name=Key,proto3" json:"Key,omitempty"`
	Version              uint64   `protobuf:"varint,2,opt,name=Version,proto3" json:"Version,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KVStateTransitionValueDeleted) Reset()         { *m = KVStateTransitionValueDeleted{} }
func (m *KVStateTransitionValueDeleted) String() string { return proto.CompactTextString(m) }
func (*KVStateTransitionValueDeleted) ProtoMessage()    {}
func (*KVStateTransitionValueDeleted) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{14}
}

func (m *KVStateTransitionValueDeleted) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KVStateTransitionValueDeleted.Unmarshal(m, b)
}
func (m *KVStateTransitionValueDeleted) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KVStateTransitionValueDeleted.Marshal(b, m, deterministic)
}
func (m *KVStateTransitionValueDeleted) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KVStateTransitionValueDeleted.Merge(m, src)
}
func (m *KVStateTransitionValueDeleted) XXX_Size() int {
	return xxx_messageInfo_KVStateTransitionValueDeleted.Size(m)
}
func (m *KVStateTransitionValueDeleted) XXX_DiscardUnknown() {
	xxx_messageInfo_KVStateTransitionValueDeleted.DiscardUnknown(m)
}

var xxx_messageInfo_KVStateTransitionValueDeleted proto.InternalMessageInfo

func (m *KVStateTransitionValueDeleted) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *KVStateTransitionValueDeleted) GetVersion() uint64 {
	if m != nil {
		return m.Version
	}
	return 0
}

type KVStateTransitionValueBatchDeleted struct {
	KeyMDs               []*KVMetadata `protobuf:"bytes,1,rep,name=KeyMDs,proto3" json:"KeyMDs,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *KVStateTransitionValueBatchDeleted) Reset()         { *m = KVStateTransitionValueBatchDeleted{} }
func (m *KVStateTransitionValueBatchDeleted) String() string { return proto.CompactTextString(m) }
func (*KVStateTransitionValueBatchDeleted) ProtoMessage()    {}
func (*KVStateTransitionValueBatchDeleted) Descriptor() ([]byte, []int) {
	return fileDescriptor_d938547f84707355, []int{15}
}

func (m *KVStateTransitionValueBatchDeleted) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KVStateTransitionValueBatchDeleted.Unmarshal(m, b)
}
func (m *KVStateTransitionValueBatchDeleted) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KVStateTransitionValueBatchDeleted.Marshal(b, m, deterministic)
}
func (m *KVStateTransitionValueBatchDeleted) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KVStateTransitionValueBatchDeleted.Merge(m, src)
}
func (m *KVStateTransitionValueBatchDeleted) XXX_Size() int {
	return xxx_messageInfo_KVStateTransitionValueBatchDeleted.Size(m)
}
func (m *KVStateTransitionValueBatchDeleted) XXX_DiscardUnknown() {
	xxx_messageInfo_KVStateTransitionValueBatchDeleted.DiscardUnknown(m)
}

var xxx_messageInfo_KVStateTransitionValueBatchDeleted proto.InternalMessageInfo

func (m *KVStateTransitionValueBatchDeleted) GetKeyMDs() []*KVMetadata {
	if m != nil {
		return m.KeyMDs
	}
	return nil
}

func init() {
	proto.RegisterType((*KVMetadata)(nil), "pb.KVMetadata")
	proto.RegisterType((*KVGetInput)(nil), "pb.KVGetInput")
	proto.RegisterType((*KVGetOutput)(nil), "pb.KVGetOutput")
	proto.RegisterType((*KVGetWithMetadataInput)(nil), "pb.KVGetWithMetadataInput")
	proto.RegisterType((*KVGetWithMetadataOutput)(nil), "pb.KVGetWithMetadataOutput")
	proto.RegisterType((*KVGetMetadataInput)(nil), "pb.KVGetMetadataInput")
	proto.RegisterType((*KVGetMetadataOutput)(nil), "pb.KVGetMetadataOutput")
	proto.RegisterType((*KVSetInput)(nil), "pb.KVSetInput")
	proto.RegisterType((*KVSetOutput)(nil), "pb.KVSetOutput")
	proto.RegisterType((*KVDeleteInput)(nil), "pb.KVDeleteInput")
	proto.RegisterType((*KVDeleteOutput)(nil), "pb.KVDeleteOutput")
	proto.RegisterType((*KVStateTransitionSet)(nil), "pb.KVStateTransitionSet")
	proto.RegisterType((*KVStateTransition)(nil), "pb.KVStateTransition")
	proto.RegisterType((*KVStateTransitionValueSet)(nil), "pb.KVStateTransitionValueSet")
	proto.RegisterType((*KVStateTransitionValueDeleted)(nil), "pb.KVStateTransitionValueDeleted")
	proto.RegisterType((*KVStateTransitionValueBatchDeleted)(nil), "pb.KVStateTransitionValueBatchDeleted")
}

func init() { proto.RegisterFile("types.proto", fileDescriptor_d938547f84707355) }

var fileDescriptor_d938547f84707355 = []byte{
	// 565 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x54, 0xdb, 0x6e, 0xda, 0x40,
	0x10, 0x8d, 0x71, 0x42, 0xd2, 0x71, 0xae, 0xdb, 0x34, 0xa1, 0xae, 0x12, 0x51, 0x3f, 0x20, 0x14,
	0x09, 0x68, 0xa9, 0xfa, 0x14, 0xa9, 0x52, 0x23, 0xa2, 0xa4, 0x71, 0xa2, 0x4a, 0x36, 0xa2, 0xcf,
	0x6b, 0x18, 0x15, 0xab, 0x60, 0xbb, 0x30, 0xa0, 0xf2, 0x27, 0xfd, 0xb4, 0x7e, 0x4e, 0xe5, 0xf5,
	0xae, 0x31, 0x17, 0x97, 0xe6, 0xcd, 0xeb, 0x39, 0xe7, 0xcc, 0x99, 0x99, 0x9d, 0x05, 0x83, 0x66,
	0x11, 0x8e, 0xeb, 0xd1, 0x28, 0xa4, 0x90, 0x15, 0x22, 0xcf, 0x7c, 0xf7, 0xdd, 0xa7, 0xfe, 0xc4,
	0xab, 0x77, 0xc3, 0x61, 0x63, 0xfa, 0xab, 0x36, 0xe0, 0xde, 0xb8, 0x31, 0xfc, 0x49, 0x54, 0x13,
	0x90, 0x6e, 0x38, 0x68, 0x44, 0xbc, 0xfb, 0x03, 0xa9, 0x11, 0x79, 0x09, 0xcb, 0x6a, 0x03, 0xd8,
	0x9d, 0x27, 0x24, 0xde, 0xe3, 0xc4, 0x59, 0x09, 0x76, 0x3b, 0x38, 0x1a, 0xfb, 0x61, 0x50, 0xd2,
	0xca, 0x5a, 0x75, 0xdb, 0x51, 0x47, 0x66, 0xc2, 0x5e, 0x0b, 0x79, 0x6f, 0xe0, 0x07, 0x58, 0x2a,
	0x88, 0x50, 0x7a, 0x66, 0xc7, 0xa0, 0xdb, 0x38, 0x2b, 0xe9, 0x65, 0xad, 0xba, 0xef, 0xc4, 0x9f,
	0xd6, 0x65, 0xac, 0x7a, 0x87, 0xf4, 0x25, 0x88, 0x26, 0xa4, 0xe2, 0xda, 0x3c, 0xfe, 0x11, 0x0c,
	0x11, 0xff, 0x3a, 0xa1, 0xb5, 0x00, 0x76, 0x0a, 0x3b, 0x1d, 0x3e, 0x98, 0x24, 0xb9, 0xf6, 0x9d,
	0xe4, 0x60, 0x5d, 0xc1, 0x99, 0xa0, 0x7d, 0xf3, 0xa9, 0xaf, 0x3c, 0xe7, 0xa5, 0x18, 0xc2, 0xf9,
	0x0a, 0xf6, 0x79, 0xe9, 0xd8, 0x15, 0xec, 0x29, 0xa6, 0x28, 0xce, 0x68, 0x1e, 0xd6, 0x23, 0xaf,
	0x3e, 0xef, 0x97, 0x93, 0xc6, 0xad, 0x0a, 0x30, 0x91, 0x6e, 0x93, 0xad, 0xcf, 0xf0, 0x72, 0x01,
	0x27, 0x2d, 0x65, 0x53, 0x69, 0x1b, 0x52, 0x05, 0x71, 0x73, 0xdd, 0xdc, 0xe6, 0xe6, 0x14, 0x73,
	0x09, 0xd0, 0xf6, 0x87, 0xd8, 0x0e, 0x1f, 0xfd, 0x29, 0x8a, 0x72, 0xb6, 0x9d, 0xcc, 0x9f, 0xec,
	0xe8, 0xb7, 0x17, 0x46, 0x6f, 0x1d, 0xc4, 0xc3, 0x72, 0xd5, 0xb0, 0xac, 0x6b, 0x38, 0xb0, 0x3b,
	0x2d, 0x1c, 0x20, 0x61, 0x9e, 0x83, 0x8c, 0x56, 0x61, 0x51, 0xeb, 0x18, 0x0e, 0x15, 0x59, 0xca,
	0xdd, 0xc2, 0xa9, 0xdd, 0x71, 0x89, 0x13, 0xb6, 0x47, 0x3c, 0x18, 0xfb, 0xe4, 0x87, 0x81, 0x8b,
	0xc4, 0x6a, 0x50, 0xc4, 0x29, 0x06, 0x34, 0x2e, 0x69, 0x65, 0xbd, 0x6a, 0x34, 0x5f, 0x25, 0xfd,
	0x58, 0x42, 0x3a, 0x12, 0x64, 0xfd, 0xd1, 0xe0, 0x64, 0x25, 0xca, 0xde, 0x83, 0xee, 0x22, 0xc9,
	0x8e, 0x5e, 0xac, 0x55, 0x10, 0xdd, 0x71, 0x91, 0xee, 0xb7, 0x9c, 0x18, 0xcb, 0xae, 0xa1, 0x98,
	0xf8, 0x13, 0xd6, 0x8d, 0xe6, 0xdb, 0x7c, 0x56, 0x82, 0xeb, 0xdd, 0x6f, 0x39, 0x92, 0xc2, 0x1e,
	0xc0, 0x48, 0xbe, 0x6e, 0x38, 0x75, 0xfb, 0xf2, 0xd2, 0x54, 0xf2, 0x15, 0x04, 0x6c, 0x2e, 0x93,
	0x25, 0xdf, 0xec, 0xc2, 0xce, 0x6d, 0x5c, 0x9b, 0x35, 0x83, 0xd7, 0xb9, 0xae, 0xff, 0x7b, 0xfc,
	0xd9, 0xfd, 0xd5, 0x97, 0xf6, 0x37, 0x7f, 0xf4, 0x36, 0x5c, 0xfc, 0xb3, 0xf4, 0x67, 0xcd, 0xfe,
	0x11, 0xac, 0xcd, 0x5d, 0x60, 0x15, 0x28, 0xda, 0x38, 0x7b, 0x6a, 0xa9, 0xb9, 0x2f, 0xef, 0x81,
	0x8c, 0x36, 0x7f, 0x17, 0xe0, 0x45, 0x7c, 0x2d, 0x47, 0x53, 0xbf, 0x8b, 0xac, 0x02, 0xfa, 0x1d,
	0x12, 0x93, 0x60, 0xf5, 0xf2, 0x98, 0x47, 0xe9, 0x59, 0xee, 0xd9, 0x03, 0x1c, 0x2d, 0xbd, 0x09,
	0xcc, 0x4c, 0x31, 0x2b, 0xcf, 0x8a, 0xf9, 0x66, 0x6d, 0x4c, 0x6a, 0x7d, 0x02, 0x23, 0xb3, 0xc8,
	0xec, 0x2c, 0xc5, 0x2e, 0x6a, 0x9c, 0xaf, 0xfc, 0x97, 0xfc, 0x8a, 0xb8, 0x9c, 0xca, 0xb3, 0xbb,
	0xe4, 0x39, 0x5d, 0x38, 0xd6, 0x50, 0x37, 0x92, 0x9d, 0x24, 0xa1, 0xcc, 0xf2, 0x99, 0x2c, 0xfb,
	0x2b, 0x21, 0x78, 0x45, 0xf1, 0xb4, 0x7f, 0xf8, 0x1b, 0x00, 0x00, 0xff, 0xff, 0x97, 0x72, 0x94,
	0x3c, 0x1f, 0x06, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// KVServiceClient is the client API for KVService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type KVServiceClient interface {
	Get(ctx context.Context, in *KVGetInput, opts ...grpc.CallOption) (*KVGetOutput, error)
	GetWithMetadata(ctx context.Context, in *KVGetWithMetadataInput, opts ...grpc.CallOption) (*KVGetWithMetadataOutput, error)
	GetMetadata(ctx context.Context, in *KVGetMetadataInput, opts ...grpc.CallOption) (*KVGetMetadataOutput, error)
	Set(ctx context.Context, in *KVSetInput, opts ...grpc.CallOption) (*KVSetOutput, error)
	Delete(ctx context.Context, in *KVDeleteInput, opts ...grpc.CallOption) (*KVDeleteOutput, error)
}

type kVServiceClient struct {
	cc *grpc.ClientConn
}

func NewKVServiceClient(cc *grpc.ClientConn) KVServiceClient {
	return &kVServiceClient{cc}
}

func (c *kVServiceClient) Get(ctx context.Context, in *KVGetInput, opts ...grpc.CallOption) (*KVGetOutput, error) {
	out := new(KVGetOutput)
	err := c.cc.Invoke(ctx, "/pb.KVService/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kVServiceClient) GetWithMetadata(ctx context.Context, in *KVGetWithMetadataInput, opts ...grpc.CallOption) (*KVGetWithMetadataOutput, error) {
	out := new(KVGetWithMetadataOutput)
	err := c.cc.Invoke(ctx, "/pb.KVService/GetWithMetadata", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kVServiceClient) GetMetadata(ctx context.Context, in *KVGetMetadataInput, opts ...grpc.CallOption) (*KVGetMetadataOutput, error) {
	out := new(KVGetMetadataOutput)
	err := c.cc.Invoke(ctx, "/pb.KVService/GetMetadata", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kVServiceClient) Set(ctx context.Context, in *KVSetInput, opts ...grpc.CallOption) (*KVSetOutput, error) {
	out := new(KVSetOutput)
	err := c.cc.Invoke(ctx, "/pb.KVService/Set", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kVServiceClient) Delete(ctx context.Context, in *KVDeleteInput, opts ...grpc.CallOption) (*KVDeleteOutput, error) {
	out := new(KVDeleteOutput)
	err := c.cc.Invoke(ctx, "/pb.KVService/Delete", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// KVServiceServer is the server API for KVService service.
type KVServiceServer interface {
	Get(context.Context, *KVGetInput) (*KVGetOutput, error)
	GetWithMetadata(context.Context, *KVGetWithMetadataInput) (*KVGetWithMetadataOutput, error)
	GetMetadata(context.Context, *KVGetMetadataInput) (*KVGetMetadataOutput, error)
	Set(context.Context, *KVSetInput) (*KVSetOutput, error)
	Delete(context.Context, *KVDeleteInput) (*KVDeleteOutput, error)
}

// UnimplementedKVServiceServer can be embedded to have forward compatible implementations.
type UnimplementedKVServiceServer struct {
}

func (*UnimplementedKVServiceServer) Get(ctx context.Context, req *KVGetInput) (*KVGetOutput, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (*UnimplementedKVServiceServer) GetWithMetadata(ctx context.Context, req *KVGetWithMetadataInput) (*KVGetWithMetadataOutput, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetWithMetadata not implemented")
}
func (*UnimplementedKVServiceServer) GetMetadata(ctx context.Context, req *KVGetMetadataInput) (*KVGetMetadataOutput, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMetadata not implemented")
}
func (*UnimplementedKVServiceServer) Set(ctx context.Context, req *KVSetInput) (*KVSetOutput, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Set not implemented")
}
func (*UnimplementedKVServiceServer) Delete(ctx context.Context, req *KVDeleteInput) (*KVDeleteOutput, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}

func RegisterKVServiceServer(s *grpc.Server, srv KVServiceServer) {
	s.RegisterService(&_KVService_serviceDesc, srv)
}

func _KVService_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(KVGetInput)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KVServiceServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.KVService/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KVServiceServer).Get(ctx, req.(*KVGetInput))
	}
	return interceptor(ctx, in, info, handler)
}

func _KVService_GetWithMetadata_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(KVGetWithMetadataInput)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KVServiceServer).GetWithMetadata(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.KVService/GetWithMetadata",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KVServiceServer).GetWithMetadata(ctx, req.(*KVGetWithMetadataInput))
	}
	return interceptor(ctx, in, info, handler)
}

func _KVService_GetMetadata_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(KVGetMetadataInput)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KVServiceServer).GetMetadata(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.KVService/GetMetadata",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KVServiceServer).GetMetadata(ctx, req.(*KVGetMetadataInput))
	}
	return interceptor(ctx, in, info, handler)
}

func _KVService_Set_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(KVSetInput)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KVServiceServer).Set(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.KVService/Set",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KVServiceServer).Set(ctx, req.(*KVSetInput))
	}
	return interceptor(ctx, in, info, handler)
}

func _KVService_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(KVDeleteInput)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KVServiceServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.KVService/Delete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KVServiceServer).Delete(ctx, req.(*KVDeleteInput))
	}
	return interceptor(ctx, in, info, handler)
}

var _KVService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "pb.KVService",
	HandlerType: (*KVServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Get",
			Handler:    _KVService_Get_Handler,
		},
		{
			MethodName: "GetWithMetadata",
			Handler:    _KVService_GetWithMetadata_Handler,
		},
		{
			MethodName: "GetMetadata",
			Handler:    _KVService_GetMetadata_Handler,
		},
		{
			MethodName: "Set",
			Handler:    _KVService_Set_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _KVService_Delete_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "types.proto",
}