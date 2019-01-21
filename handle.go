package gogo

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"github.com/dolab/gogo/internal/params"
	"github.com/dolab/httpdispatch"
)

// ContextHandle wraps handler with extra metadata, such as package name, controller name and action name, etc.
type ContextHandle struct {
	pkg     string
	ctrl    string
	action  string
	server  *AppServer
	handler http.HandlerFunc
	filters []Middleware
}

// NewContextHandle returns new *ContextHandle with handler and metadata
func NewContextHandle(server *AppServer, handler http.HandlerFunc, filters []Middleware) *ContextHandle {
	var rval reflect.Value

	if handler == nil {
		rval = reflect.ValueOf(filters[len(filters)-1])
	} else {
		rval = reflect.ValueOf(handler)
	}

	// formated in "/path/to/main.(*_Controller).Action-fm"
	name := runtime.FuncForPC(rval.Pointer()).Name()

	vars := strings.Split(name, "/")
	if len(vars) > 1 {
		name = vars[len(vars)-1]
	}

	vars = strings.Split(name, ".")
	switch len(vars) {
	case 3:
		// adjust controller name
		vars[1] = strings.TrimLeft(vars[1], "(")
		vars[1] = strings.TrimRight(vars[1], ")")

		// adjust action name
		vars[2] = strings.SplitN(vars[2], "-", 2)[0]

	case 2:
		// package func
		vars = []string{vars[0], vars[0], vars[1]}

	default:
		vars = []string{vars[0], vars[0], "<http.HandlerFunc>"}
	}

	return &ContextHandle{
		pkg:     vars[0],
		ctrl:    vars[1],
		action:  vars[2],
		server:  server,
		handler: handler,
		filters: filters,
	}
}

// Handle implements httpdispatch.Handler interface
func (ch *ContextHandle) Handle(w http.ResponseWriter, r *http.Request, ps httpdispatch.Params) {
	// invoke ResponseAlways
	defer ch.server.hooks.ResponseAlways.Run(w, r)

	if !ch.server.hooks.RequestRouted.Run(w, r) {
		return
	}

	ctx := contextNew(w, r, params.NewParams(r, ps), ch.pkg, ch.ctrl, ch.action)
	defer contextReuse(ctx)

	if ch.handler == nil {
		ctx.run(nil, ch.filters)
	} else {
		ctx.run(ch.handler, ch.filters)
	}
}

// FakeHandle defines a wrapper of handler for testing
//
// NOTE: DO NOT use this for real!!!
type FakeHandle struct {
	*ContextHandle

	recorder http.ResponseWriter
}

// NewFakeHandle returns new handler with stubbed http.ResponseWriter
func NewFakeHandle(server *AppServer, handler http.HandlerFunc, filters []Middleware, recorder http.ResponseWriter) *FakeHandle {
	ch := &FakeHandle{
		ContextHandle: NewContextHandle(server, handler, filters),
		recorder:      recorder,
	}

	return ch
}

// Handle implements httpdispatch.Handler interface
func (ch *FakeHandle) Handle(w http.ResponseWriter, r *http.Request, params httpdispatch.Params) {
	ch.ContextHandle.Handle(ch.recorder, r, params)
}

// NotFoundHandle defines a wrapper of handler for route not found
type NotFoundHandle struct {
	*ContextHandle
}

// NewNotFoundHandle creates a new handler with route not found
func NewNotFoundHandle(server *AppServer) *NotFoundHandle {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, fmt.Sprintf("Request(%s %s): not found", r.Method, r.URL.RequestURI()), http.StatusNotFound)
	})

	return &NotFoundHandle{
		ContextHandle: NewContextHandle(server, handler, nil),
	}
}

// ServeHTTP implements http.Handler interface
func (h *NotFoundHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.ContextHandle.Handle(w, r, nil)
}

// MethodNotAllowedHandle defines a wrapper of handler for not allowed request mehtod
type MethodNotAllowedHandle struct {
	*ContextHandle
}

// NewMethodNotAllowedHandle creates a new handler with request method not allowed
func NewMethodNotAllowedHandle(server *AppServer) *MethodNotAllowedHandle {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, fmt.Sprintf("Request(%s %s): method not allowed", r.Method, r.URL.RequestURI()), http.StatusMethodNotAllowed)
	})

	return &MethodNotAllowedHandle{
		ContextHandle: NewContextHandle(server, handler, nil),
	}
}

// ServeHTTP implements http.Handler interface
func (h *MethodNotAllowedHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.ContextHandle.Handle(w, r, nil)
}
