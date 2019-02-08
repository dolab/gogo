package render

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golib/assert"
)

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
