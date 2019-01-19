package gogo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golib/assert"
)

var (
	fakeControllerHandler = &_fakeControllerHandler{}
	fakeController        = &_fakeController{}

	fakeGlobalHandler = func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("fakeGlobalHandler"))
	}

	fakeGlobalAction = func(ctx *Context) {
		ctx.Text("fakeGlobalAction")
	}
)

func fakePackageHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write(nil)
}

func fakePackageAction(ctx *Context) {
	ctx.Text("fakePackageAction")
}

type _fakeControllerHandler struct{}

func (_ *_fakeControllerHandler) Action(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("fakeControllerHandler"))
}

type _fakeController struct{}

func (_ *_fakeController) Action(ctx *Context) {
	ctx.Text("fakeController")
}

func Test_ContextHandle(t *testing.T) {
	it := assert.New(t)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "https://exmaple.com", nil)
	r = r.WithContext(context.WithValue(context.Background(), ctxLoggerKey, NewAppLogger("nil", "")))

	ch := NewContextHandle(fakeServer(), fakePackageHandler, nil)
	ch.Handle(w, r, nil)

	it.Equal(http.StatusNotImplemented, w.Code)
	it.Empty(w.Body.Bytes())
}

func Benchmark_ContextHandle(b *testing.B) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "https://exmaple.com", nil)
	r = r.WithContext(context.WithValue(context.Background(), ctxLoggerKey, NewAppLogger("nil", "")))

	ch := NewContextHandle(fakeServer(), fakePackageHandler, nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch.Handle(w, r, nil)
	}
}

func Test_ContextHandleWithHandler(t *testing.T) {
	it := assert.New(t)

	// global
	ch := NewContextHandle(nil, fakeGlobalHandler, nil)
	it.Equal("gogo", ch.pkg)
	it.Equal("gogo", ch.ctrl)
	it.Equal("<http.HandlerFunc>", ch.action)

	// package
	ch = NewContextHandle(nil, fakePackageHandler, nil)
	it.Equal("gogo", ch.pkg)
	it.Equal("gogo", ch.ctrl)
	it.Equal("fakePackageHandler", ch.action)

	// controller
	ch = NewContextHandle(nil, fakeControllerHandler.Action, nil)
	it.Equal("gogo", ch.pkg)
	it.Equal("*_fakeControllerHandler", ch.ctrl)
	it.Equal("Action", ch.action)
}

func Test_ContextHandleWithAction(t *testing.T) {
	it := assert.New(t)

	// global
	ch := NewContextHandle(nil, nil, []Middleware{fakeGlobalAction})
	it.Equal("gogo", ch.pkg)
	it.Equal("gogo", ch.ctrl)
	it.Equal("<http.HandlerFunc>", ch.action)

	// package
	ch = NewContextHandle(nil, nil, []Middleware{fakePackageAction})
	it.Equal("gogo", ch.pkg)
	it.Equal("gogo", ch.ctrl)
	it.Equal("fakePackageAction", ch.action)

	// controller
	ch = NewContextHandle(nil, nil, []Middleware{fakeController.Action})
	it.Equal("gogo", ch.pkg)
	it.Equal("*_fakeController", ch.ctrl)
	it.Equal("Action", ch.action)
}

func Test_FakeHandle(t *testing.T) {
	it := assert.New(t)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "https://exmaple.com", nil)
	r = r.WithContext(context.WithValue(context.Background(), ctxLoggerKey, NewAppLogger("nil", "")))

	fh := NewFakeHandle(fakeServer(), fakePackageHandler, nil, w)
	fh.Handle(nil, r, nil)

	it.Equal(http.StatusNotImplemented, w.Code)
	it.Empty(w.Body.Bytes())
}

func Test_NotFoundHandle(t *testing.T) {
	it := assert.New(t)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	r = r.WithContext(context.WithValue(context.Background(), ctxLoggerKey, NewAppLogger("nil", "")))

	h := NewNotFoundHandle(fakeServer())
	h.ServeHTTP(w, r)

	it.Equal(http.StatusNotFound, w.Code)
	it.Contains(w.Body.String(), "Request(GET /): not found")
}

func Benchmark_NotFoundHandle(b *testing.B) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	r = r.WithContext(context.WithValue(context.Background(), ctxLoggerKey, NewAppLogger("nil", "")))

	h := NewNotFoundHandle(fakeServer())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()

		h.ServeHTTP(w, r)
	}
}

func Test_MethodNotAllowedHandle(t *testing.T) {
	it := assert.New(t)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	r = r.WithContext(context.WithValue(context.Background(), ctxLoggerKey, NewAppLogger("nil", "")))

	h := NewMethodNotAllowedHandle(fakeServer())
	h.ServeHTTP(w, r)

	it.Equal(http.StatusMethodNotAllowed, w.Code)
	it.Contains(w.Body.String(), "Request(GET /): method not allowed")
}

func Benchmark_MethodNotAllowedHandle(b *testing.B) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	r = r.WithContext(context.WithValue(context.Background(), ctxLoggerKey, NewAppLogger("nil", "")))

	h := NewMethodNotAllowedHandle(fakeServer())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()

		h.ServeHTTP(w, r)
	}
}
