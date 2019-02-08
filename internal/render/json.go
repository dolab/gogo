package render

import (
	"encoding/json"
	"net/http"
)

// JsonRender responses with Content-Type: application/json header
// It transform response data by json.Marshal.
type JsonRender struct {
	w http.ResponseWriter
}

func NewJsonRender(w http.ResponseWriter) Render {
	render := &JsonRender{w}

	return render
}

func (render *JsonRender) ContentType() string {
	return ContentTypeJSON
}

func (render *JsonRender) Render(v interface{}) error {
	if v == nil {
		return nil
	}

	return json.NewEncoder(render.w).Encode(v)
}
