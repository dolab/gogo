package gogo

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
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
		buf := make([]byte, 0xffff) // 64k
		for {
			n, err := r.Read(buf)
			if err != nil {
				if err != io.EOF {
					return err
				}

				return nil
			}

			_, err = render.response.Write(buf[:n])
			if err != nil {
				return err
			}
		}

		return nil
	}

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

	data := fmt.Sprintf("%v", v)

	// could we auto transfer response status here?
	// if data == "" {
	// 	render.response.WriteHeader(http.StatusNoContent)

	// 	render.response.Write([]byte(""))
	// 	return nil
	// }

	_, err := render.response.Write([]byte(data))
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
