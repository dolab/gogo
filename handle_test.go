package gogo

import (
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
	w.Write([]byte("fakePackageHandler"))
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
	assertion := assert.New(t)

	r, _ := http.NewRequest(http.MethodGet, "https://exmaple.com", nil)
	w := httptest.NewRecorder()

	ch := NewContextHandle(fakeServer(), fakePackageHandler, nil)
	ch.Handle(w, r, nil)

	assertion.Equal("fakePackageHandler", w.Body.String())
}

func Test_ContextHandleWithHandler(t *testing.T) {
	assertion := assert.New(t)

	// global
	ch := NewContextHandle(nil, fakeGlobalHandler, nil)
	assertion.Equal("gogo", ch.pkg)
	assertion.Equal("gogo", ch.ctrl)
	assertion.Equal("<http.HandlerFunc>", ch.action)

	// package
	ch = NewContextHandle(nil, fakePackageHandler, nil)
	assertion.Equal("gogo", ch.pkg)
	assertion.Equal("gogo", ch.ctrl)
	assertion.Equal("fakePackageHandler", ch.action)

	// controller
	ch = NewContextHandle(nil, fakeControllerHandler.Action, nil)
	assertion.Equal("gogo", ch.pkg)
	assertion.Equal("*_fakeControllerHandler", ch.ctrl)
	assertion.Equal("Action", ch.action)
}

func Test_ContextHandleWithAction(t *testing.T) {
	assertion := assert.New(t)

	// global
	ch := NewContextHandle(nil, nil, []Middleware{fakeGlobalAction})
	assertion.Equal("gogo", ch.pkg)
	assertion.Equal("gogo", ch.ctrl)
	assertion.Equal("<http.HandlerFunc>", ch.action)

	// package
	ch = NewContextHandle(nil, nil, []Middleware{fakePackageAction})
	assertion.Equal("gogo", ch.pkg)
	assertion.Equal("gogo", ch.ctrl)
	assertion.Equal("fakePackageAction", ch.action)

	// controller
	ch = NewContextHandle(nil, nil, []Middleware{fakeController.Action})
	assertion.Equal("gogo", ch.pkg)
	assertion.Equal("*_fakeController", ch.ctrl)
	assertion.Equal("Action", ch.action)
}

func Test_FakeHandle(t *testing.T) {
	assertion := assert.New(t)

	r, _ := http.NewRequest(http.MethodGet, "https://exmaple.com", nil)
	w := httptest.NewRecorder()

	fh := NewFakeHandle(fakeServer(), fakePackageHandler, nil, w)
	fh.Handle(nil, r, nil)

	assertion.Equal("fakePackageHandler", w.Body.String())
}
