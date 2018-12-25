package gogo

import (
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"time"

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

	tmpvars := strings.Split(name, "/")
	if len(tmpvars) > 1 {
		name = tmpvars[len(tmpvars)-1]
	}

	tmpvars = strings.Split(name, ".")
	switch len(tmpvars) {
	case 3:
		// adjust controller name
		tmpvars[1] = strings.TrimLeft(tmpvars[1], "(")
		tmpvars[1] = strings.TrimRight(tmpvars[1], ")")

		// adjust action name
		tmpvars[2] = strings.SplitN(tmpvars[2], "-", 2)[0]

	case 2:
		// package func
		tmpvars = []string{tmpvars[0], tmpvars[0], tmpvars[1]}

	default:
		tmpvars = []string{tmpvars[0], tmpvars[0], "<http.HandlerFunc>"}
	}

	return &ContextHandle{
		pkg:     tmpvars[0],
		ctrl:    tmpvars[1],
		action:  tmpvars[2],
		server:  server,
		handler: handler,
		filters: filters,
	}
}

// Handle implements httpdispatch.Handler interface
func (ch *ContextHandle) Handle(w http.ResponseWriter, r *http.Request, params httpdispatch.Params) {
	ctx := ch.server.newContext(r, ch.ctrl, ch.action, NewAppParams(r, params))

	ctx.Logger.Print("Started ", r.Method, " ", ch.server.filterParameters(r.URL))
	defer func() {
		ctx.Logger.Print("Completed ", ctx.Response.Status(), " ", http.StatusText(ctx.Response.Status()), " in ", time.Since(ctx.startedAt))

		ch.server.reuseContext(ctx)
	}()

	w.Header().Set(ch.server.requestID, ctx.RequestID())

	if ch.handler == nil {
		ctx.run(w, nil, ch.filters)
	} else {
		ctx.run(w, ch.handler, ch.filters)
	}
}

// FakeHandle defines a wrapper of handler for testing
type FakeHandle struct {
	*ContextHandle

	w http.ResponseWriter
}

// NewFakeHandle returns new handler with stubbed http.ResponseWriter
func NewFakeHandle(server *AppServer, handler http.HandlerFunc, filters []Middleware, w http.ResponseWriter) *FakeHandle {
	ch := &FakeHandle{
		ContextHandle: NewContextHandle(server, handler, filters),
		w:             w,
	}

	return ch
}

// Handle implements httpdispatch.Handler interface
func (ch *FakeHandle) Handle(w http.ResponseWriter, r *http.Request, params httpdispatch.Params) {
	ctx := ch.server.newContext(r, ch.ctrl, ch.action, NewAppParams(r, params))

	ctx.Logger.Print("Started ", r.Method, " ", ch.server.filterParameters(r.URL))
	defer func() {
		ctx.Logger.Print("Completed ", ctx.Response.Status(), " ", http.StatusText(ctx.Response.Status()), " in ", time.Since(ctx.startedAt))

		ch.server.reuseContext(ctx)
	}()

	ch.w.Header().Set(ch.server.requestID, ctx.RequestID())

	if ch.handler == nil {
		ctx.run(ch.w, nil, ch.filters)
	} else {
		ctx.run(ch.w, ch.handler, ch.filters)
	}
}
