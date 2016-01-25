package gogo

import (
	"bytes"
	"crypto"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
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

// HashRender responses with ETag header calculated from render data dynamically.
// Supported hash are [MD5|SHA1|SHA256], default to MD5. Its useful when render data is an io.Reader instance.
// NOTE: This always write response by copy!!!
type HashRender struct {
	response Responser
	hash     hash.Hash
}

func NewHashRender(w Responser, hashType crypto.Hash) Render {
	render := &HashRender{
		response: w,
	}

	switch hashType {
	// case crypto.MD5:
	// 	render.hash = md5.New()

	case crypto.SHA1:
		render.hash = sha1.New()

	case crypto.SHA256:
		render.hash = sha256.New()

	default:
		render.hash = md5.New()
	}

	return render
}

func (render *HashRender) ContentType() string {
	return render.response.Header().Get("Content-Type")
}

func (render *HashRender) Render(v interface{}) error {
	// reset hash
	render.hash.Reset()

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
		var err error

		r, _ := v.(io.Reader)

		// using bytes.Buffer for efficient I/O
		bbuf := bytes.NewBuffer(nil)
		bufSize := int64(2 << 14)

		for {
			rn, rerr := io.CopyN(bbuf, r, bufSize)

			if rn == 0 {
				err = rerr
				break
			}

			// NOTE: always proccess bytes readed!
			if rn > 0 {
				// write to response
				wn, werr := render.response.Write(bbuf.Bytes())
				if werr != nil {
					err = werr
					break
				}
				if rn != int64(wn) {
					err = io.ErrShortWrite
					break
				}

				// write to hash
				hn, herr := bbuf.WriteTo(render.hash)
				if herr != nil {
					err = herr
					break
				}
				if rn != hn {
					err = io.ErrShortWrite
					break
				}
			}

			if rerr != nil {
				// ignore io.EOF
				if rerr == io.EOF {
					rerr = nil
				}

				err = rerr
				break
			}
		}

		// add etag header
		render.response.Header().Add("ETag", hex.EncodeToString(render.hash.Sum(nil)))

		return err
	}

	// add etag header
	render.hash.Write(b)
	render.response.Header().Add("ETag", hex.EncodeToString(render.hash.Sum(nil)))

	_, err := render.response.Write(b)
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
