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
	w Responser
}

func NewDefaultRender(w Responser) Render {
	render := &DefaultRender{w}

	return render
}

func (render *DefaultRender) ContentType() string {
	contentType := render.w.Header().Get("Content-Type")
	if contentType == "" {
		contentType = RenderDefaultContentType
	}

	return contentType
}

func (render *DefaultRender) Render(v interface{}) error {
	var (
		err error
	)

	switch v.(type) {
	case []byte:
		_, err = render.w.Write(v.([]byte))

	case string:
		_, err = render.w.Write([]byte(v.(string)))

	case url.Values:
		_, err = render.w.Write([]byte(v.(url.Values).Encode()))

	case io.Reader:
		// optimized for io.Reader
		_, err = io.Copy(render.w, v.(io.Reader))

	default:
		switch render.ContentType() {
		case RenderJsonContentType, RenderJsonPContentType:
			err = json.NewEncoder(render.w).Encode(v)

		case RednerXmlContentType:
			err = xml.NewEncoder(render.w).Encode(v)

		default:
			_, err = render.w.Write([]byte(fmt.Sprint(v)))
		}
	}

	return err
}

// HashRender responses with Etag header calculated from render data dynamically.
// NOTE: This always write response by copy if the render data is an io.Reader!!!
type HashRender struct {
	w    Responser
	hash hash.Hash
}

func NewHashRender(w Responser, hasher crypto.Hash) Render {
	if !hasher.Available() {
		panic(ErrHash.Error())
	}

	render := &HashRender{
		w:    w,
		hash: hasher.New(),
	}

	return render
}

func (render *HashRender) ContentType() string {
	contentType := render.w.Header().Get("Content-Type")
	if contentType == "" {
		contentType = RenderDefaultContentType
	}

	return contentType
}

func (render *HashRender) Render(v interface{}) error {
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
		case RenderJsonContentType, RenderJsonPContentType:
			buf = bytes.NewBuffer(nil)
			err = json.NewEncoder(buf).Encode(v)

		case RednerXmlContentType:
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
	render.hash.Reset()
	_, err = render.hash.Write(buf.Bytes())
	if err != nil {
		return err
	}

	render.w.Header().Set("Etag", hex.EncodeToString(render.hash.Sum(nil)))

	_, err = io.Copy(render.w, buf)
	return err
}

// TextRender responses with Content-Type: text/plain header
// It transform response data by stringify.
type TextRender struct {
	w Responser
}

func NewTextRender(w Responser) Render {
	render := &TextRender{w}

	return render
}

func (render *TextRender) ContentType() string {
	return RenderDefaultContentType
}

func (render *TextRender) Render(v interface{}) error {
	var (
		err error
	)

	switch v.(type) {
	case []byte:
		_, err = render.w.Write(v.([]byte))

	case string:
		_, err = render.w.Write([]byte(v.(string)))

	case url.Values:
		_, err = render.w.Write([]byte(v.(url.Values).Encode()))

	case io.Reader:
		// optimized for io.Reader
		_, err = io.Copy(render.w, v.(io.Reader))

	default:
		_, err = render.w.Write([]byte(fmt.Sprint(v)))
	}

	return err
}

// JsonRender responses with Content-Type: application/json header
// It transform response data by json.Marshal.
type JsonRender struct {
	w Responser
}

func NewJsonRender(w Responser) Render {
	render := &JsonRender{w}

	return render
}

func (render *JsonRender) ContentType() string {
	return RenderJsonContentType
}

func (render *JsonRender) Render(v interface{}) error {
	return json.NewEncoder(render.w).Encode(v)
}

// XmlRender responses with Content-Type: text/xml header
// It transform response data by xml.Marshal.
type XmlRender struct {
	w      Responser
	header io.Reader
}

func NewXmlRender(w Responser) Render {
	render := &XmlRender{
		w:      w,
		header: bytes.NewBufferString(xml.Header),
	}

	return render
}

func (render *XmlRender) ContentType() string {
	return RednerXmlContentType
}

func (render *XmlRender) Render(v interface{}) error {
	// hijack xml header
	_, err := io.Copy(render.w, render.header)
	if err != nil {
		return err
	}

	return xml.NewEncoder(render.w).Encode(v)
}

// JsonpRender responses with Content-Type: application/javascript header
// It transform response data by json.Marshal.
type JsonpRender struct {
	w        Responser
	callback io.Reader
	tailer   io.Reader
}

func NewJsonpRender(w Responser, callback string) Render {
	render := &JsonpRender{
		w:        w,
		callback: bytes.NewBufferString(callback + "("),
		tailer:   bytes.NewBufferString(");"),
	}

	return render
}

func (render *JsonpRender) ContentType() string {
	return RenderJsonPContentType
}

func (render *JsonpRender) Render(v interface{}) error {
	_, err := io.Copy(render.w, render.callback)
	if err != nil {
		return err
	}

	err = json.NewEncoder(render.w).Encode(v)
	if err != nil {
		return err
	}

	_, err = io.Copy(render.w, render.tailer)
	return err
}
