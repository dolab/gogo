package protobuf

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/gogo/protobuf/proto"
)

type ProtocolProtobuf struct {
	// ignore
}

func New() *ProtocolProtobuf {
	return &ProtocolProtobuf{}
}

func (p *ProtocolProtobuf) ContentType() string {
	return "application/protobuf"
}

func (p *ProtocolProtobuf) Unmarshal(r io.Reader, v interface{}) error {
	if r == nil {
		return fmt.Errorf("unmarshaler reader can not be nil")
	}

	pm, ok := v.(proto.Message)
	if !ok {
		return fmt.Errorf("unexpected type %T of value, expect proto.Message", v)
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	return proto.Unmarshal(b, pm)
}

func (p *ProtocolProtobuf) Marshal(v interface{}) ([]byte, error) {
	pm, ok := v.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("unexpected type %T of value, expect proto.Message", v)
	}

	return proto.Marshal(pm)
}
