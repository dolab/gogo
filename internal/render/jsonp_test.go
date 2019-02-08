package render

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golib/assert"
)

func Test_JsonpRender(t *testing.T) {
	it := assert.New(t)
	recorder := httptest.NewRecorder()

	data := struct {
		Success bool   `json:"success"`
		Content string `json:"content"`
	}{true, "Hello, world!"}

	render := NewJsonpRender(recorder, "jsCallback")

	err := render.Render(data)
	if it.Nil(err) {
		it.Equal(http.StatusOK, recorder.Code)
		it.Equal("application/javascript", render.ContentType())
		it.Equal("jsCallback({\"success\":true,\"content\":\"Hello, world!\"}\n);", recorder.Body.String())
	}
}

func Benchmark_JsonpRender(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	recorder := httptest.NewRecorder()

	data := struct {
		Success bool   `json:"success"`
		Content string `json:"content"`
	}{true, "Hello, world!"}

	render := NewJsonpRender(recorder, "js_callback")

	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		render.Render(data)
	}
}
