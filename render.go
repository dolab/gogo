package gogo

import (
	"bytes"
	"crypto"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"hash"
	"io"
	"net/url"
)

// DefaultRender responses with default Content-Type header
type DefaultRender struct {
	response Responser
}

func NewDefaultRender(w Responser) Render {
	render := &DefaultRender{w}

	return render
}

func (render *DefaultRender) ContentType() string {
	return render.response.Header().Get("Content-Type")
}

func (render *DefaultRender) Render(v interface{}) error {
	if scr, ok := v.(StatusCoder); ok {
		render.response.WriteHeader(scr.StatusCode())
	}

	b := []byte("")
	switch v.(type) {
	case []byte:
		b, _ = v.([]byte)

	case string:
		s, _ := v.(string)

		b = []byte(s)

	case url.Values:
		p, _ := v.(url.Values)

		b = []byte(p.Encode())

	case io.Reader:
		r, _ := v.(io.Reader)

		// optimized for io.Reader
		_, err := io.Copy(render.response, r)
		return err
	}

	_, err := render.response.Write(b)
	return err
}

// HashRender responses with Etag header calculated from render data dynamically.
// NOTE: This always write response by copy if the render data is an io.Reader!!!
type HashRender struct {
	response Responser
	hash     hash.Hash
}

func NewHashRender(w Responser, hasher crypto.Hash) Render {
	if !hasher.Available() {
		panic(ErrHash.Error())
	}

	render := &HashRender{
		response: w,
		hash:     hasher.New(),
	}

	return render
}

func (render *HashRender) ContentType() string {
	return render.response.Header().Get("Content-Type")
}

func (render *HashRender) Render(v interface{}) error {
	if scr, ok := v.(StatusCoder); ok {
		render.response.WriteHeader(scr.StatusCode())
	}

	// reset hash
	render.hash.Reset()

	// using bytes.Buffer for efficient I/O
	var bbuf *bytes.Buffer

	switch v.(type) {
	case []byte:
		b, _ := v.([]byte)
		bbuf = bytes.NewBuffer(b)

	case string:
		s, _ := v.(string)
		bbuf = bytes.NewBufferString(s)

	case url.Values:
		p, _ := v.(url.Values)
		bbuf = bytes.NewBufferString(p.Encode())

	case io.Reader:
		r, _ := v.(io.Reader)

		bbuf = bytes.NewBuffer(nil)
		_, err := bbuf.ReadFrom(r)
		if err != nil {
			return err
		}
	}

	// add etag header
	render.hash.Write(bbuf.Bytes())
	render.response.Header().Add("Etag", hex.EncodeToString(render.hash.Sum(nil)))

	_, err := io.Copy(render.response, bbuf)
	return err
}

// TextRender responses with Content-Type: text/plain header
// It transform response data by stringify.
type TextRender struct {
	response Responser
}

func NewTextRender(w Responser) Render {
	render := &TextRender{w}

	return render
}

func (render *TextRender) ContentType() string {
	return "text/plain"
}

func (render *TextRender) Render(v interface{}) error {
	if scr, ok := v.(StatusCoder); ok {
		render.response.WriteHeader(scr.StatusCode())
	}

	render.response.Header().Set("Content-Type", render.ContentType())

	_, err := render.response.Write([]byte(fmt.Sprintf("%v", v)))
	return err
}

// JsonRender responses with Content-Type: application/json header
// It transform response data by json.Marshal.
type JsonRender struct {
	response Responser
}

func NewJsonRender(w Responser) Render {
	render := &JsonRender{w}

	return render
}

func (render *JsonRender) ContentType() string {
	return "application/json"
}

func (render *JsonRender) Render(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	if scr, ok := v.(StatusCoder); ok {
		render.response.WriteHeader(scr.StatusCode())
	}

	render.response.Header().Set("Content-Type", render.ContentType())

	_, err = render.response.Write(data)
	return err
}

// XmlRender responses with Content-Type: text/xml header
// It transform response data by xml.Marshal.
type XmlRender struct {
	response Responser
}

func NewXmlRender(w Responser) Render {
	render := &XmlRender{w}

	return render
}

func (render *XmlRender) ContentType() string {
	return "text/xml"
}

func (render *XmlRender) Render(v interface{}) error {
	data, err := xml.Marshal(v)
	if err != nil {
		return err
	}

	if scr, ok := v.(StatusCoder); ok {
		render.response.WriteHeader(scr.StatusCode())
	}

	render.response.Header().Set("Content-Type", render.ContentType())

	// write xml header
	_, err = render.response.Write([]byte(xml.Header))
	if err != nil {
		return err
	}

	_, err = render.response.Write(data)
	return err
}

// JsonpRender responses with Content-Type: application/javascript header
// It transform response data by json.Marshal.
type JsonpRender struct {
	response Responser
	callback string
}

func NewJsonpRender(w Responser, callback string) Render {
	render := &JsonpRender{
		response: w,
		callback: callback,
	}

	return render
}

func (render *JsonpRender) ContentType() string {
	return "application/javascript"
}

func (render *JsonpRender) Render(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	if scr, ok := v.(StatusCoder); ok {
		render.response.WriteHeader(scr.StatusCode())
	}

	render.response.Header().Set("Content-Type", render.ContentType())

	if _, err := render.response.Write([]byte(render.callback + "(")); err != nil {
		return err
	}
	if _, err := render.response.Write(data); err != nil {
		return err
	}
	if _, err := render.response.Write([]byte(");")); err != nil {
		return err
	}

	return nil
}
