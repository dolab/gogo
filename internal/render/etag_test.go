package render

import (
	"crypto"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golib/assert"
)

func Test_HashRender(t *testing.T) {
	it := assert.New(t)
	recorder := httptest.NewRecorder()

	// render with normal string
	s := "Hello, world!"

	render := NewHashRender(recorder, crypto.MD5)
	it.Equal(ContentTypeDefault, render.ContentType())

	err := render.Render(s)
	if it.Nil(err) {
		it.Equal(http.StatusOK, recorder.Code)
		it.Equal("6cd3556deb0da54bca060b4c39479839", recorder.Header().Get("Etag"))
		it.Equal(s, recorder.Body.String())
	}
}

func Test_HashRenderWithReader(t *testing.T) {
	it := assert.New(t)
	recorder := httptest.NewRecorder()

	// render with io.Reader
	reader := strings.NewReader("Hello, world!")

	render := NewHashRender(recorder, crypto.MD5)
	it.Equal(ContentTypeDefault, render.ContentType())

	err := render.Render(reader)
	if it.Nil(err) {
		it.Equal(http.StatusOK, recorder.Code)
		it.Equal("6cd3556deb0da54bca060b4c39479839", recorder.Header().Get("Etag"))
		it.Equal("Hello, world!", recorder.Body.String())
	}
}

func Benchmark_HashRenderWithReader(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	recorder := httptest.NewRecorder()

	reader := strings.NewReader("Hello, world!")

	render := NewHashRender(recorder, crypto.MD5)
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		render.Render(reader)
	}
}

func Test_HashRenderWithJson(t *testing.T) {
	it := assert.New(t)
	recorder := httptest.NewRecorder()
	recorder.Header().Set("Content-Type", "application/json")

	// render with complex data type
	data := struct {
		Name string
		Age  int
	}{"gogo", 5}

	render := NewHashRender(recorder, crypto.MD5)

	err := render.Render(data)
	if it.Nil(err) {
		it.Equal("54843ae1dec66f4fefe6dfa7bcdf1567", recorder.Header().Get("Etag"))
		it.Equal(`{"Name":"gogo","Age":5}`, strings.TrimSpace(recorder.Body.String()))
	}
}

func Benchmark_HashRenderWithJson(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	recorder := httptest.NewRecorder()
	recorder.Header().Set("Content-Type", "application/json")

	data := struct {
		Name string
		Age  int
	}{"gogo", 5}

	render := NewHashRender(recorder, crypto.MD5)
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		render.Render(data)
	}
}

func Test_HashRenderWithXml(t *testing.T) {
	it := assert.New(t)

	recorder := httptest.NewRecorder()
	recorder.Header().Set("Content-Type", "text/xml")

	// render with complex data type
	data := struct {
		XMLName xml.Name `xml:"recorder"`
		Success bool     `xml:"Result>Success"`
		Content string   `xml:"Result>Content"`
	}{
		Success: true,
		Content: "Hello, world!",
	}

	render := NewHashRender(recorder, crypto.MD5)

	err := render.Render(data)
	if it.Nil(err) {
		it.Equal("65693ee59f678f04bc8bedf16f980f5a", recorder.Header().Get("Etag"))
		it.Equal("<recorder><Result><Success>true</Success><Content>Hello, world!</Content></Result></recorder>", recorder.Body.String())
	}
}

func Benchmark_HashRenderWithXml(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	recorder := httptest.NewRecorder()
	recorder.Header().Set("Content-Type", "text/xml")

	data := struct {
		XMLName xml.Name `xml:"recorder"`
		Success bool     `xml:"Result>Success"`
		Content string   `xml:"Result>Content"`
	}{
		Success: true,
		Content: "Hello, world!",
	}

	render := NewHashRender(recorder, crypto.MD5)
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		render.Render(data)
	}
}

func Test_HashRenderWithStringify(t *testing.T) {
	it := assert.New(t)
	recorder := httptest.NewRecorder()

	// render with complex data type
	data := struct {
		Name string
		Age  int
	}{"gogo", 5}

	render := NewHashRender(recorder, crypto.MD5)

	err := render.Render(data)
	if it.Nil(err) {
		it.Equal("1b9f54d6753f2e8e4d4a819a44d90ce1", recorder.Header().Get("Etag"))
		it.Contains(recorder.Body.String(), `{gogo 5}`)
	}
}

func Benchmark_HashRenderWithStringify(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	recorder := httptest.NewRecorder()

	data := struct {
		Name string
		Age  int
	}{"gogo", 5}

	render := NewHashRender(recorder, crypto.MD5)
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		render.Render(data)
	}
}
