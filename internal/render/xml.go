package render

import (
	"bytes"
	"encoding/xml"
	"io"
	"net/http"
)

// XmlRender responses with Content-Type: text/xml header
// It transform response data by xml.Marshal.
type XmlRender struct {
	w      http.ResponseWriter
	header io.Reader
}

func NewXmlRender(w http.ResponseWriter) Render {
	render := &XmlRender{
		w:      w,
		header: bytes.NewBufferString(xml.Header),
	}

	return render
}

func (render *XmlRender) ContentType() string {
	return ContentTypeXML
}

func (render *XmlRender) Render(v interface{}) error {
	if v == nil {
		return nil
	}

	// hijack xml header
	_, err := io.Copy(render.w, render.header)
	if err != nil {
		return err
	}

	return xml.NewEncoder(render.w).Encode(v)
}
