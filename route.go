package gogo

import (
	"net/http"
	"path"

	"github.com/golib/httprouter"
)

type AppRoute struct {
	Handlers []Middleware

	prefix string
	server *AppServer
}

// NewAppRoute creates a new app route with specified prefix and server
func NewAppRoute(prefix string, server *AppServer) *AppRoute {
	return &AppRoute{
		prefix: prefix,
		server: server,
	}
}

// Use registers new middlewares to the route
// TODO: ignore duplicated middlewares?
func (r *AppRoute) Use(middlewares ...Middleware) {
	r.Handlers = append(r.Handlers, middlewares...)
}

// Group returns a new app route group which has the same prefix path and middlewares
func (r *AppRoute) Group(prefix string, middlewares ...Middleware) *AppRoute {
	return &AppRoute{
		Handlers: r.combineHandlers(middlewares...),
		prefix:   r.calculatePrefix(prefix),
		server:   r.server,
	}
}

// Handle registers a new resource with its handler
func (r *AppRoute) Handle(method string, path string, handler Middleware) {
	uri := r.calculatePrefix(path)
	handlers := r.combineHandlers(handler)

	r.server.router.Handle(method, uri, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		ctx := r.server.New(resp, req, NewAppParams(req, params), handlers)
		ctx.Next()
		ctx.Response.FlushHeader()

		r.server.Reuse(ctx)
	})
}

// MockHandle mocks a new resource with specified response and handler, useful for testing
func (r *AppRoute) MockHandle(method string, path string, response http.ResponseWriter, handler Middleware) {
	uri := r.calculatePrefix(path)
	handlers := r.combineHandlers(handler)

	r.server.router.Handle(method, uri, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		ctx := r.server.New(response, req, NewAppParams(req, params), handlers)
		ctx.Next()
		ctx.Response.FlushHeader()

		r.server.Reuse(ctx)
	})
}

// PUT is a shortcut of route.Handle("PUT", path, handler)
func (r *AppRoute) PUT(path string, handler Middleware) {
	r.Handle("PUT", path, handler)
}

// POST is a shortcut of route.Handle("POST", path, handler)
func (r *AppRoute) POST(path string, handler Middleware) {
	r.Handle("POST", path, handler)
}

// GET is a shortcut of route.Handle("GET", path, handler)
func (r *AppRoute) GET(path string, handler Middleware) {
	r.Handle("GET", path, handler)
}

// PATCH is a shortcut of route.Handle("PATCH", path, handler)
func (r *AppRoute) PATCH(path string, handler Middleware) {
	r.Handle("PATCH", path, handler)
}

// DELETE is a shortcut of route.Handle("DELETE", path, handler)
func (r *AppRoute) DELETE(path string, handler Middleware) {
	r.Handle("DELETE", path, handler)
}

// HEAD is a shortcut of route.Handle("HEAD", path, handler)
func (r *AppRoute) HEAD(path string, handler Middleware) {
	r.Handle("HEAD", path, handler)
}

// OPTIONS is a shortcut of route.Handle("OPTIONS", path, handler)
func (r *AppRoute) OPTIONS(path string, handler Middleware) {
	r.Handle("OPTIONS", path, handler)
}

// Any is a shortcut for all request methods
func (r *AppRoute) Any(path string, handler Middleware) {
	r.Handle("PUT", path, handler)
	r.Handle("POST", path, handler)
	r.Handle("GET", path, handler)
	r.Handle("PATCH", path, handler)
	r.Handle("DELETE", path, handler)
	r.Handle("HEAD", path, handler)
	r.Handle("OPTIONS", path, handler)
}

// Static serves files from the given dir
func (r *AppRoute) Static(path, root string) {
	if path[len(path)-1] != '/' {
		path += "/"
	}
	path += "*filepath"

	r.server.router.ServeFiles(path, http.Dir(root))
}

func (r *AppRoute) combineHandlers(handlers ...Middleware) []Middleware {
	combined := make([]Middleware, 0, len(r.Handlers)+len(handlers))
	combined = append(combined, r.Handlers...)
	combined = append(combined, handlers...)

	return combined
}

func (r *AppRoute) calculatePrefix(suffix string) string {
	if len(suffix) == 0 {
		return r.prefix
	}

	prefix := path.Join(r.prefix, suffix)

	// adjust path.Join side effect
	if suffix[len(suffix)-1] == '/' && prefix[len(prefix)-1] != '/' {
		prefix += "/"
	}

	return prefix
}
