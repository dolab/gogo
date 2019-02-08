package render

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// JsonpRender responses with Content-Type: application/javascript header
// It transform response data by json.Marshal.
type JsonpRender struct {
	w      http.ResponseWriter
	prefix io.Reader
	suffix io.Reader
}

func NewJsonpRender(w http.ResponseWriter, prefix string) Render {
	render := &JsonpRender{
		w:      w,
		prefix: bytes.NewBufferString(prefix + "("),
		suffix: bytes.NewBufferString(");"),
	}

	return render
}

func (render *JsonpRender) ContentType() string {
	return ContentTypeJSONP
}

func (render *JsonpRender) Render(v interface{}) error {
	_, err := io.Copy(render.w, render.prefix)
	if err != nil {
		return err
	}

	if v != nil {
		err = json.NewEncoder(render.w).Encode(v)
		if err != nil {
			return err
		}
	}

	_, err = io.Copy(render.w, render.suffix)
	return err
}
