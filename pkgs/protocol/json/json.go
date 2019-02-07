package json

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

type ProtocolJSON struct {
	// ignore
}

func New() *ProtocolJSON {
	return &ProtocolJSON{}
}

func (p *ProtocolJSON) ContentType() string {
	return "application/json"
}

func (p *ProtocolJSON) Unmarshal(r io.Reader, v interface{}) error {
	if r == nil {
		return fmt.Errorf("unmarshaler reader can not be nil")
	}
	if v == nil {
		return fmt.Errorf("unmarshaler value can not be nil")
	}

	vtype := reflect.TypeOf(v)
	if vtype.Kind() != reflect.Ptr {
		return fmt.Errorf("unexpect type %T of value, expect pointer to value", v)
	}

	err := json.NewDecoder(r).Decode(&v)
	if err == io.EOF {
		err = nil
	}

	return err
}

func (p *ProtocolJSON) Marshal(v interface{}) ([]byte, error) {
	if v == nil {
		return nil, fmt.Errorf("marshaler value can not be nil")
	}

	return json.Marshal(v)
}
