package render

import (
	"crypto"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golib/assert"
)

func Test_DefaultRender(t *testing.T) {
	it := assert.New(t)
	recorder := httptest.NewRecorder()

	// render with normal string
	s := "Hello, world!"

	render := NewDefaultRender(recorder)
	it.Equal(ContentTypeDefault, render.ContentType())

	err := render.Render(s)
	if it.Nil(err) {
		it.Equal(http.StatusOK, recorder.Code)
		it.Equal(ContentTypeDefault, recorder.Header().Get("Content-Type"))
		it.Equal(s, recorder.Body.String())
	}
}

func Benchmark_DefaultRender(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	recorder := httptest.NewRecorder()

	s := "Hello, world!"

	render := NewDefaultRender(recorder)
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		render.Render(s)
	}
}

func Test_DefaultRenderWithReader(t *testing.T) {
	it := assert.New(t)
	recorder := httptest.NewRecorder()

	s := "Hello, world!"

	recorder.Header().Add("Content-Length", fmt.Sprintf("%d", len(s)))
	recorder.Header().Add("Content-Type", "text/plain")

	// render with normal string
	reader := strings.NewReader(s)

	render := NewDefaultRender(recorder)
	it.Equal("text/plain", render.ContentType())

	err := render.Render(reader)
	if it.Nil(err) {
		it.Equal(http.StatusOK, recorder.Code)
		it.Equal(s, recorder.Body.String())
	}
}

func Benchmark_DefaultRenderWithReader(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	recorder := httptest.NewRecorder()

	reader := strings.NewReader("Hello, world!")

	render := NewDefaultRender(recorder)
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		render.Render(reader)
	}
}

func Test_DefaultRenderWithJson(t *testing.T) {
	it := assert.New(t)
	recorder := httptest.NewRecorder()
	recorder.Header().Set("Content-Type", "application/json")

	// render with complex data type
	data := struct {
		Name string
		Age  int
	}{"gogo", 5}

	render := NewDefaultRender(recorder)

	err := render.Render(data)
	if it.Nil(err) {
		it.Equal(`{"Name":"gogo","Age":5}`, strings.TrimSpace(recorder.Body.String()))
	}
}

func Benchmark_DefaultRenderWithJson(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	recorder := httptest.NewRecorder()
	recorder.Header().Set("Content-Type", "application/json")

	data := struct {
		Name string
		Age  int
	}{"gogo", 5}

	render := NewDefaultRender(recorder)
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		render.Render(data)
	}
}

func Test_DefaultRenderWithXml(t *testing.T) {
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

	render := NewDefaultRender(recorder)

	err := render.Render(data)
	if it.Nil(err) {
		it.Equal("<recorder><Result><Success>true</Success><Content>Hello, world!</Content></Result></recorder>", recorder.Body.String())
	}
}

func Benchmark_DefaultRenderWithXml(b *testing.B) {
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

	render := NewDefaultRender(recorder)
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		render.Render(data)
	}
}

func Test_DefaultRenderWithStringify(t *testing.T) {
	it := assert.New(t)
	recorder := httptest.NewRecorder()

	// render with complex data type
	data := struct {
		Name string
		Age  int
	}{"gogo", 5}

	render := NewDefaultRender(recorder)

	err := render.Render(data)
	if it.Nil(err) {
		it.Equal(`{gogo 5}`, recorder.Body.String())
	}
}

func Benchmark_DefaultRenderWithStringify(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	recorder := httptest.NewRecorder()

	data := struct {
		Name string
		Age  int
	}{"gogo", 5}

	render := NewDefaultRender(recorder)
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		render.Render(data)
	}
}

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

func Test_XmlRender(t *testing.T) {
	it := assert.New(t)
	recorder := httptest.NewRecorder()

	data := struct {
		XMLName xml.Name `xml:"recorder"`
		Success bool     `xml:"Result>Success"`
		Content string   `xml:"Result>Content"`
	}{
		Success: true,
		Content: "Hello, world!",
	}

	render := NewXmlRender(recorder)

	err := render.Render(data)
	if it.Nil(err) {
		it.Equal(http.StatusOK, recorder.Code)
		it.Equal("text/xml", render.ContentType())
		it.Equal("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<recorder><Result><Success>true</Success><Content>Hello, world!</Content></Result></recorder>", recorder.Body.String())
	}
}

func Benchmark_XmlRender(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	recorder := httptest.NewRecorder()

	data := struct {
		XMLName xml.Name `xml:"recorder"`
		Success bool     `xml:"Result>Success"`
		Content string   `xml:"Result>Content"`
	}{
		Success: true,
		Content: "Hello, world!",
	}

	render := NewXmlRender(recorder)
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		render.Render(data)
	}
}

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
