package gogo

import (
	"net/http"
	"time"

	"github.com/dolab/httpdispatch"
)

// ContextHandle wraps extra info for handler, such as package name, controller name and action name, etc.
type ContextHandle struct {
	pkg     string
	ctrl    string
	action  string
	server  *AppServer
	handler http.HandlerFunc
	filters []Middleware
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
