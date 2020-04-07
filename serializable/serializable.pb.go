// Code generated by protoc-gen-go. DO NOT EDIT.
// source: serializable.proto

package serializable

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	any "github.com/golang/protobuf/ptypes/any"
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

type Packet_OpCode int32

const (
	Packet_Invalid Packet_OpCode = 0
	// connection related
	Packet_ClientDisconnect Packet_OpCode = 100
	// game server related
	Packet_SeatConfiguration Packet_OpCode = 200
	Packet_Seat              Packet_OpCode = 201
	Packet_RunSimulation     Packet_OpCode = 202
	Packet_ReloadSimulation  Packet_OpCode = 203
)

var Packet_OpCode_name = map[int32]string{
	0:   "Invalid",
	100: "ClientDisconnect",
	200: "SeatConfiguration",
	201: "Seat",
	202: "RunSimulation",
	203: "ReloadSimulation",
}

var Packet_OpCode_value = map[string]int32{
	"Invalid":           0,
	"ClientDisconnect":  100,
	"SeatConfiguration": 200,
	"Seat":              201,
	"RunSimulation":     202,
	"ReloadSimulation":  203,
}

func (x Packet_OpCode) String() string {
	return proto.EnumName(Packet_OpCode_name, int32(x))
}

func (Packet_OpCode) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_cd0c84d25a72589c, []int{0, 0}
}

type Packet struct {
	OpCode               Packet_OpCode `protobuf:"varint,1,opt,name=opCode,proto3,enum=serializable.Packet_OpCode" json:"opCode,omitempty"`
	Data                 *any.Any      `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *Packet) Reset()         { *m = Packet{} }
func (m *Packet) String() string { return proto.CompactTextString(m) }
func (*Packet) ProtoMessage()    {}
func (*Packet) Descriptor() ([]byte, []int) {
	return fileDescriptor_cd0c84d25a72589c, []int{0}
}

func (m *Packet) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Packet.Unmarshal(m, b)
}
func (m *Packet) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Packet.Marshal(b, m, deterministic)
}
func (m *Packet) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Packet.Merge(m, src)
}
func (m *Packet) XXX_Size() int {
	return xxx_messageInfo_Packet.Size(m)
}
func (m *Packet) XXX_DiscardUnknown() {
	xxx_messageInfo_Packet.DiscardUnknown(m)
}

var xxx_messageInfo_Packet proto.InternalMessageInfo

func (m *Packet) GetOpCode() Packet_OpCode {
	if m != nil {
		return m.OpCode
	}
	return Packet_Invalid
}

func (m *Packet) GetData() *any.Any {
	if m != nil {
		return m.Data
	}
	return nil
}

type SeatConfiguration struct {
	Seats                []*Seat  `protobuf:"bytes,1,rep,name=seats,proto3" json:"seats,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SeatConfiguration) Reset()         { *m = SeatConfiguration{} }
func (m *SeatConfiguration) String() string { return proto.CompactTextString(m) }
func (*SeatConfiguration) ProtoMessage()    {}
func (*SeatConfiguration) Descriptor() ([]byte, []int) {
	return fileDescriptor_cd0c84d25a72589c, []int{1}
}

func (m *SeatConfiguration) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SeatConfiguration.Unmarshal(m, b)
}
func (m *SeatConfiguration) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SeatConfiguration.Marshal(b, m, deterministic)
}
func (m *SeatConfiguration) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SeatConfiguration.Merge(m, src)
}
func (m *SeatConfiguration) XXX_Size() int {
	return xxx_messageInfo_SeatConfiguration.Size(m)
}
func (m *SeatConfiguration) XXX_DiscardUnknown() {
	xxx_messageInfo_SeatConfiguration.DiscardUnknown(m)
}

var xxx_messageInfo_SeatConfiguration proto.InternalMessageInfo

func (m *SeatConfiguration) GetSeats() []*Seat {
	if m != nil {
		return m.Seats
	}
	return nil
}

type Context3D struct {
	Tick                 int32          `protobuf:"varint,1,opt,name=tick,proto3" json:"tick,omitempty"`
	Client               bool           `protobuf:"varint,2,opt,name=client,proto3" json:"client,omitempty"`
	Transforms           []*Transform   `protobuf:"bytes,3,rep,name=transforms,proto3" json:"transforms,omitempty"`
	RigidBodies          []*Rigidbody3D `protobuf:"bytes,4,rep,name=rigidBodies,proto3" json:"rigidBodies,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *Context3D) Reset()         { *m = Context3D{} }
func (m *Context3D) String() string { return proto.CompactTextString(m) }
func (*Context3D) ProtoMessage()    {}
func (*Context3D) Descriptor() ([]byte, []int) {
	return fileDescriptor_cd0c84d25a72589c, []int{2}
}

func (m *Context3D) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Context3D.Unmarshal(m, b)
}
func (m *Context3D) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Context3D.Marshal(b, m, deterministic)
}
func (m *Context3D) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Context3D.Merge(m, src)
}
func (m *Context3D) XXX_Size() int {
	return xxx_messageInfo_Context3D.Size(m)
}
func (m *Context3D) XXX_DiscardUnknown() {
	xxx_messageInfo_Context3D.DiscardUnknown(m)
}

var xxx_messageInfo_Context3D proto.InternalMessageInfo

func (m *Context3D) GetTick() int32 {
	if m != nil {
		return m.Tick
	}
	return 0
}

func (m *Context3D) GetClient() bool {
	if m != nil {
		return m.Client
	}
	return false
}

func (m *Context3D) GetTransforms() []*Transform {
	if m != nil {
		return m.Transforms
	}
	return nil
}

func (m *Context3D) GetRigidBodies() []*Rigidbody3D {
	if m != nil {
		return m.RigidBodies
	}
	return nil
}

type Context2D struct {
	Tick                 int32          `protobuf:"varint,1,opt,name=tick,proto3" json:"tick,omitempty"`
	Client               bool           `protobuf:"varint,2,opt,name=client,proto3" json:"client,omitempty"`
	Transforms           []*Transform   `protobuf:"bytes,3,rep,name=transforms,proto3" json:"transforms,omitempty"`
	RigidBodies          []*Rigidbody2D `protobuf:"bytes,4,rep,name=rigidBodies,proto3" json:"rigidBodies,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *Context2D) Reset()         { *m = Context2D{} }
func (m *Context2D) String() string { return proto.CompactTextString(m) }
func (*Context2D) ProtoMessage()    {}
func (*Context2D) Descriptor() ([]byte, []int) {
	return fileDescriptor_cd0c84d25a72589c, []int{3}
}

func (m *Context2D) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Context2D.Unmarshal(m, b)
}
func (m *Context2D) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Context2D.Marshal(b, m, deterministic)
}
func (m *Context2D) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Context2D.Merge(m, src)
}
func (m *Context2D) XXX_Size() int {
	return xxx_messageInfo_Context2D.Size(m)
}
func (m *Context2D) XXX_DiscardUnknown() {
	xxx_messageInfo_Context2D.DiscardUnknown(m)
}

var xxx_messageInfo_Context2D proto.InternalMessageInfo

func (m *Context2D) GetTick() int32 {
	if m != nil {
		return m.Tick
	}
	return 0
}

func (m *Context2D) GetClient() bool {
	if m != nil {
		return m.Client
	}
	return false
}

func (m *Context2D) GetTransforms() []*Transform {
	if m != nil {
		return m.Transforms
	}
	return nil
}

func (m *Context2D) GetRigidBodies() []*Rigidbody2D {
	if m != nil {
		return m.RigidBodies
	}
	return nil
}

type Transform struct {
	ID                   string      `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	Position             *Vector3    `protobuf:"bytes,2,opt,name=position,proto3" json:"position,omitempty"`
	Rotation             *Quaternion `protobuf:"bytes,3,opt,name=rotation,proto3" json:"rotation,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *Transform) Reset()         { *m = Transform{} }
func (m *Transform) String() string { return proto.CompactTextString(m) }
func (*Transform) ProtoMessage()    {}
func (*Transform) Descriptor() ([]byte, []int) {
	return fileDescriptor_cd0c84d25a72589c, []int{4}
}

func (m *Transform) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Transform.Unmarshal(m, b)
}
func (m *Transform) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Transform.Marshal(b, m, deterministic)
}
func (m *Transform) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Transform.Merge(m, src)
}
func (m *Transform) XXX_Size() int {
	return xxx_messageInfo_Transform.Size(m)
}
func (m *Transform) XXX_DiscardUnknown() {
	xxx_messageInfo_Transform.DiscardUnknown(m)
}

var xxx_messageInfo_Transform proto.InternalMessageInfo

func (m *Transform) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *Transform) GetPosition() *Vector3 {
	if m != nil {
		return m.Position
	}
	return nil
}

func (m *Transform) GetRotation() *Quaternion {
	if m != nil {
		return m.Rotation
	}
	return nil
}

type Rigidbody2D struct {
	ID                   string      `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	Position             *Vector2    `protobuf:"bytes,2,opt,name=position,proto3" json:"position,omitempty"`
	Rotation             *Quaternion `protobuf:"bytes,3,opt,name=rotation,proto3" json:"rotation,omitempty"`
	Velocity             *Vector2    `protobuf:"bytes,4,opt,name=velocity,proto3" json:"velocity,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *Rigidbody2D) Reset()         { *m = Rigidbody2D{} }
func (m *Rigidbody2D) String() string { return proto.CompactTextString(m) }
func (*Rigidbody2D) ProtoMessage()    {}
func (*Rigidbody2D) Descriptor() ([]byte, []int) {
	return fileDescriptor_cd0c84d25a72589c, []int{5}
}

func (m *Rigidbody2D) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Rigidbody2D.Unmarshal(m, b)
}
func (m *Rigidbody2D) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Rigidbody2D.Marshal(b, m, deterministic)
}
func (m *Rigidbody2D) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Rigidbody2D.Merge(m, src)
}
func (m *Rigidbody2D) XXX_Size() int {
	return xxx_messageInfo_Rigidbody2D.Size(m)
}
func (m *Rigidbody2D) XXX_DiscardUnknown() {
	xxx_messageInfo_Rigidbody2D.DiscardUnknown(m)
}

var xxx_messageInfo_Rigidbody2D proto.InternalMessageInfo

func (m *Rigidbody2D) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *Rigidbody2D) GetPosition() *Vector2 {
	if m != nil {
		return m.Position
	}
	return nil
}

func (m *Rigidbody2D) GetRotation() *Quaternion {
	if m != nil {
		return m.Rotation
	}
	return nil
}

func (m *Rigidbody2D) GetVelocity() *Vector2 {
	if m != nil {
		return m.Velocity
	}
	return nil
}

type Rigidbody3D struct {
	ID                   string      `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	Position             *Vector3    `protobuf:"bytes,2,opt,name=position,proto3" json:"position,omitempty"`
	Rotation             *Quaternion `protobuf:"bytes,3,opt,name=rotation,proto3" json:"rotation,omitempty"`
	Velocity             *Vector3    `protobuf:"bytes,4,opt,name=velocity,proto3" json:"velocity,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *Rigidbody3D) Reset()         { *m = Rigidbody3D{} }
func (m *Rigidbody3D) String() string { return proto.CompactTextString(m) }
func (*Rigidbody3D) ProtoMessage()    {}
func (*Rigidbody3D) Descriptor() ([]byte, []int) {
	return fileDescriptor_cd0c84d25a72589c, []int{6}
}

func (m *Rigidbody3D) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Rigidbody3D.Unmarshal(m, b)
}
func (m *Rigidbody3D) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Rigidbody3D.Marshal(b, m, deterministic)
}
func (m *Rigidbody3D) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Rigidbody3D.Merge(m, src)
}
func (m *Rigidbody3D) XXX_Size() int {
	return xxx_messageInfo_Rigidbody3D.Size(m)
}
func (m *Rigidbody3D) XXX_DiscardUnknown() {
	xxx_messageInfo_Rigidbody3D.DiscardUnknown(m)
}

var xxx_messageInfo_Rigidbody3D proto.InternalMessageInfo

func (m *Rigidbody3D) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *Rigidbody3D) GetPosition() *Vector3 {
	if m != nil {
		return m.Position
	}
	return nil
}

func (m *Rigidbody3D) GetRotation() *Quaternion {
	if m != nil {
		return m.Rotation
	}
	return nil
}

func (m *Rigidbody3D) GetVelocity() *Vector3 {
	if m != nil {
		return m.Velocity
	}
	return nil
}

type Vector2 struct {
	X                    float32  `protobuf:"fixed32,1,opt,name=x,proto3" json:"x,omitempty"`
	Y                    float32  `protobuf:"fixed32,2,opt,name=y,proto3" json:"y,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Vector2) Reset()         { *m = Vector2{} }
func (m *Vector2) String() string { return proto.CompactTextString(m) }
func (*Vector2) ProtoMessage()    {}
func (*Vector2) Descriptor() ([]byte, []int) {
	return fileDescriptor_cd0c84d25a72589c, []int{7}
}

func (m *Vector2) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Vector2.Unmarshal(m, b)
}
func (m *Vector2) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Vector2.Marshal(b, m, deterministic)
}
func (m *Vector2) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Vector2.Merge(m, src)
}
func (m *Vector2) XXX_Size() int {
	return xxx_messageInfo_Vector2.Size(m)
}
func (m *Vector2) XXX_DiscardUnknown() {
	xxx_messageInfo_Vector2.DiscardUnknown(m)
}

var xxx_messageInfo_Vector2 proto.InternalMessageInfo

func (m *Vector2) GetX() float32 {
	if m != nil {
		return m.X
	}
	return 0
}

func (m *Vector2) GetY() float32 {
	if m != nil {
		return m.Y
	}
	return 0
}

type Vector3 struct {
	X                    float32  `protobuf:"fixed32,1,opt,name=x,proto3" json:"x,omitempty"`
	Y                    float32  `protobuf:"fixed32,2,opt,name=y,proto3" json:"y,omitempty"`
	Z                    float32  `protobuf:"fixed32,3,opt,name=z,proto3" json:"z,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Vector3) Reset()         { *m = Vector3{} }
func (m *Vector3) String() string { return proto.CompactTextString(m) }
func (*Vector3) ProtoMessage()    {}
func (*Vector3) Descriptor() ([]byte, []int) {
	return fileDescriptor_cd0c84d25a72589c, []int{8}
}

func (m *Vector3) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Vector3.Unmarshal(m, b)
}
func (m *Vector3) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Vector3.Marshal(b, m, deterministic)
}
func (m *Vector3) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Vector3.Merge(m, src)
}
func (m *Vector3) XXX_Size() int {
	return xxx_messageInfo_Vector3.Size(m)
}
func (m *Vector3) XXX_DiscardUnknown() {
	xxx_messageInfo_Vector3.DiscardUnknown(m)
}

var xxx_messageInfo_Vector3 proto.InternalMessageInfo

func (m *Vector3) GetX() float32 {
	if m != nil {
		return m.X
	}
	return 0
}

func (m *Vector3) GetY() float32 {
	if m != nil {
		return m.Y
	}
	return 0
}

func (m *Vector3) GetZ() float32 {
	if m != nil {
		return m.Z
	}
	return 0
}

type Quaternion struct {
	W                    float32  `protobuf:"fixed32,1,opt,name=w,proto3" json:"w,omitempty"`
	X                    float32  `protobuf:"fixed32,2,opt,name=x,proto3" json:"x,omitempty"`
	Y                    float32  `protobuf:"fixed32,3,opt,name=y,proto3" json:"y,omitempty"`
	Z                    float32  `protobuf:"fixed32,4,opt,name=z,proto3" json:"z,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Quaternion) Reset()         { *m = Quaternion{} }
func (m *Quaternion) String() string { return proto.CompactTextString(m) }
func (*Quaternion) ProtoMessage()    {}
func (*Quaternion) Descriptor() ([]byte, []int) {
	return fileDescriptor_cd0c84d25a72589c, []int{9}
}

func (m *Quaternion) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Quaternion.Unmarshal(m, b)
}
func (m *Quaternion) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Quaternion.Marshal(b, m, deterministic)
}
func (m *Quaternion) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Quaternion.Merge(m, src)
}
func (m *Quaternion) XXX_Size() int {
	return xxx_messageInfo_Quaternion.Size(m)
}
func (m *Quaternion) XXX_DiscardUnknown() {
	xxx_messageInfo_Quaternion.DiscardUnknown(m)
}

var xxx_messageInfo_Quaternion proto.InternalMessageInfo

func (m *Quaternion) GetW() float32 {
	if m != nil {
		return m.W
	}
	return 0
}

func (m *Quaternion) GetX() float32 {
	if m != nil {
		return m.X
	}
	return 0
}

func (m *Quaternion) GetY() float32 {
	if m != nil {
		return m.Y
	}
	return 0
}

func (m *Quaternion) GetZ() float32 {
	if m != nil {
		return m.Z
	}
	return 0
}

type Seat struct {
	Owner                int32    `protobuf:"varint,1,opt,name=owner,proto3" json:"owner,omitempty"`
	GUID                 string   `protobuf:"bytes,2,opt,name=GUID,proto3" json:"GUID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Seat) Reset()         { *m = Seat{} }
func (m *Seat) String() string { return proto.CompactTextString(m) }
func (*Seat) ProtoMessage()    {}
func (*Seat) Descriptor() ([]byte, []int) {
	return fileDescriptor_cd0c84d25a72589c, []int{10}
}

func (m *Seat) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Seat.Unmarshal(m, b)
}
func (m *Seat) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Seat.Marshal(b, m, deterministic)
}
func (m *Seat) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Seat.Merge(m, src)
}
func (m *Seat) XXX_Size() int {
	return xxx_messageInfo_Seat.Size(m)
}
func (m *Seat) XXX_DiscardUnknown() {
	xxx_messageInfo_Seat.DiscardUnknown(m)
}

var xxx_messageInfo_Seat proto.InternalMessageInfo

func (m *Seat) GetOwner() int32 {
	if m != nil {
		return m.Owner
	}
	return 0
}

func (m *Seat) GetGUID() string {
	if m != nil {
		return m.GUID
	}
	return ""
}

func init() {
	proto.RegisterEnum("serializable.Packet_OpCode", Packet_OpCode_name, Packet_OpCode_value)
	proto.RegisterType((*Packet)(nil), "serializable.Packet")
	proto.RegisterType((*SeatConfiguration)(nil), "serializable.SeatConfiguration")
	proto.RegisterType((*Context3D)(nil), "serializable.Context3D")
	proto.RegisterType((*Context2D)(nil), "serializable.Context2D")
	proto.RegisterType((*Transform)(nil), "serializable.Transform")
	proto.RegisterType((*Rigidbody2D)(nil), "serializable.Rigidbody2D")
	proto.RegisterType((*Rigidbody3D)(nil), "serializable.Rigidbody3D")
	proto.RegisterType((*Vector2)(nil), "serializable.Vector2")
	proto.RegisterType((*Vector3)(nil), "serializable.Vector3")
	proto.RegisterType((*Quaternion)(nil), "serializable.Quaternion")
	proto.RegisterType((*Seat)(nil), "serializable.Seat")
}

func init() { proto.RegisterFile("serializable.proto", fileDescriptor_cd0c84d25a72589c) }

var fileDescriptor_cd0c84d25a72589c = []byte{
	// 552 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xcc, 0x94, 0xcf, 0x8e, 0xd3, 0x3e,
	0x10, 0xc7, 0x7f, 0x4e, 0xb3, 0xdd, 0x76, 0xba, 0xbf, 0x55, 0x18, 0x75, 0x97, 0x2c, 0x5c, 0xaa,
	0x48, 0x48, 0x3d, 0x65, 0x21, 0x41, 0xe2, 0x80, 0x38, 0x40, 0x23, 0xa1, 0x9e, 0x00, 0x2f, 0x70,
	0x77, 0x13, 0xb7, 0xb2, 0x36, 0x6b, 0x57, 0x89, 0xbb, 0x6d, 0x7a, 0xe6, 0x75, 0x78, 0x00, 0xde,
	0x80, 0x3f, 0x6f, 0xc1, 0x93, 0xa0, 0x38, 0x4d, 0x49, 0x10, 0x5a, 0x21, 0x38, 0xc0, 0x2d, 0xe3,
	0xf9, 0xcc, 0x77, 0xbe, 0x1e, 0xc7, 0x06, 0xcc, 0x79, 0x26, 0x58, 0x2a, 0xb6, 0x6c, 0x96, 0x72,
	0x7f, 0x99, 0x29, 0xad, 0xf0, 0xa8, 0xb9, 0x76, 0xe7, 0x6c, 0xa1, 0xd4, 0x22, 0xe5, 0xe7, 0x26,
	0x37, 0x5b, 0xcd, 0xcf, 0x99, 0x2c, 0x2a, 0xd0, 0xfb, 0x4a, 0xa0, 0xfb, 0x92, 0xc5, 0x97, 0x5c,
	0x63, 0x08, 0x5d, 0xb5, 0x9c, 0xa8, 0x84, 0xbb, 0x64, 0x44, 0xc6, 0xc7, 0xc1, 0x5d, 0xbf, 0x25,
	0x5c, 0x51, 0xfe, 0x0b, 0x83, 0xd0, 0x1d, 0x8a, 0x63, 0xb0, 0x13, 0xa6, 0x99, 0x6b, 0x8d, 0xc8,
	0x78, 0x10, 0x0c, 0xfd, 0xaa, 0x93, 0x5f, 0x77, 0xf2, 0x9f, 0xca, 0x82, 0x1a, 0xc2, 0x2b, 0xa0,
	0x5b, 0xd5, 0xe2, 0x00, 0x0e, 0xa7, 0xf2, 0x9a, 0xa5, 0x22, 0x71, 0xfe, 0xc3, 0x21, 0x38, 0x93,
	0x54, 0x70, 0xa9, 0x23, 0x91, 0xc7, 0x4a, 0x4a, 0x1e, 0x6b, 0x27, 0xc1, 0x53, 0xb8, 0x75, 0xc1,
	0x99, 0x9e, 0x28, 0x39, 0x17, 0x8b, 0x55, 0xc6, 0xb4, 0x50, 0xd2, 0xf9, 0x48, 0xb0, 0x0f, 0x76,
	0xb9, 0xee, 0x7c, 0x22, 0x88, 0xf0, 0x3f, 0x5d, 0xc9, 0x0b, 0x71, 0xb5, 0x4a, 0xab, 0xf4, 0x67,
	0x82, 0x27, 0xe0, 0x50, 0x9e, 0x2a, 0x96, 0x34, 0x96, 0xbf, 0x10, 0xef, 0xc9, 0x4f, 0xd4, 0x70,
	0x0c, 0x07, 0x39, 0x67, 0x3a, 0x77, 0xc9, 0xa8, 0x33, 0x1e, 0x04, 0xd8, 0xde, 0x6d, 0xc9, 0xd3,
	0x0a, 0xf0, 0xde, 0x13, 0xe8, 0x4f, 0x94, 0xd4, 0x7c, 0xa3, 0xc3, 0x08, 0x11, 0x6c, 0x2d, 0xe2,
	0x4b, 0x33, 0xa4, 0x03, 0x6a, 0xbe, 0xf1, 0x14, 0xba, 0xb1, 0xd9, 0x84, 0x99, 0x43, 0x8f, 0xee,
	0x22, 0x7c, 0x04, 0xa0, 0x33, 0x26, 0xf3, 0xb9, 0xca, 0xae, 0x72, 0xb7, 0x63, 0x1a, 0xdd, 0x6e,
	0x37, 0x7a, 0x5d, 0xe7, 0x69, 0x03, 0xc5, 0xc7, 0x30, 0xc8, 0xc4, 0x42, 0x24, 0xcf, 0x54, 0x22,
	0x78, 0xee, 0xda, 0xa6, 0xf2, 0xac, 0x5d, 0x49, 0x4b, 0x60, 0xa6, 0x92, 0x22, 0x8c, 0x68, 0x93,
	0x6e, 0xfa, 0x0d, 0xfe, 0x45, 0xbf, 0xc1, 0x0f, 0x7e, 0xdf, 0x11, 0xe8, 0xef, 0x65, 0xf1, 0x18,
	0xac, 0x69, 0x64, 0xdc, 0xf6, 0xa9, 0x35, 0x8d, 0xf0, 0x01, 0xf4, 0x96, 0x2a, 0x17, 0xe5, 0x99,
	0xed, 0xfe, 0xb2, 0x93, 0xb6, 0xee, 0x5b, 0x1e, 0x6b, 0x95, 0x85, 0x74, 0x8f, 0xe1, 0x43, 0xe8,
	0x65, 0x4a, 0x9b, 0x63, 0x76, 0x3b, 0xa6, 0xc4, 0x6d, 0x97, 0xbc, 0x5a, 0x31, 0xcd, 0x33, 0x29,
	0x94, 0xa4, 0x7b, 0xd2, 0xfb, 0x40, 0x60, 0xd0, 0xf0, 0xf8, 0xbb, 0x46, 0x82, 0x3f, 0x35, 0x52,
	0x36, 0xba, 0xe6, 0xa9, 0x8a, 0x85, 0x2e, 0x5c, 0xfb, 0xc6, 0x46, 0x35, 0xd6, 0xf6, 0x1e, 0x46,
	0x7f, 0x6d, 0x88, 0xbf, 0xea, 0x3d, 0x6c, 0x78, 0xbf, 0x07, 0x87, 0xbb, 0x0d, 0xe1, 0x11, 0x90,
	0x8d, 0x71, 0x6d, 0x51, 0xb2, 0x29, 0xa3, 0xc2, 0xb8, 0xb5, 0x28, 0x29, 0xbc, 0xb0, 0xc6, 0xc2,
	0x9b, 0xb0, 0x32, 0xda, 0x1a, 0xbf, 0x16, 0x25, 0x5b, 0x2f, 0x02, 0xf8, 0x6e, 0xb3, 0xcc, 0xad,
	0xeb, 0xba, 0x75, 0xa5, 0x62, 0xb5, 0x54, 0x3a, 0x2d, 0x15, 0xbb, 0x56, 0xb9, 0x5f, 0xbd, 0x3a,
	0x38, 0x84, 0x03, 0xb5, 0x96, 0x3c, 0xdb, 0xdd, 0xa5, 0x2a, 0x28, 0x2f, 0xd8, 0xf3, 0x37, 0xd3,
	0xc8, 0x48, 0xf5, 0xa9, 0xf9, 0x9e, 0x75, 0xcd, 0x03, 0x18, 0x7e, 0x0b, 0x00, 0x00, 0xff, 0xff,
	0xac, 0x98, 0x01, 0x18, 0x9c, 0x05, 0x00, 0x00,
}
