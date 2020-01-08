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

type TransportContext struct {
	Encrypted            bool     `protobuf:"varint,1,opt,name=encrypted,proto3" json:"encrypted,omitempty"`
	RemoteAddress        string   `protobuf:"bytes,2,opt,name=remoteAddress,proto3" json:"remoteAddress,omitempty"`
	X509Certificate      []byte   `protobuf:"bytes,3,opt,name=x509Certificate,proto3" json:"x509Certificate,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TransportContext) Reset()         { *m = TransportContext{} }
func (m *TransportContext) String() string { return proto.CompactTextString(m) }
func (*TransportContext) ProtoMessage()    {}
func (*TransportContext) Descriptor() ([]byte, []int) {
	return fileDescriptor_f80abaa17e25ccc8, []int{0}
}

func (m *TransportContext) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TransportContext.Unmarshal(m, b)
}
func (m *TransportContext) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TransportContext.Marshal(b, m, deterministic)
}
func (m *TransportContext) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TransportContext.Merge(m, src)
}
func (m *TransportContext) XXX_Size() int {
	return xxx_messageInfo_TransportContext.Size(m)
}
func (m *TransportContext) XXX_DiscardUnknown() {
	xxx_messageInfo_TransportContext.DiscardUnknown(m)
}

var xxx_messageInfo_TransportContext proto.InternalMessageInfo

func (m *TransportContext) GetEncrypted() bool {
	if m != nil {
		return m.Encrypted
	}
	return false
}

func (m *TransportContext) GetRemoteAddress() string {
	if m != nil {
		return m.RemoteAddress
	}
	return ""
}

func (m *TransportContext) GetX509Certificate() []byte {
	if m != nil {
		return m.X509Certificate
	}
	return nil
}

type ProtocolContext struct {
	Username             string   `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	Password             string   `protobuf:"bytes,2,opt,name=password,proto3" json:"password,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ProtocolContext) Reset()         { *m = ProtocolContext{} }
func (m *ProtocolContext) String() string { return proto.CompactTextString(m) }
func (*ProtocolContext) ProtoMessage()    {}
func (*ProtocolContext) Descriptor() ([]byte, []int) {
	return fileDescriptor_f80abaa17e25ccc8, []int{1}
}

func (m *ProtocolContext) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ProtocolContext.Unmarshal(m, b)
}
func (m *ProtocolContext) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ProtocolContext.Marshal(b, m, deterministic)
}
func (m *ProtocolContext) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProtocolContext.Merge(m, src)
}
func (m *ProtocolContext) XXX_Size() int {
	return xxx_messageInfo_ProtocolContext.Size(m)
}
func (m *ProtocolContext) XXX_DiscardUnknown() {
	xxx_messageInfo_ProtocolContext.DiscardUnknown(m)
}

var xxx_messageInfo_ProtocolContext proto.InternalMessageInfo

func (m *ProtocolContext) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

func (m *ProtocolContext) GetPassword() string {
	if m != nil {
		return m.Password
	}
	return ""
}

type CreateTokenInput struct {
	Protocol             *ProtocolContext  `protobuf:"bytes,1,opt,name=protocol,proto3" json:"protocol,omitempty"`
	Transport            *TransportContext `protobuf:"bytes,2,opt,name=transport,proto3" json:"transport,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *CreateTokenInput) Reset()         { *m = CreateTokenInput{} }
func (m *CreateTokenInput) String() string { return proto.CompactTextString(m) }
func (*CreateTokenInput) ProtoMessage()    {}
func (*CreateTokenInput) Descriptor() ([]byte, []int) {
	return fileDescriptor_f80abaa17e25ccc8, []int{2}
}

func (m *CreateTokenInput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateTokenInput.Unmarshal(m, b)
}
func (m *CreateTokenInput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateTokenInput.Marshal(b, m, deterministic)
}
func (m *CreateTokenInput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateTokenInput.Merge(m, src)
}
func (m *CreateTokenInput) XXX_Size() int {
	return xxx_messageInfo_CreateTokenInput.Size(m)
}
func (m *CreateTokenInput) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateTokenInput.DiscardUnknown(m)
}

var xxx_messageInfo_CreateTokenInput proto.InternalMessageInfo

func (m *CreateTokenInput) GetProtocol() *ProtocolContext {
	if m != nil {
		return m.Protocol
	}
	return nil
}

func (m *CreateTokenInput) GetTransport() *TransportContext {
	if m != nil {
		return m.Transport
	}
	return nil
}

type CreateTokenOutput struct {
	JWT                  string   `protobuf:"bytes,1,opt,name=JWT,proto3" json:"JWT,omitempty"`
	Tenant               string   `protobuf:"bytes,2,opt,name=Tenant,proto3" json:"Tenant,omitempty"`
	EntityID             string   `protobuf:"bytes,3,opt,name=EntityID,proto3" json:"EntityID,omitempty"`
	SessionID            string   `protobuf:"bytes,4,opt,name=SessionID,proto3" json:"SessionID,omitempty"`
	RefreshToken         string   `protobuf:"bytes,5,opt,name=RefreshToken,proto3" json:"RefreshToken,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CreateTokenOutput) Reset()         { *m = CreateTokenOutput{} }
func (m *CreateTokenOutput) String() string { return proto.CompactTextString(m) }
func (*CreateTokenOutput) ProtoMessage()    {}
func (*CreateTokenOutput) Descriptor() ([]byte, []int) {
	return fileDescriptor_f80abaa17e25ccc8, []int{3}
}

func (m *CreateTokenOutput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateTokenOutput.Unmarshal(m, b)
}
func (m *CreateTokenOutput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateTokenOutput.Marshal(b, m, deterministic)
}
func (m *CreateTokenOutput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateTokenOutput.Merge(m, src)
}
func (m *CreateTokenOutput) XXX_Size() int {
	return xxx_messageInfo_CreateTokenOutput.Size(m)
}
func (m *CreateTokenOutput) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateTokenOutput.DiscardUnknown(m)
}

var xxx_messageInfo_CreateTokenOutput proto.InternalMessageInfo

func (m *CreateTokenOutput) GetJWT() string {
	if m != nil {
		return m.JWT
	}
	return ""
}

func (m *CreateTokenOutput) GetTenant() string {
	if m != nil {
		return m.Tenant
	}
	return ""
}

func (m *CreateTokenOutput) GetEntityID() string {
	if m != nil {
		return m.EntityID
	}
	return ""
}

func (m *CreateTokenOutput) GetSessionID() string {
	if m != nil {
		return m.SessionID
	}
	return ""
}

func (m *CreateTokenOutput) GetRefreshToken() string {
	if m != nil {
		return m.RefreshToken
	}
	return ""
}

type RefreshTokenInput struct {
	RefreshToken         string   `protobuf:"bytes,1,opt,name=RefreshToken,proto3" json:"RefreshToken,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RefreshTokenInput) Reset()         { *m = RefreshTokenInput{} }
func (m *RefreshTokenInput) String() string { return proto.CompactTextString(m) }
func (*RefreshTokenInput) ProtoMessage()    {}
func (*RefreshTokenInput) Descriptor() ([]byte, []int) {
	return fileDescriptor_f80abaa17e25ccc8, []int{4}
}

func (m *RefreshTokenInput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RefreshTokenInput.Unmarshal(m, b)
}
func (m *RefreshTokenInput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RefreshTokenInput.Marshal(b, m, deterministic)
}
func (m *RefreshTokenInput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RefreshTokenInput.Merge(m, src)
}
func (m *RefreshTokenInput) XXX_Size() int {
	return xxx_messageInfo_RefreshTokenInput.Size(m)
}
func (m *RefreshTokenInput) XXX_DiscardUnknown() {
	xxx_messageInfo_RefreshTokenInput.DiscardUnknown(m)
}

var xxx_messageInfo_RefreshTokenInput proto.InternalMessageInfo

func (m *RefreshTokenInput) GetRefreshToken() string {
	if m != nil {
		return m.RefreshToken
	}
	return ""
}

type RefreshTokenOutput struct {
	JWT                  string   `protobuf:"bytes,1,opt,name=JWT,proto3" json:"JWT,omitempty"`
	Tenant               string   `protobuf:"bytes,2,opt,name=Tenant,proto3" json:"Tenant,omitempty"`
	EntityID             string   `protobuf:"bytes,3,opt,name=EntityID,proto3" json:"EntityID,omitempty"`
	SessionID            string   `protobuf:"bytes,4,opt,name=SessionID,proto3" json:"SessionID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RefreshTokenOutput) Reset()         { *m = RefreshTokenOutput{} }
func (m *RefreshTokenOutput) String() string { return proto.CompactTextString(m) }
func (*RefreshTokenOutput) ProtoMessage()    {}
func (*RefreshTokenOutput) Descriptor() ([]byte, []int) {
	return fileDescriptor_f80abaa17e25ccc8, []int{5}
}

func (m *RefreshTokenOutput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RefreshTokenOutput.Unmarshal(m, b)
}
func (m *RefreshTokenOutput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RefreshTokenOutput.Marshal(b, m, deterministic)
}
func (m *RefreshTokenOutput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RefreshTokenOutput.Merge(m, src)
}
func (m *RefreshTokenOutput) XXX_Size() int {
	return xxx_messageInfo_RefreshTokenOutput.Size(m)
}
func (m *RefreshTokenOutput) XXX_DiscardUnknown() {
	xxx_messageInfo_RefreshTokenOutput.DiscardUnknown(m)
}

var xxx_messageInfo_RefreshTokenOutput proto.InternalMessageInfo

func (m *RefreshTokenOutput) GetJWT() string {
	if m != nil {
		return m.JWT
	}
	return ""
}

func (m *RefreshTokenOutput) GetTenant() string {
	if m != nil {
		return m.Tenant
	}
	return ""
}

func (m *RefreshTokenOutput) GetEntityID() string {
	if m != nil {
		return m.EntityID
	}
	return ""
}

func (m *RefreshTokenOutput) GetSessionID() string {
	if m != nil {
		return m.SessionID
	}
	return ""
}

type Entity struct {
	ID                   string   `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	Tenant               string   `protobuf:"bytes,2,opt,name=Tenant,proto3" json:"Tenant,omitempty"`
	AuthProvider         string   `protobuf:"bytes,3,opt,name=AuthProvider,proto3" json:"AuthProvider,omitempty"`
	Permissions          []string `protobuf:"bytes,4,rep,name=permissions,proto3" json:"permissions,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Entity) Reset()         { *m = Entity{} }
func (m *Entity) String() string { return proto.CompactTextString(m) }
func (*Entity) ProtoMessage()    {}
func (*Entity) Descriptor() ([]byte, []int) {
	return fileDescriptor_f80abaa17e25ccc8, []int{6}
}

func (m *Entity) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Entity.Unmarshal(m, b)
}
func (m *Entity) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Entity.Marshal(b, m, deterministic)
}
func (m *Entity) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Entity.Merge(m, src)
}
func (m *Entity) XXX_Size() int {
	return xxx_messageInfo_Entity.Size(m)
}
func (m *Entity) XXX_DiscardUnknown() {
	xxx_messageInfo_Entity.DiscardUnknown(m)
}

var xxx_messageInfo_Entity proto.InternalMessageInfo

func (m *Entity) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *Entity) GetTenant() string {
	if m != nil {
		return m.Tenant
	}
	return ""
}

func (m *Entity) GetAuthProvider() string {
	if m != nil {
		return m.AuthProvider
	}
	return ""
}

func (m *Entity) GetPermissions() []string {
	if m != nil {
		return m.Permissions
	}
	return nil
}

func init() {
	proto.RegisterType((*TransportContext)(nil), "TransportContext")
	proto.RegisterType((*ProtocolContext)(nil), "ProtocolContext")
	proto.RegisterType((*CreateTokenInput)(nil), "CreateTokenInput")
	proto.RegisterType((*CreateTokenOutput)(nil), "CreateTokenOutput")
	proto.RegisterType((*RefreshTokenInput)(nil), "RefreshTokenInput")
	proto.RegisterType((*RefreshTokenOutput)(nil), "RefreshTokenOutput")
	proto.RegisterType((*Entity)(nil), "Entity")
}

func init() { proto.RegisterFile("pb.proto", fileDescriptor_f80abaa17e25ccc8) }

var fileDescriptor_f80abaa17e25ccc8 = []byte{
	// 421 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xbc, 0x92, 0x41, 0x6f, 0xd3, 0x40,
	0x10, 0x85, 0xe5, 0xa4, 0x44, 0xf1, 0x38, 0xd0, 0x64, 0x91, 0x90, 0x15, 0x71, 0xb0, 0x56, 0x1c,
	0x7c, 0x40, 0x06, 0x05, 0x50, 0xc5, 0xb1, 0x4a, 0x38, 0x98, 0x0b, 0xd5, 0x36, 0x12, 0x67, 0x37,
	0x9e, 0xaa, 0x16, 0x64, 0x77, 0xd9, 0x1d, 0x87, 0x54, 0xe2, 0xc0, 0xff, 0xe0, 0xcf, 0x22, 0x6f,
	0x1d, 0xc7, 0x76, 0xe0, 0xda, 0x9b, 0xdf, 0xf3, 0xec, 0x7c, 0xb3, 0x6f, 0x16, 0xc6, 0xfa, 0x26,
	0xd1, 0x46, 0x91, 0xe2, 0xbf, 0x3d, 0x98, 0xae, 0x4d, 0x26, 0xad, 0x56, 0x86, 0x96, 0x4a, 0x12,
	0xee, 0x89, 0xbd, 0x04, 0x1f, 0xe5, 0xc6, 0xdc, 0x6b, 0xc2, 0x3c, 0xf4, 0x22, 0x2f, 0x1e, 0x8b,
	0xa3, 0xc1, 0x5e, 0xc1, 0x53, 0x83, 0x5b, 0x45, 0x78, 0x99, 0xe7, 0x06, 0xad, 0x0d, 0x07, 0x91,
	0x17, 0xfb, 0xa2, 0x6b, 0xb2, 0x18, 0xce, 0xf7, 0x1f, 0xde, 0x7e, 0x5c, 0xa2, 0xa1, 0xe2, 0xb6,
	0xd8, 0x64, 0x84, 0xe1, 0x30, 0xf2, 0xe2, 0x89, 0xe8, 0xdb, 0x3c, 0x85, 0xf3, 0xab, 0x6a, 0x96,
	0x8d, 0xfa, 0x7e, 0x18, 0x60, 0x0e, 0xe3, 0xd2, 0xa2, 0x91, 0xd9, 0x16, 0x1d, 0xdf, 0x17, 0x8d,
	0xae, 0xfe, 0xe9, 0xcc, 0xda, 0x9f, 0xca, 0xe4, 0x35, 0xb9, 0xd1, 0xfc, 0x07, 0x4c, 0x97, 0x06,
	0x33, 0xc2, 0xb5, 0xfa, 0x86, 0x32, 0x95, 0xba, 0x24, 0xf6, 0x1a, 0xc6, 0xba, 0x6e, 0xef, 0x7a,
	0x05, 0x8b, 0x69, 0xd2, 0xe3, 0x89, 0xa6, 0x82, 0xbd, 0x01, 0x9f, 0x0e, 0x71, 0xb8, 0xf6, 0xc1,
	0x62, 0x96, 0xf4, 0x03, 0x12, 0xc7, 0x1a, 0xfe, 0xc7, 0x83, 0x59, 0x8b, 0xf9, 0xa5, 0xa4, 0x0a,
	0x3a, 0x85, 0xe1, 0xe7, 0xaf, 0xeb, 0x7a, 0xf6, 0xea, 0x93, 0xbd, 0x80, 0xd1, 0x1a, 0x65, 0x26,
	0xa9, 0x1e, 0xba, 0x56, 0xd5, 0x75, 0x3e, 0x49, 0x2a, 0xe8, 0x3e, 0x5d, 0xb9, 0x80, 0x7c, 0xd1,
	0xe8, 0x6a, 0x0f, 0xd7, 0x68, 0x6d, 0xa1, 0x64, 0xba, 0x0a, 0xcf, 0xdc, 0xcf, 0xa3, 0xc1, 0x38,
	0x4c, 0x04, 0xde, 0x1a, 0xb4, 0x77, 0x8e, 0x1c, 0x3e, 0x71, 0x05, 0x1d, 0x8f, 0x5f, 0xc0, 0xac,
	0xad, 0x1f, 0x12, 0xe9, 0x1f, 0xf4, 0xfe, 0x71, 0x70, 0x0f, 0xac, 0xad, 0x1f, 0xef, 0x5a, 0x7c,
	0x07, 0xa3, 0x87, 0x4a, 0xf6, 0x0c, 0x06, 0xe9, 0xaa, 0x86, 0x0d, 0xd2, 0xd5, 0x7f, 0x59, 0x1c,
	0x26, 0x97, 0x25, 0xdd, 0x5d, 0x19, 0xb5, 0x2b, 0x72, 0x34, 0x35, 0xaf, 0xe3, 0xb1, 0x08, 0x02,
	0x8d, 0x66, 0x5b, 0x38, 0x8a, 0x0d, 0xcf, 0xa2, 0x61, 0xec, 0x8b, 0xb6, 0xb5, 0xf8, 0x05, 0x41,
	0x75, 0xe2, 0x1a, 0xcd, 0xae, 0xd8, 0x20, 0x7b, 0x0f, 0x41, 0x6b, 0xad, 0x6c, 0x96, 0xf4, 0x1f,
	0xd6, 0x9c, 0x25, 0xa7, 0x7b, 0xbf, 0xe8, 0x46, 0xcb, 0x58, 0x72, 0x12, 0xff, 0xfc, 0x79, 0x72,
	0x9a, 0xec, 0xcd, 0xc8, 0xbd, 0xc0, 0x77, 0x7f, 0x03, 0x00, 0x00, 0xff, 0xff, 0xc3, 0x31, 0x4e,
	0xeb, 0x9a, 0x03, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// AuthServiceClient is the client API for AuthService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type AuthServiceClient interface {
	CreateToken(ctx context.Context, in *CreateTokenInput, opts ...grpc.CallOption) (*CreateTokenOutput, error)
	RefreshToken(ctx context.Context, in *RefreshTokenInput, opts ...grpc.CallOption) (*RefreshTokenOutput, error)
}

type authServiceClient struct {
	cc *grpc.ClientConn
}

func NewAuthServiceClient(cc *grpc.ClientConn) AuthServiceClient {
	return &authServiceClient{cc}
}

func (c *authServiceClient) CreateToken(ctx context.Context, in *CreateTokenInput, opts ...grpc.CallOption) (*CreateTokenOutput, error) {
	out := new(CreateTokenOutput)
	err := c.cc.Invoke(ctx, "/AuthService/CreateToken", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authServiceClient) RefreshToken(ctx context.Context, in *RefreshTokenInput, opts ...grpc.CallOption) (*RefreshTokenOutput, error) {
	out := new(RefreshTokenOutput)
	err := c.cc.Invoke(ctx, "/AuthService/RefreshToken", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AuthServiceServer is the server API for AuthService service.
type AuthServiceServer interface {
	CreateToken(context.Context, *CreateTokenInput) (*CreateTokenOutput, error)
	RefreshToken(context.Context, *RefreshTokenInput) (*RefreshTokenOutput, error)
}

// UnimplementedAuthServiceServer can be embedded to have forward compatible implementations.
type UnimplementedAuthServiceServer struct {
}

func (*UnimplementedAuthServiceServer) CreateToken(ctx context.Context, req *CreateTokenInput) (*CreateTokenOutput, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateToken not implemented")
}
func (*UnimplementedAuthServiceServer) RefreshToken(ctx context.Context, req *RefreshTokenInput) (*RefreshTokenOutput, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RefreshToken not implemented")
}

func RegisterAuthServiceServer(s *grpc.Server, srv AuthServiceServer) {
	s.RegisterService(&_AuthService_serviceDesc, srv)
}

func _AuthService_CreateToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateTokenInput)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).CreateToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/AuthService/CreateToken",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).CreateToken(ctx, req.(*CreateTokenInput))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthService_RefreshToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RefreshTokenInput)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).RefreshToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/AuthService/RefreshToken",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).RefreshToken(ctx, req.(*RefreshTokenInput))
	}
	return interceptor(ctx, in, info, handler)
}

var _AuthService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "AuthService",
	HandlerType: (*AuthServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateToken",
			Handler:    _AuthService_CreateToken_Handler,
		},
		{
			MethodName: "RefreshToken",
			Handler:    _AuthService_RefreshToken_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pb.proto",
}
