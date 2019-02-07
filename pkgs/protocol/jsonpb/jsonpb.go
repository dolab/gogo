package jsonpb

import (
	"bytes"
	"fmt"
	"io"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

type ProtocolJSONPB struct {
	unmarshaler jsonpb.Unmarshaler
	marshaler   jsonpb.Marshaler
}

func New() *ProtocolJSONPB {
	return &ProtocolJSONPB{
		unmarshaler: jsonpb.Unmarshaler{
			AllowUnknownFields: true,
		},
		marshaler: jsonpb.Marshaler{
			OrigName: true,
		},
	}
}

func (p *ProtocolJSONPB) ContentType() string {
	return "application/jsonpb"
}

func (p *ProtocolJSONPB) Unmarshal(r io.Reader, v interface{}) error {
	if r == nil {
		return fmt.Errorf("unmarshaler reader can not be nil")
	}

	pm, ok := v.(proto.Message)
	if !ok {
		return fmt.Errorf("unexpected type %T of value, expect proto.Message", v)
	}

	return p.unmarshaler.Unmarshal(r, pm)
}

func (p *ProtocolJSONPB) Marshal(v interface{}) ([]byte, error) {
	pm, ok := v.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("unexpected type %T of value, expect proto.Message", v)
	}

	buf := bytes.NewBuffer(nil)

	err := p.marshaler.Marshal(buf, pm)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
