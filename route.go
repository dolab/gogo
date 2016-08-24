package gogo

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"time"

	"strings"

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
		ctx := r.server.new(resp, req, NewAppParams(req, params), handlers)
		ctx.Logger.Print("Started ", req.Method, " ", r.filterParameters(req.URL))

		ctx.Next()
		ctx.Response.FlushHeader()

		ctx.Logger.Print("Completed ", ctx.Response.Status(), " ", http.StatusText(ctx.Response.Status()), " in ", time.Since(ctx.startedAt))

		r.server.reuse(ctx)
	})
}

// ProxyHandle registers a new resource with its handler
func (r *AppRoute) ProxyHandle(method string, path string, proxy *httputil.ReverseProxy, filters ...ResponseFilter) {
	uri := r.calculatePrefix(path)
	handlers := r.combineHandlers(func(ctx *Context) {
		for _, filter := range filters {
			ctx.Response.Before(filter)
		}

		proxy.ServeHTTP(ctx.Response, ctx.Request)
	})

	handler := func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		ctx := r.server.new(resp, req, NewAppParams(req, params), handlers)
		ctx.Logger.Print("Started ", req.Method, " ", r.filterParameters(req.URL))

		ctx.Next()
		ctx.Response.FlushHeader()

		ctx.Logger.Print("Completed ", ctx.Response.Status(), " ", http.StatusText(ctx.Response.Status()), " in ", time.Since(ctx.startedAt))

		r.server.reuse(ctx)
	}

	if method == "Any" {
		r.server.router.Handle("GET", uri, handler)
		r.server.router.Handle("POST", uri, handler)
		r.server.router.Handle("PUT", uri, handler)
		r.server.router.Handle("PATCH", uri, handler)
		r.server.router.Handle("DELETE", uri, handler)
		r.server.router.Handle("HEAD", uri, handler)
		r.server.router.Handle("OPTIONS", uri, handler)
	} else {
		r.server.router.Handle(method, uri, handler)
	}
}

// MockHandle mocks a new resource with specified response and handler, useful for testing
func (r *AppRoute) MockHandle(method string, path string, response http.ResponseWriter, handler Middleware) {
	uri := r.calculatePrefix(path)
	handlers := r.combineHandlers(handler)

	r.server.router.Handle(method, uri, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		ctx := r.server.new(response, req, NewAppParams(req, params), handlers)
		ctx.Logger.Print("Started ", req.Method, " ", r.filterParameters(req.URL))

		ctx.Next()
		ctx.Response.FlushHeader()

		ctx.Logger.Print("Completed ", ctx.Response.Status(), " ", http.StatusText(ctx.Response.Status()), " in ", time.Since(ctx.startedAt))

		r.server.reuse(ctx)
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

// Router resource controller
// server := r.Resource("bucket", ServerCI) server is group
// GET /bucket/:id
// ip := server.Resource("ip", IpCI)
// SubResource
// r.Resource("parent", ParentResource).Resource("child", ChildResource)
// GET /bucket/:bucket/ip/:id
// note: parent's id key must not be the same with child (panic)
func (r *AppRoute) Resource(resource string, controller interface{}) *AppRoute {
	resource = strings.TrimSuffix(resource, "/")
	if resource[0] != '/' {
		resource = "/" + resource
	}

	// for common purpose
	var (
		resourceSpec string
		idSuffix     string
	)

	id, ok := controller.(ControllerID)
	if ok {
		idSuffix = strings.TrimSpace(id.Id())
	}

	// default id key
	if idSuffix == "" {
		idSuffix = "id"
	}

	resourceSpec = resource + "/:" + idSuffix

	index, ok := controller.(ControllerIndex)
	if ok {
		r.GET(resource, index.Index)
	}

	// for POST /resource
	create, ok := controller.(ControllerCreate)
	if ok {
		r.POST(resource, create.Create)
	}

	// for GET /resource/:id
	show, ok := controller.(ControllerShow)
	if ok {
		r.GET(resourceSpec, show.Show)
	}

	// for PUT /resource/:id
	update, ok := controller.(ControllerUpdate)
	if ok {
		r.PUT(resourceSpec, update.Update)
	}

	// for DELETE /resource/:id
	delete, ok := controller.(ControllerDestroy)
	if ok {
		r.DELETE(resourceSpec, delete.Destroy)
	}

	// for HEAD /resource/:id
	head, ok := controller.(ControllerExplore)
	if ok {
		r.HEAD(resourceSpec, head.Explore)
	}

	return r.Group(resourceSpec)
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

func (r *AppRoute) filterParameters(lru *url.URL) string {
	s := lru.Path

	query := lru.Query()
	if len(query) > 0 {
		for _, key := range r.server.filterParams {
			if _, ok := query[key]; ok {
				query.Set(key, "[FILTERED]")
			}
		}

		s += "?" + query.Encode()
	}

	return s
}
