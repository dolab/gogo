package render

import (
	"bytes"
	"crypto"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"hash"
	"io"
	"net/http"
	"net/url"
)

// HashRender responses with Etag header calculated from render data dynamically.
// NOTE: This always write response by copy if the render data is an io.Reader!!!
type HashRender struct {
	w http.ResponseWriter
	h hash.Hash
}

func NewHashRender(w http.ResponseWriter, hasher crypto.Hash) Render {
	if !hasher.Available() {
		panic(ErrHash.Error())
	}

	render := &HashRender{
		w: w,
		h: hasher.New(),
	}

	return render
}

func (render *HashRender) ContentType() string {
	contentType := render.w.Header().Get("Content-Type")
	if contentType == "" {
		contentType = ContentTypeDefault
	}

	return contentType
}

func (render *HashRender) Render(v interface{}) error {
	if v == nil {
		render.h.Reset()
		render.h.Write([]byte(""))
		render.w.Header().Set("Etag", hex.EncodeToString(render.h.Sum(nil)))

		return nil
	}

	var (
		// using bytes.Buffer for efficient I/O
		buf *bytes.Buffer
		err error
	)

	switch v.(type) {
	case []byte:
		buf = bytes.NewBuffer(v.([]byte))

	case string:
		buf = bytes.NewBufferString(v.(string))

	case url.Values:
		buf = bytes.NewBufferString(v.(url.Values).Encode())

	case io.Reader:
		buf = bytes.NewBuffer(nil)
		_, err = buf.ReadFrom(v.(io.Reader))

	default:
		switch render.ContentType() {
		case ContentTypeJSON, ContentTypeJSONP:
			buf = bytes.NewBuffer(nil)
			err = json.NewEncoder(buf).Encode(v)

		case ContentTypeXML:
			buf = bytes.NewBuffer(nil)
			err = xml.NewEncoder(buf).Encode(v)

		default:
			buf = bytes.NewBufferString(fmt.Sprintf("%v", v))
		}
	}

	if err != nil {
		return err
	}

	// hijack response header of etag
	render.h.Reset()
	_, err = render.h.Write(buf.Bytes())
	if err != nil {
		return err
	}

	render.w.Header().Set("Etag", hex.EncodeToString(render.h.Sum(nil)))

	_, err = io.Copy(render.w, buf)
	return err
}
