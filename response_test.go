package gogo

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golib/assert"
)

func Test_NewResponse(t *testing.T) {
	it := assert.New(t)
	recorder := httptest.NewRecorder()

	response := NewResponse(recorder)
	it.Implements((*Responser)(nil), response)
	it.Equal(http.StatusOK, response.Status())
	it.Equal(nonHeaderFlushed, response.Size())
}

func Test_ResponseWriteHeader(t *testing.T) {
	it := assert.New(t)
	recorder := httptest.NewRecorder()

	response := NewResponse(recorder)
	response.WriteHeader(http.StatusRequestTimeout)
	it.Equal(http.StatusRequestTimeout, response.Status())
	it.Equal(nonHeaderFlushed, response.Size())
}

func Test_ResponseFlushHeader(t *testing.T) {
	it := assert.New(t)
	recorder := httptest.NewRecorder()

	response := NewResponse(recorder)
	response.WriteHeader(http.StatusRequestTimeout)
	response.FlushHeader()
	it.Equal(http.StatusRequestTimeout, recorder.Code)
	it.NotEqual(nonHeaderFlushed, response.Size())

	// no effect after flushed headers
	response.WriteHeader(http.StatusOK)
	it.Equal(http.StatusOK, response.Status())
	response.FlushHeader()
	it.NotEqual(http.StatusOK, recorder.Code)
}

func Test_ResponseWrite(t *testing.T) {
	it := assert.New(t)
	recorder := httptest.NewRecorder()
	hello := []byte("Hello,")
	world := []byte("world!")
	expected := []byte("Hello,world!")

	response := NewResponse(recorder)
	response.Write(hello)
	response.Write(world)
	it.True(response.HeaderFlushed())
	it.Equal(len(expected), response.Size())
	it.Equal(expected, recorder.Body.Bytes())
}

func Benchmark_ResponseWrite(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	hello := []byte("Hello,")
	world := []byte("world!")

	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)
	for i := 0; i < b.N; i++ {
		response.Write(hello)
		response.Write(world)
	}
}

func Test_ResponseHijack(t *testing.T) {
	it := assert.New(t)
	recorder := httptest.NewRecorder()
	expected := []byte("Hello,world!")

	response := NewResponse(recorder)
	response.Write(expected)
	it.True(response.HeaderFlushed())
	it.Equal(recorder, response.(*Response).ResponseWriter)
	it.Equal(len(expected), response.Size())

	response.Hijack(httptest.NewRecorder())
	it.False(response.HeaderFlushed())
	it.NotEqual(recorder, response.(*Response).ResponseWriter)
	it.Equal(nonHeaderFlushed, response.Size())
}

func Benchmark_ResponseHijack(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)
	for i := 0; i < b.N; i++ {
		response.Hijack(recorder)
	}
}
