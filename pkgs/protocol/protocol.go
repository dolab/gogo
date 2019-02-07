package protocol

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/dolab/gogo/pkgs/protocol/json"
	"github.com/dolab/gogo/pkgs/protocol/jsonpb"
	"github.com/dolab/gogo/pkgs/protocol/protobuf"
)

type Protocol struct {
	codec Protocoler
}

func New(pt ProtocolType) (proto *Protocol, err error) {
	if !pt.IsValid() {
		err = errors.New("unexpected protocol type")
		return
	}

	switch pt {
	case ProtocolTypeJSON:
		proto = &Protocol{
			codec: json.New(),
		}

	case ProtocolTypeJSONPB:
		proto = &Protocol{
			codec: jsonpb.New(),
		}

	case ProtocolTypeProtobuf:
		proto = &Protocol{
			codec: protobuf.New(),
		}
	}

	return
}

func NewFromHTTPRequest(r *http.Request) (proto *Protocol, err error) {
	var (
		pt ProtocolType
		ok bool
	)
	if v := r.Context().Value(ctxProtocolKey); v != nil {
		pt, ok = v.(ProtocolType)
	}
	if !ok {
		pt, ok = NewProtocolTypeFromContentType(r.Header.Get("Content-Type"))
	}
	if !ok {
		err = fmt.Errorf("unexpected Content-Type: %q", r.Header.Get("Content-Type"))
		return
	}

	return New(pt)
}

func (p *Protocol) Unmarshal(r io.Reader, v interface{}) error {
	return p.codec.Unmarshal(r, v)
}

func (p *Protocol) Marshal(v interface{}) ([]byte, error) {
	return p.codec.Marshal(v)
}

func (p *Protocol) ContentType() string {
	return p.codec.ContentType()
}

func (p *Protocol) NewRequest(client HTTPClient) *Request {
	if httpclient, ok := client.(*http.Client); ok {
		client = hijackHTTPClientRedirects(httpclient)
	}

	return &Request{
		Protocoler: p.codec,
		HTTPClient: client,
	}
}
