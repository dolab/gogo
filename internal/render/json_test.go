package render

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golib/assert"
)

func Test_JsonRender(t *testing.T) {
	it := assert.New(t)
	recorder := httptest.NewRecorder()

	data := struct {
		Success bool   `json:"success"`
		Content string `json:"content"`
	}{true, "Hello, world!"}

	render := NewJsonRender(recorder)

	err := render.Render(data)
	if it.Nil(err) {
		it.Equal(http.StatusOK, recorder.Code)
		it.Equal("application/json", render.ContentType())
		it.Contains(recorder.Body.String(), `{"success":true,"content":"Hello, world!"}`)
	}
}

func Benchmark_JsonRender(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	recorder := httptest.NewRecorder()

	data := struct {
		Success bool   `json:"success"`
		Content string `json:"content"`
	}{true, "Hello, world!"}

	render := NewJsonRender(recorder)
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		render.Render(data)
	}
}
