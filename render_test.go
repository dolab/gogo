package gogo

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
	assertion := assert.New(t)
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)

	// render with normal string
	s := "Hello, world!"

	render := NewDefaultRender(response)
	assertion.Equal(RenderDefaultContentType, render.ContentType())

	err := render.Render(s)
	assertion.Nil(err)
	assertion.Equal(http.StatusOK, recorder.Code)
	assertion.Empty(recorder.Header())
	assertion.Equal(s, recorder.Body.String())
}

func Test_DefaultRenderWithReader(t *testing.T) {
	assertion := assert.New(t)
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)

	s := "Hello, world!"

	response.Header().Add("Content-Length", fmt.Sprintf("%d", len(s)))
	response.Header().Add("Content-Type", "text/plain")

	// render with normal string
	reader := strings.NewReader(s)

	render := NewDefaultRender(response)
	assertion.Equal("text/plain", render.ContentType())

	err := render.Render(reader)
	assertion.Nil(err)
	assertion.Equal(http.StatusOK, recorder.Code)
	assertion.Equal(s, recorder.Body.String())
}

func Benchmark_DefaultRenderWithReader(b *testing.B) {
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)

	reader := strings.NewReader("Hello, world!")

	render := NewDefaultRender(response)
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		render.Render(reader)
	}
}

func Test_DefaultRenderWithJson(t *testing.T) {
	assertion := assert.New(t)
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)
	response.Header().Set("Content-Type", "application/json")

	// render with complex data type
	data := struct {
		Name string
		Age  int
	}{"gogo", 5}

	render := NewDefaultRender(response)

	err := render.Render(data)
	assertion.Nil(err)
	assertion.Contains(recorder.Body.String(), `{"Name":"gogo","Age":5}`)
}

func Benchmark_DefaultRenderWithJson(b *testing.B) {
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)
	response.Header().Set("Content-Type", "application/json")

	data := struct {
		Name string
		Age  int
	}{"gogo", 5}

	render := NewDefaultRender(response)
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		render.Render(data)
	}
}

func Test_DefaultRenderWithXml(t *testing.T) {
	assertion := assert.New(t)
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)
	response.Header().Set("Content-Type", "text/xml")

	// render with complex data type
	data := struct {
		XMLName xml.Name `xml:"Response"`
		Success bool     `xml:"Result>Success"`
		Content string   `xml:"Result>Content"`
	}{
		Success: true,
		Content: "Hello, world!",
	}

	render := NewDefaultRender(response)

	err := render.Render(data)
	assertion.Nil(err)
	assertion.Equal("<Response><Result><Success>true</Success><Content>Hello, world!</Content></Result></Response>", recorder.Body.String())
}

func Benchmark_DefaultRenderWithXml(b *testing.B) {
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)
	response.Header().Set("Content-Type", "text/xml")

	data := struct {
		XMLName xml.Name `xml:"Response"`
		Success bool     `xml:"Result>Success"`
		Content string   `xml:"Result>Content"`
	}{
		Success: true,
		Content: "Hello, world!",
	}

	render := NewDefaultRender(response)
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		render.Render(data)
	}
}

func Test_DefaultRenderWithStringify(t *testing.T) {
	assertion := assert.New(t)
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)

	// render with complex data type
	data := struct {
		Name string
		Age  int
	}{"gogo", 5}

	render := NewDefaultRender(response)

	err := render.Render(data)
	assertion.Nil(err)
	assertion.Equal(`{gogo 5}`, recorder.Body.String())
}

func Benchmark_DefaultRenderWithStringify(b *testing.B) {
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)

	data := struct {
		Name string
		Age  int
	}{"gogo", 5}

	render := NewDefaultRender(response)
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		render.Render(data)
	}
}

func Test_HashRender(t *testing.T) {
	assertion := assert.New(t)
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)

	// render with normal string
	s := "Hello, world!"

	render := NewHashRender(response, crypto.MD5)
	assertion.Equal(RenderDefaultContentType, render.ContentType())

	err := render.Render(s)
	assertion.Nil(err)
	assertion.Equal(http.StatusOK, recorder.Code)
	assertion.Equal("6cd3556deb0da54bca060b4c39479839", recorder.Header().Get("Etag"))
	assertion.Equal(s, recorder.Body.String())
}

func Test_HashRenderWithReader(t *testing.T) {
	assertion := assert.New(t)
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)

	// render with io.Reader
	reader := strings.NewReader("Hello, world!")

	render := NewHashRender(response, crypto.MD5)
	assertion.Equal(RenderDefaultContentType, render.ContentType())

	err := render.Render(reader)
	assertion.Nil(err)
	assertion.Equal(http.StatusOK, recorder.Code)
	assertion.Equal("6cd3556deb0da54bca060b4c39479839", recorder.Header().Get("Etag"))
	assertion.Equal("Hello, world!", recorder.Body.String())
}

func Benchmark_HashRenderWithReader(b *testing.B) {
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)

	reader := strings.NewReader("Hello, world!")

	render := NewHashRender(response, crypto.MD5)
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		render.Render(reader)
	}
}

func Test_HashRenderWithJson(t *testing.T) {
	assertion := assert.New(t)
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)
	response.Header().Set("Content-Type", "application/json")

	// render with complex data type
	data := struct {
		Name string
		Age  int
	}{"gogo", 5}

	render := NewHashRender(response, crypto.MD5)

	err := render.Render(data)
	assertion.Nil(err)
	assertion.Equal("54843ae1dec66f4fefe6dfa7bcdf1567", recorder.Header().Get("Etag"))
	assertion.Contains(recorder.Body.String(), `{"Name":"gogo","Age":5}`)
}

func Benchmark_HashRenderWithJson(b *testing.B) {
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)
	response.Header().Set("Content-Type", "application/json")

	data := struct {
		Name string
		Age  int
	}{"gogo", 5}

	render := NewHashRender(response, crypto.MD5)
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		render.Render(data)
	}
}

func Test_HashRenderWithXml(t *testing.T) {
	assertion := assert.New(t)
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)
	response.Header().Set("Content-Type", "text/xml")

	// render with complex data type
	data := struct {
		XMLName xml.Name `xml:"Response"`
		Success bool     `xml:"Result>Success"`
		Content string   `xml:"Result>Content"`
	}{
		Success: true,
		Content: "Hello, world!",
	}

	render := NewHashRender(response, crypto.MD5)

	err := render.Render(data)
	assertion.Nil(err)
	assertion.Equal("882dcefe3dd48e4dc99354002c4ce6e8", recorder.Header().Get("Etag"))
	assertion.Equal("<Response><Result><Success>true</Success><Content>Hello, world!</Content></Result></Response>", recorder.Body.String())
}

func Benchmark_HashRenderWithXml(b *testing.B) {
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)
	response.Header().Set("Content-Type", "text/xml")

	data := struct {
		XMLName xml.Name `xml:"Response"`
		Success bool     `xml:"Result>Success"`
		Content string   `xml:"Result>Content"`
	}{
		Success: true,
		Content: "Hello, world!",
	}

	render := NewHashRender(response, crypto.MD5)
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		render.Render(data)
	}
}

func Test_HashRenderWithStringify(t *testing.T) {
	assertion := assert.New(t)
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)

	// render with complex data type
	data := struct {
		Name string
		Age  int
	}{"gogo", 5}

	render := NewHashRender(response, crypto.MD5)

	err := render.Render(data)
	assertion.Nil(err)
	assertion.Equal("1b9f54d6753f2e8e4d4a819a44d90ce1", recorder.Header().Get("Etag"))
	assertion.Contains(recorder.Body.String(), `{gogo 5}`)
}

func Benchmark_HashRenderWithStringify(b *testing.B) {
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)

	data := struct {
		Name string
		Age  int
	}{"gogo", 5}

	render := NewHashRender(response, crypto.MD5)
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		render.Render(data)
	}
}

func Test_TextRender(t *testing.T) {
	assertion := assert.New(t)
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)

	// render with normal string
	s := "Hello, world!"

	render := NewTextRender(response)
	assertion.Equal(RenderDefaultContentType, render.ContentType())

	err := render.Render(s)
	assertion.Nil(err)
	assertion.Equal(http.StatusOK, recorder.Code)
	assertion.Equal(s, recorder.Body.String())
}

func Test_TextRenderWithReader(t *testing.T) {
	assertion := assert.New(t)
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)
	response.Header().Add("Content-Type", "text/plain")

	// render with io.Reader
	s := "Hello, world!"
	reader := strings.NewReader(s)

	render := NewTextRender(response)
	assertion.Equal(RenderDefaultContentType, render.ContentType())

	err := render.Render(reader)
	assertion.Nil(err)
	assertion.Equal(s, recorder.Body.String())
}

func Benchmark_TextRenderWithReader(b *testing.B) {
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)

	reader := strings.NewReader("Hello, world!")

	render := NewTextRender(response)
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		render.Render(reader)
	}
}

func Test_TextRenderWithStringify(t *testing.T) {
	assertion := assert.New(t)
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)
	response.Header().Add("Content-Type", "text/plain")

	// render with complex data type
	data := struct {
		Success bool
		Content string
	}{true, "Hello, world!"}

	render := NewTextRender(response)

	err := render.Render(data)
	assertion.Nil(err)
	assertion.Equal(`{true Hello, world!}`, recorder.Body.String())
}

func Benchmark_TextRenderWithStringify(b *testing.B) {
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)

	data := struct {
		Success bool
		Content string
	}{true, "Hello, world!"}

	render := NewTextRender(response)
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		render.Render(data)
	}
}

func Test_JsonRender(t *testing.T) {
	assertion := assert.New(t)
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)

	data := struct {
		Success bool   `json:"success"`
		Content string `json:"content"`
	}{true, "Hello, world!"}

	render := NewJsonRender(response)

	err := render.Render(data)
	assertion.Nil(err)
	assertion.Equal(http.StatusOK, recorder.Code)
	assertion.Equal("application/json", render.ContentType())
	assertion.Contains(recorder.Body.String(), `{"success":true,"content":"Hello, world!"}`)
}

func Benchmark_JsonRender(b *testing.B) {
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)

	data := struct {
		Success bool   `json:"success"`
		Content string `json:"content"`
	}{true, "Hello, world!"}

	render := NewJsonRender(response)
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		render.Render(data)
	}
}

func Test_XmlRender(t *testing.T) {
	assertion := assert.New(t)
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)

	data := struct {
		XMLName xml.Name `xml:"Response"`
		Success bool     `xml:"Result>Success"`
		Content string   `xml:"Result>Content"`
	}{
		Success: true,
		Content: "Hello, world!",
	}

	render := NewXmlRender(response)

	err := render.Render(data)
	assertion.Nil(err)
	assertion.Equal(http.StatusOK, recorder.Code)
	assertion.Equal("text/xml", render.ContentType())
	assertion.Equal("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<Response><Result><Success>true</Success><Content>Hello, world!</Content></Result></Response>", recorder.Body.String())
}

func Benchmark_XmlRender(b *testing.B) {
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)

	data := struct {
		XMLName xml.Name `xml:"Response"`
		Success bool     `xml:"Result>Success"`
		Content string   `xml:"Result>Content"`
	}{
		Success: true,
		Content: "Hello, world!",
	}

	render := NewXmlRender(response)
	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		render.Render(data)
	}
}

func Test_JsonpRender(t *testing.T) {
	assertion := assert.New(t)
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)

	data := struct {
		Success bool   `json:"success"`
		Content string `json:"content"`
	}{true, "Hello, world!"}

	render := NewJsonpRender(response, "js_callback")

	err := render.Render(data)
	assertion.Nil(err)
	assertion.Equal(http.StatusOK, recorder.Code)
	assertion.Equal("application/javascript", render.ContentType())
	assertion.Equal(`js_callback({"success":true,"content":"Hello, world!"});`, recorder.Body.String())
}

func Benchmark_JsonpRender(b *testing.B) {
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)

	data := struct {
		Success bool   `json:"success"`
		Content string `json:"content"`
	}{true, "Hello, world!"}

	render := NewJsonpRender(response, "js_callback")

	for i := 0; i < b.N; i++ {
		recorder.Body.Reset()

		render.Render(data)
	}
}
