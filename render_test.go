package gogo

import (
	"crypto"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testRenderStatusCoder struct {
	code int
}

func (statusCoder *testRenderStatusCoder) StatusCode() int {
	return statusCoder.code
}

func Test_DefaultRender(t *testing.T) {
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)
	assertion := assert.New(t)

	// render with normal string
	render := NewDefaultRender(response)
	render.Render("Hello, world!")
	assertion.Equal(http.StatusOK, recorder.Code)
	assertion.Equal(render.ContentType(), recorder.Header().Get("Content-Type"))
	assertion.Equal("Hello, world!", recorder.Body.String())
}

func Test_DefaultRenderWithReader(t *testing.T) {
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)
	assertion := assert.New(t)

	s := strings.Repeat("Hello, world!", 0xffff)
	reader := strings.NewReader(s)

	response.Header().Add("Content-Length", fmt.Sprintf("%d", len(s)))
	response.Header().Add("Content-Type", "text/plain")

	// render with normal string
	render := NewDefaultRender(response)
	render.Render(reader)
	assertion.Equal(http.StatusOK, recorder.Code)
	assertion.Equal(render.ContentType(), recorder.Header().Get("Content-Type"))
	assertion.Equal(fmt.Sprintf("%d", len(s)), recorder.Header().Get("Content-Length"))
	assertion.Equal(s, recorder.Body.String())
}

func Benchmark_DefaultRenderWithReader(b *testing.B) {
	reader := strings.NewReader(strings.Repeat("Hello, world!", 0xffff))

	for i := 0; i < b.N; i++ {
		recorder := httptest.NewRecorder()
		response := NewResponse(recorder)

		render := NewDefaultRender(response)
		render.Render(reader)
	}
}

func Test_HashRender(t *testing.T) {
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)
	assertion := assert.New(t)

	// render with normal string
	render := NewHashRender(response, crypto.MD5)
	render.Render("Hello, world!")
	assertion.Equal(http.StatusOK, recorder.Code)
	assertion.Equal(render.ContentType(), recorder.Header().Get("Content-Type"))
	assertion.Equal("6cd3556deb0da54bca060b4c39479839", recorder.Header().Get("ETag"))
	assertion.Equal("Hello, world!", recorder.Body.String())
}

func Test_HashRenderWithReader(t *testing.T) {
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)
	assertion := assert.New(t)

	s := strings.Repeat("Hello, world!", 0xffff)
	reader := strings.NewReader(s)

	response.Header().Add("Content-Length", fmt.Sprintf("%d", len(s)))
	response.Header().Add("Content-Type", "text/plain")

	// render with normal string
	render := NewHashRender(response, crypto.MD5)
	render.Render(reader)
	assertion.Equal(http.StatusOK, recorder.Code)
	assertion.Equal(render.ContentType(), recorder.Header().Get("Content-Type"))
	assertion.Equal(fmt.Sprintf("%d", len(s)), recorder.Header().Get("Content-Length"))
	assertion.Equal("bc7f283e5babc5ed5e231c9bf961af1a", recorder.Header().Get("ETag"))
	assertion.Equal(s, recorder.Body.String())
}

func Benchmark_HashRenderWithReader(b *testing.B) {
	reader := strings.NewReader(strings.Repeat("Hello, world!", 0xffff))

	for i := 0; i < b.N; i++ {
		recorder := httptest.NewRecorder()
		response := NewResponse(recorder)

		render := NewHashRender(response, crypto.MD5)
		render.Render(reader)
	}
}

func Test_TextRender(t *testing.T) {
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)
	assertion := assert.New(t)

	// render with normal string
	render := NewTextRender(response)
	render.Render("Hello, world!")
	assertion.Equal(http.StatusOK, recorder.Code)
	assertion.Equal(render.ContentType(), recorder.Header().Get("Content-Type"))
	assertion.Equal("Hello, world!", recorder.Body.String())

	// render with complex data
	result := struct {
		Success bool
		Content string
	}{true, "Hello, world!"}

	recorder.Body.Reset()

	render.Render(result)
	assertion.Equal(fmt.Sprintf("%v", result), recorder.Body.String())
}

func Test_TextRenderWithoutContent(t *testing.T) {
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)
	assertion := assert.New(t)

	// render with normal string
	render := NewTextRender(response)
	render.Render("")
	// assertion.Equal(http.StatusNoContent, recorder.Code)
	assertion.Empty(recorder.Body.String())
}

func Test_TextRenderWithStatusCoder(t *testing.T) {
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)
	assertion := assert.New(t)

	// render with normal string
	render := NewTextRender(response)
	render.Render(&testRenderStatusCoder{http.StatusTeapot})
	assertion.Equal(http.StatusTeapot, recorder.Code)
}

func Test_JsonRender(t *testing.T) {
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)
	assertion := assert.New(t)

	result := struct {
		Success bool   `json:"success"`
		Content string `json:"content"`
	}{true, "Hello, world!"}

	render := NewJsonRender(response)
	render.Render(&result)
	assertion.Equal(http.StatusOK, recorder.Code)
	assertion.Equal(render.ContentType(), recorder.Header().Get("Content-Type"))
	assertion.Equal(`{"success":true,"content":"Hello, world!"}`, recorder.Body.String())
}

func Test_JsonRenderWithStatusCoder(t *testing.T) {
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)
	assertion := assert.New(t)

	// render with normal string
	render := NewJsonRender(response)
	render.Render(&testRenderStatusCoder{http.StatusTeapot})
	assertion.Equal(http.StatusTeapot, recorder.Code)
}

func Test_JsonpRender(t *testing.T) {
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)
	assertion := assert.New(t)

	result := struct {
		Success bool   `json:"success"`
		Content string `json:"content"`
	}{true, "Hello, world!"}

	render := NewJsonpRender(response, "js_callback")
	render.Render(&result)
	assertion.Equal(http.StatusOK, recorder.Code)
	assertion.Equal(render.ContentType(), recorder.Header().Get("Content-Type"))
	assertion.Equal(`js_callback({"success":true,"content":"Hello, world!"});`, recorder.Body.String())
}

func Test_JsonpRenderWithStatusCoder(t *testing.T) {
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)
	assertion := assert.New(t)

	// render with normal string
	render := NewJsonpRender(response, "js_callback")
	render.Render(&testRenderStatusCoder{http.StatusTeapot})
	assertion.Equal(http.StatusTeapot, recorder.Code)
}

func Test_XmlRender(t *testing.T) {
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)
	assertion := assert.New(t)

	result := struct {
		XMLName xml.Name `xml:"Response"`
		Success bool     `xml:"Result>Success"`
		Content string   `xml:"Result>Content"`
	}{
		Success: true,
		Content: "Hello, world!",
	}

	render := NewXmlRender(response)
	render.Render(&result)
	assertion.Equal(http.StatusOK, recorder.Code)
	assertion.Equal(render.ContentType(), recorder.Header().Get("Content-Type"))
	assertion.Equal("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<Response><Result><Success>true</Success><Content>Hello, world!</Content></Result></Response>", recorder.Body.String())
}

func Test_XmlRenderWithStatusCoder(t *testing.T) {
	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)
	assertion := assert.New(t)

	// render with normal string
	render := NewXmlRender(response)
	render.Render(&testRenderStatusCoder{http.StatusTeapot})
	assertion.Equal(http.StatusTeapot, recorder.Code)
}
