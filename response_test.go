package gogo

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golib/assert"
)

func Test_NewResponse(t *testing.T) {
	recorder := httptest.NewRecorder()
	assertion := assert.New(t)

	response := NewResponse(recorder)
	assertion.Implements((*Responser)(nil), response)
	assertion.Equal(http.StatusOK, response.Status())
	assertion.Equal(noneHeaderFlushed, response.Size())
}

func Test_ResponseWriteHeader(t *testing.T) {
	recorder := httptest.NewRecorder()
	assertion := assert.New(t)

	response := NewResponse(recorder)
	response.WriteHeader(http.StatusRequestTimeout)
	assertion.Equal(http.StatusRequestTimeout, response.Status())
	assertion.Equal(noneHeaderFlushed, response.Size())
}

func Test_ResponseFlushHeader(t *testing.T) {
	recorder := httptest.NewRecorder()
	assertion := assert.New(t)

	response := NewResponse(recorder)
	response.WriteHeader(http.StatusRequestTimeout)
	response.FlushHeader()
	assertion.Equal(http.StatusRequestTimeout, recorder.Code)
	assertion.NotEqual(noneHeaderFlushed, response.Size())

	// no effect after flushed headers
	response.WriteHeader(http.StatusOK)
	assertion.Equal(http.StatusOK, response.Status())
	response.FlushHeader()
	assertion.NotEqual(http.StatusOK, recorder.Code)
}

func Test_ResponseFulshHeaderWithFilters(t *testing.T) {
	counter := 0
	recorder := httptest.NewRecorder()
	filter1 := func(r Responser, b []byte) []byte {
		counter += 1

		return b
	}
	filter2 := func(r Responser, b []byte) []byte {
		counter += 2

		return b
	}
	assertion := assert.New(t)

	response := NewResponse(recorder)
	response.Before(filter1)
	response.Before(filter2)

	response.Write([]byte(""))
	assertion.Equal(3, counter)
}

func Test_ResponseWrite(t *testing.T) {
	recorder := httptest.NewRecorder()
	data := "Hello,world!"
	assertion := assert.New(t)

	response := NewResponse(recorder)
	response.Write([]byte("Hello,"))
	response.Write([]byte("world!"))
	assertion.Equal(len(data), response.Size())
	assertion.Equal(data, recorder.Body.String())
}
