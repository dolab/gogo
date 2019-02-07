package json

import (
	"bytes"
	"testing"

	"github.com/golib/assert"
)

var (
	buf = []byte{0x7b, 0x22, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x22, 0x3a, 0x22, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x22, 0x7d}
)

func Test_ProtocolJSON_Unmarshal(t *testing.T) {
	it := assert.New(t)
	pb := New()

	// it should work
	r := bytes.NewReader(buf)

	pm := new(PingInput)

	err := pb.Unmarshal(r, pm)
	if it.Nil(err) {
		it.Equal("Protobuf", pm.GetSubject())
	}

	// it should error for nil container
	r = bytes.NewReader(buf)

	var in PingInput

	err = pb.Unmarshal(r, in)
	if it.NotNil(err) {
		it.Empty(in.String())
		it.Contains(err.Error(), "expect pointer to value")
	}

	// it should return error for nil
	var ni PingInput

	err = pb.Unmarshal(nil, ni)
	if it.Error(err) {
		it.Empty(ni.String())
		it.Contains(err.Error(), "unmarshaler reader can not be nil")
	}

	err = pb.Unmarshal(bytes.NewBuffer(nil), nil)
	if it.Error(err) {
		it.Empty(ni.String())
		it.Contains(err.Error(), "unmarshaler value can not be nil")
	}
}

func Test_ProtocolJSON_Marshal(t *testing.T) {
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

	// it should work for zero
	var in PingInput

	b, err = pb.Marshal(in)
	if it.Nil(err) {
		it.Equal("{}", string(b))
	}

	// it should return error for nil
	b, err = pb.Marshal(nil)
	if it.NotNil(err) {
		it.Nil(b)
		it.Contains(err.Error(), "marshaler value can not be nil")
	}
}

type PingInput struct {
	Subject              string   `protobuf:"bytes,1,opt,name=subject,proto3" json:"subject,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PingInput) GetSubject() string {
	if m != nil {
		return m.Subject
	}
	return ""
}

func (m *PingInput) String() string {
	return m.GetSubject()
}

type PingOutput struct {
	Text                 string   `protobuf:"bytes,1,opt,name=text,proto3" json:"text,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PingOutput) GetText() string {
	if m != nil {
		return m.Text
	}
	return ""
}

func (m *PingOutput) String() string {
	return m.GetText()
}

type PongInput struct {
	Text                 string   `protobuf:"bytes,1,opt,name=text,proto3" json:"text,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PongInput) GetText() string {
	if m != nil {
		return m.Text
	}
	return ""
}

func (m *PongInput) String() string {
	return m.GetText()
}

type PongOutput struct {
	Subject              string   `protobuf:"bytes,1,opt,name=subject,proto3" json:"subject,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PongOutput) GetSubject() string {
	if m != nil {
		return m.Subject
	}
	return ""
}

func (m *PongOutput) String() string {
	return m.GetSubject()
}
