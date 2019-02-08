package render

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golib/assert"
)

func Test_TextRender(t *testing.T) {
	it := assert.New(t)
	recorder := httptest.NewRecorder()

	// render with normal string
	s := "Hello, world!"

	render := NewTextRender(recorder)
	it.Equal(ContentTypeDefault, render.ContentType())

	err := render.Render(s)
	if it.Nil(err) {
		it.Equal(http.StatusOK, recorder.Code)
		it.Equal(s, recorder.Body.String())
	}
}

func Test_TextRenderWithReader(t *testing.T) {
	it := assert.New(t)
	recorder := httptest.NewRecorder()
	recorder.Header().Add("Content-Type", "text/plain")

	// render with io.Reader
	s := "Hello, world!"
	reader := strings.NewReader(s)

	render := NewTextRender(recorder)
	it.Equal(ContentTypeDefault, render.ContentType())

	err := render.Render(reader)
	if it.Nil(err) {
		it.Equal(s, recorder.Body.String())
	}
}

func Benchmark_TextRenderWithReader(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	recorder := httptest.NewRecorder()

	reader := strings.NewReader("Hello, world!")

	render := NewTextRender(recorder)
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		render.Render(reader)
	}
}

func Test_TextRenderWithStringify(t *testing.T) {
	it := assert.New(t)
	recorder := httptest.NewRecorder()
	recorder.Header().Add("Content-Type", "text/plain")

	// render with complex data type
	data := struct {
		Success bool
		Content string
	}{true, "Hello, world!"}

	render := NewTextRender(recorder)

	err := render.Render(data)
	if it.Nil(err) {
		it.Equal(`{true Hello, world!}`, recorder.Body.String())
	}
}

func Benchmark_TextRenderWithStringify(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	recorder := httptest.NewRecorder()

	data := struct {
		Success bool
		Content string
	}{true, "Hello, world!"}

	render := NewTextRender(recorder)
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		render.Render(data)
	}
}
