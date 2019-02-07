package protocol

import (
	"io"
	"net/http"
)

type Protocoler interface {
	Marshaler
	Unmarshaler

	ContentType() string
}

type Unmarshaler interface {
	Unmarshal(r io.Reader, v interface{}) error
}

type Marshaler interface {
	Marshal(v interface{}) ([]byte, error)
}

// HTTPClient is the interface used by generated clients to send HTTP requests.
// It is fulfilled by *(net/http).Client, which is sufficient for most users.
// Users can provide their own implementation for special retry policies.
//
// HTTPClient implementations should not follow redirects. Redirects are
// automatically disabled if *(net/http).Client is passed to client
// constructors. See the withoutRedirects function in this file for more
// details.
type HTTPClient interface {
	Do(r *http.Request) (resp *http.Response, err error)
}
