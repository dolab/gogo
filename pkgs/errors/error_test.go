package errors

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/golib/assert"
)

func Test_Error(t *testing.T) {
	it := assert.New(t)

	err := New("NotFound", "resource not found", nil)
	if it.NotNil(err) {
		it.Equal("NotFound", err.Code())
		it.Equal("resource not found", err.Message())
		it.Nil(err.OrigErr())
	}
}

func Test_RequestFailure(t *testing.T) {
	it := assert.New(t)

	origErr := New("NotFound", "resource not found", http.ErrHijacked)
	httpErr := NewRequestFailure(origErr, http.StatusNotFound, "request id")
	if it.NotNil(httpErr) {
		it.Equal("NotFound", httpErr.Code())
		it.Equal("resource not found", httpErr.Message())
		it.Equal(http.StatusNotFound, httpErr.StatusCode())
		it.Equal("request id", httpErr.RequestID())
		it.Equal(http.ErrHijacked, httpErr.OrigErr())
	}
}

func Test_WrappedRequestFailure(t *testing.T) {
	it := assert.New(t)

	httpErr := NewWrappedRequestFailure(http.StatusNotFound, "NotFound", "resource not found")
	if it.NotNil(httpErr) {
		it.Equal("NotFound", httpErr.Code())
		it.Equal("resource not found", httpErr.Message())
		it.Equal(http.StatusNotFound, httpErr.StatusCode())
		it.Empty(httpErr.RequestID())
		it.Nil(httpErr.OrigErr())
	}

	origErr := New("MethodNotAllowed", "resource method does not allowed", http.ErrHijacked)
	httpErr.WithStatusCode(http.StatusMethodNotAllowed)
	httpErr.WithRequestID("request id")
	httpErr.WithError(origErr)

	it.Equal(http.StatusMethodNotAllowed, httpErr.StatusCode())
	it.Equal("request id", httpErr.RequestID())
	it.Equal(origErr, httpErr.OrigErr())
}

func Test_SprintError_RequestFailure(t *testing.T) {
	it := assert.New(t)

	origErr := New("MethodNotAllowed", "resource method does not allowed", http.ErrHijacked)
	httpErr := NewRequestFailure(origErr, http.StatusNotFound, "request id")

	s := httpErr.Error()
	b, _ := json.Marshal(httpErr)
	it.Equal(s, string(b))
}

func Test_SprintError_WrappedRequestFailure(t *testing.T) {
	it := assert.New(t)

	origErr := New("MethodNotAllowed", "resource method does not allowed", http.ErrHijacked)
	httpErr := NewWrappedRequestFailure(http.StatusNotFound, "NotFound", "resource not found")
	httpErr.WithRequestID("request id")
	httpErr.WithError(origErr)

	s := httpErr.Error()
	b, _ := json.Marshal(httpErr)
	it.Equal(s, string(b))
}
