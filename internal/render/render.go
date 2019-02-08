package render

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// DefaultRender responses with default Content-Type header
type DefaultRender struct {
	w http.ResponseWriter
}

func NewDefaultRender(w http.ResponseWriter) Render {
	render := &DefaultRender{w}

	return render
}

func (render *DefaultRender) ContentType() string {
	contentType := render.w.Header().Get("Content-Type")
	if contentType == "" {
		contentType = ContentTypeDefault
	}

	return contentType
}

func (render *DefaultRender) Render(v interface{}) error {
	if v == nil {
		return nil
	}

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
		case ContentTypeJSON, ContentTypeJSONP:
			err = json.NewEncoder(render.w).Encode(v)

		case ContentTypeXML:
			err = xml.NewEncoder(render.w).Encode(v)

		default:
			_, err = render.w.Write([]byte(fmt.Sprint(v)))
		}
	}

	return err
}
