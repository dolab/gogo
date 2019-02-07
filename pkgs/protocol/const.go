package protocol

import (
	"strings"
)

const (
	_ProtocolType ProtocolType = iota
	ProtocolTypeJSON
	ProtocolTypeJSONPB
	ProtocolTypeProtobuf
	ProtocolType_
)

var (
	protocolType2ContentType = map[ProtocolType]string{
		ProtocolTypeJSON:     "application/json",
		ProtocolTypeJSONPB:   "application/jsonpb",
		ProtocolTypeProtobuf: "application/protobuf",
	}
)

type ProtocolType int

func NewProtocolTypeFromContentType(contentType string) (ProtocolType, bool) {
	i := strings.Index(contentType, ";")
	if i < 0 {
		i = len(contentType)
	}

	contentType = strings.TrimSpace(strings.ToLower(contentType[:i]))
	for pt, ct := range protocolType2ContentType {
		if strings.Compare(contentType, ct) == 0 {
			return pt, true
		}
	}

	return 0, false
}

func (pt ProtocolType) IsValid() bool {
	return pt > _ProtocolType && pt < ProtocolType_
}

func (pt ProtocolType) ContentType() string {
	return protocolType2ContentType[pt]
}
