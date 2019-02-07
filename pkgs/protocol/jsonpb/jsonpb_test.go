package jsonpb

import (
	"bytes"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/golib/assert"
)

var (
	buf = []byte{0x7b, 0x22, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x22, 0x3a, 0x22, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x22, 0x7d}
)

func Test_ProtocolJSONPB_Unmarshal(t *testing.T) {
	it := assert.New(t)
	pb := New()

	// it should work
	r := bytes.NewReader(buf)

	pm := new(PingInput)

	err := pb.Unmarshal(r, pm)
	if it.Nil(err) {
		it.Equal("Protobuf", pm.GetSubject())
	}

	// it should return error for invalid container
	r = bytes.NewReader(buf)

	var in PingInput

	err = pb.Unmarshal(r, in)
	if it.NotNil(err) {
		it.Empty(in.String())
		it.Contains(err.Error(), "expect proto.Message")
	}

	// it should return error for nil
	r = bytes.NewReader(buf)

	err = pb.Unmarshal(r, nil)
	if it.NotNil(err) {
		it.Contains(err.Error(), "expect proto.Message")
	}

	var ni PingInput

	err = pb.Unmarshal(nil, &ni)
	if it.Error(err) {
		it.Empty(ni.String())
		it.Contains(err.Error(), "unmarshaler reader can not be nil")
	}
}

func Test_ProtocolJSONPB_Marshal(t *testing.T) {
	it := assert.New(t)
	pb := New()

	// it should work for proto.Message
	pm := new(PingInput)

	b, err := pb.Marshal(pm)
	if it.Nil(err) {
		it.Equal("{}", string(b))
	}

	// it should work with data
	pm.Subject = "Protobuf"

	b, err = pb.Marshal(pm)
	if it.Nil(err) {
		it.Equal(buf, b)
	}

	// it should error with invalid container
	var in PingInput

	b, err = pb.Marshal(in)
	if it.Error(err) {
		it.Nil(b)
		it.Contains(err.Error(), "expect proto.Message")
	}

	// it should return error for nil
	b, err = pb.Marshal(nil)
	if it.NotNil(err) {
		it.Nil(b)
		it.Contains(err.Error(), "expect proto.Message")
	}
}

type PingInput struct {
	Subject              string   `protobuf:"bytes,1,opt,name=subject,proto3" json:"subject,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PingInput) Reset()         { *m = PingInput{} }
func (m *PingInput) String() string { return proto.CompactTextString(m) }
func (*PingInput) ProtoMessage()    {}
func (*PingInput) Descriptor() ([]byte, []int) {
	return fileDescriptor_ping_a7d452738bac4cf8, []int{0}
}
func (m *PingInput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PingInput.Unmarshal(m, b)
}
func (m *PingInput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PingInput.Marshal(b, m, deterministic)
}
func (dst *PingInput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PingInput.Merge(dst, src)
}
func (m *PingInput) XXX_Size() int {
	return xxx_messageInfo_PingInput.Size(m)
}
func (m *PingInput) XXX_DiscardUnknown() {
	xxx_messageInfo_PingInput.DiscardUnknown(m)
}

var xxx_messageInfo_PingInput proto.InternalMessageInfo

func (m *PingInput) GetSubject() string {
	if m != nil {
		return m.Subject
	}
	return ""
}

type PingOutput struct {
	Text                 string   `protobuf:"bytes,1,opt,name=text,proto3" json:"text,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PingOutput) Reset()         { *m = PingOutput{} }
func (m *PingOutput) String() string { return proto.CompactTextString(m) }
func (*PingOutput) ProtoMessage()    {}
func (*PingOutput) Descriptor() ([]byte, []int) {
	return fileDescriptor_ping_a7d452738bac4cf8, []int{1}
}
func (m *PingOutput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PingOutput.Unmarshal(m, b)
}
func (m *PingOutput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PingOutput.Marshal(b, m, deterministic)
}
func (dst *PingOutput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PingOutput.Merge(dst, src)
}
func (m *PingOutput) XXX_Size() int {
	return xxx_messageInfo_PingOutput.Size(m)
}
func (m *PingOutput) XXX_DiscardUnknown() {
	xxx_messageInfo_PingOutput.DiscardUnknown(m)
}

var xxx_messageInfo_PingOutput proto.InternalMessageInfo

func (m *PingOutput) GetText() string {
	if m != nil {
		return m.Text
	}
	return ""
}

type PongInput struct {
	Text                 string   `protobuf:"bytes,1,opt,name=text,proto3" json:"text,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PongInput) Reset()         { *m = PongInput{} }
func (m *PongInput) String() string { return proto.CompactTextString(m) }
func (*PongInput) ProtoMessage()    {}
func (*PongInput) Descriptor() ([]byte, []int) {
	return fileDescriptor_ping_a7d452738bac4cf8, []int{2}
}
func (m *PongInput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PongInput.Unmarshal(m, b)
}
func (m *PongInput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PongInput.Marshal(b, m, deterministic)
}
func (dst *PongInput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PongInput.Merge(dst, src)
}
func (m *PongInput) XXX_Size() int {
	return xxx_messageInfo_PongInput.Size(m)
}
func (m *PongInput) XXX_DiscardUnknown() {
	xxx_messageInfo_PongInput.DiscardUnknown(m)
}

var xxx_messageInfo_PongInput proto.InternalMessageInfo

func (m *PongInput) GetText() string {
	if m != nil {
		return m.Text
	}
	return ""
}

type PongOutput struct {
	Subject              string   `protobuf:"bytes,1,opt,name=subject,proto3" json:"subject,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PongOutput) Reset()         { *m = PongOutput{} }
func (m *PongOutput) String() string { return proto.CompactTextString(m) }
func (*PongOutput) ProtoMessage()    {}
func (*PongOutput) Descriptor() ([]byte, []int) {
	return fileDescriptor_ping_a7d452738bac4cf8, []int{3}
}
func (m *PongOutput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PongOutput.Unmarshal(m, b)
}
func (m *PongOutput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PongOutput.Marshal(b, m, deterministic)
}
func (dst *PongOutput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PongOutput.Merge(dst, src)
}
func (m *PongOutput) XXX_Size() int {
	return xxx_messageInfo_PongOutput.Size(m)
}
func (m *PongOutput) XXX_DiscardUnknown() {
	xxx_messageInfo_PongOutput.DiscardUnknown(m)
}

var xxx_messageInfo_PongOutput proto.InternalMessageInfo

func (m *PongOutput) GetSubject() string {
	if m != nil {
		return m.Subject
	}
	return ""
}

func init() {
	proto.RegisterType((*PingInput)(nil), "myapp.ping.PingInput")
	proto.RegisterType((*PingOutput)(nil), "myapp.ping.PingOutput")
	proto.RegisterType((*PongInput)(nil), "myapp.ping.PongInput")
	proto.RegisterType((*PongOutput)(nil), "myapp.ping.PongOutput")
}

func init() { proto.RegisterFile("ping.proto", fileDescriptor_ping_a7d452738bac4cf8) }

var fileDescriptor_ping_a7d452738bac4cf8 = []byte{
	// 166 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2a, 0xc8, 0xcc, 0x4b,
	0xd7, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0xca, 0xad, 0x4c, 0x2c, 0x28, 0xd0, 0x03, 0x89,
	0x28, 0xa9, 0x72, 0x71, 0x06, 0x64, 0xe6, 0xa5, 0x7b, 0xe6, 0x15, 0x94, 0x96, 0x08, 0x49, 0x70,
	0xb1, 0x17, 0x97, 0x26, 0x65, 0xa5, 0x26, 0x97, 0x48, 0x30, 0x2a, 0x30, 0x6a, 0x70, 0x06, 0xc1,
	0xb8, 0x4a, 0x0a, 0x5c, 0x5c, 0x20, 0x65, 0xfe, 0xa5, 0x25, 0x20, 0x75, 0x42, 0x5c, 0x2c, 0x25,
	0xa9, 0x15, 0x30, 0x45, 0x60, 0xb6, 0x92, 0x3c, 0x17, 0x67, 0x40, 0x3e, 0xcc, 0x20, 0x6c, 0x0a,
	0xd4, 0xb8, 0xb8, 0x40, 0x0a, 0xa0, 0x46, 0xe0, 0xb4, 0xca, 0xa8, 0x84, 0x8b, 0x05, 0x64, 0x95,
	0x90, 0x29, 0x94, 0x16, 0xd5, 0x43, 0x38, 0x57, 0x0f, 0xee, 0x56, 0x29, 0x31, 0x74, 0x61, 0xa8,
	0xc1, 0x20, 0x6d, 0xf9, 0x18, 0xda, 0xf2, 0xb1, 0x6b, 0x83, 0xbb, 0xc7, 0x89, 0x35, 0x8a, 0xb9,
	0x20, 0xa9, 0x38, 0x89, 0x0d, 0x1c, 0x42, 0xc6, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0xdc, 0x40,
	0x7d, 0xba, 0x2f, 0x01, 0x00, 0x00,
}
