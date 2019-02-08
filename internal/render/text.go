package render

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// TextRender responses with Content-Type: text/plain header
// It transform response data by stringify.
type TextRender struct {
	w http.ResponseWriter
}

func NewTextRender(w http.ResponseWriter) Render {
	render := &TextRender{w}

	return render
}

func (render *TextRender) ContentType() string {
	return ContentTypeDefault
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
