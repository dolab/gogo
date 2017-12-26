package gogo

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/golib/httprouter"
)

// AppRoute defines route component of gogo
type AppRoute struct {
	server   *AppServer
	prefix   string
	handlers []Middleware
}

// NewAppRoute creates a new app route with specified prefix and server
func NewAppRoute(prefix string, server *AppServer) *AppRoute {
	return &AppRoute{
		server: server,
		prefix: prefix,
	}
}

// Use registers new middlewares to the route
// TODO: ignore duplicated middlewares?
func (r *AppRoute) Use(middlewares ...Middleware) {
	r.handlers = append(r.handlers, middlewares...)
}

// Group returns a new app route group which has the same prefix path and middlewares
func (r *AppRoute) Group(prefix string, middlewares ...Middleware) *AppRoute {
	return &AppRoute{
		server:   r.server,
		prefix:   r.calculatePrefix(prefix),
		handlers: r.combineHandlers(middlewares...),
	}
}

// Handle registers a new resource with its handler
func (r *AppRoute) Handle(method string, path string, handler Middleware) {
	uri := r.calculatePrefix(path)
	handlers := r.combineHandlers(handler)

	r.server.handler.Handle(method, uri, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		ctx := r.server.newContext(resp, req, NewAppParams(req, params), handlers)

		ctx.Logger.Print("Started ", req.Method, " ", r.filterParameters(req.URL))
		defer func() {
			ctx.Logger.Print("Completed ", ctx.Response.Status(), " ", http.StatusText(ctx.Response.Status()), " in ", time.Since(ctx.startedAt))

			r.server.reuseContext(ctx)
		}()

		ctx.Next()
		ctx.Response.FlushHeader()
	})
}

// ProxyHandle registers a new resource with a proxy
func (r *AppRoute) ProxyHandle(method string, path string, proxy *httputil.ReverseProxy, filters ...ResponseFilter) {
	r.Handle(method, path, func(ctx *Context) {
		for _, filter := range filters {
			ctx.Response.Before(filter)
		}

		proxy.ServeHTTP(ctx.Response, ctx.Request)
	})
}

// MockHandle mocks a new resource with specified response and handler, useful for testing
func (r *AppRoute) MockHandle(method string, path string, response http.ResponseWriter, handler Middleware) {
	uri := r.calculatePrefix(path)
	handlers := r.combineHandlers(handler)

	r.server.handler.Handle(method, uri, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		ctx := r.server.newContext(response, req, NewAppParams(req, params), handlers)

		ctx.Logger.Print("Started ", req.Method, " ", r.filterParameters(req.URL))
		defer func() {
			ctx.Logger.Print("Completed ", ctx.Response.Status(), " ", http.StatusText(ctx.Response.Status()), " in ", time.Since(ctx.startedAt))

			r.server.reuseContext(ctx)
		}()

		ctx.Next()
		ctx.Response.FlushHeader()
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

// Resource generates routes with controller interfaces, and returns a group routes
// with resource name.
//
// Example:
//
// 	article := r.Resource("article")
// 		GET		/article			Article.Index
// 		POST	/article			Article.Create
// 		HEAD	/article/:article	Article.Explore
// 		GET		/article/:article	Article.Show
// 		PUT		/article/:article	Article.Update
// 		DELETE	/article/:article	Article.Destroy
//
func (r *AppRoute) Resource(resource string, controller interface{}) *AppRoute {
	resource = strings.TrimSuffix(resource, "/")
	if resource[0] != '/' {
		resource = "/" + resource
	}

	// for common purpose
	var (
		idSuffix     string
		resourceSpec string
	)

	id, ok := controller.(ControllerID)
	if ok {
		idSuffix = strings.TrimSpace(id.ID())
	}

	// default to resource name
	// NOTE: it's a trick for nested resource
	if idSuffix == "" {
		suffixes := strings.Split(strings.Trim(resource, "/"), "/")

		idSuffix = strings.ToLower(suffixes[len(suffixes)-1])
	}

	resourceSpec = resource + "/:" + idSuffix

	// for user-defined dispatch route
	dispatch, ok := controller.(ControllerDispatch)
	if ok {
		r.Any(resource, dispatch.DISPATCH)
		r.Any(resourceSpec, dispatch.DISPATCH)

		return r.Group(resourceSpec)
	}

	// for GET /resource
	index, ok := controller.(ControllerIndex)
	if ok {
		r.GET(resource, index.Index)
	}

	// for POST /resource
	create, ok := controller.(ControllerCreate)
	if ok {
		r.POST(resource, create.Create)
	}

	// for HEAD /resource/:resource
	head, ok := controller.(ControllerExplore)
	if ok {
		r.HEAD(resourceSpec, head.Explore)
	}

	// for GET /resource/:resource
	show, ok := controller.(ControllerShow)
	if ok {
		r.GET(resourceSpec, show.Show)
	}

	// for PUT /resource/:resource
	update, ok := controller.(ControllerUpdate)
	if ok {
		r.PUT(resourceSpec, update.Update)
	}

	// for DELETE /resource/:resource
	delete, ok := controller.(ControllerDestroy)
	if ok {
		r.DELETE(resourceSpec, delete.Destroy)
	}

	return r.Group(resourceSpec)
}

// Any is a shortcut for all request methods
func (r *AppRoute) Any(path string, handler Middleware) {
	r.Handle("GET", path, handler)
	r.Handle("POST", path, handler)
	r.Handle("PUT", path, handler)
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

	r.server.handler.ServeFiles(path, http.Dir(root))
}

func (r *AppRoute) combineHandlers(handlers ...Middleware) []Middleware {
	combined := make([]Middleware, 0, len(r.handlers)+len(handlers))
	combined = append(combined, r.handlers...)
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

func (r *AppRoute) notFoundHandle(resp http.ResponseWriter, req *http.Request) {
	r.server.logger.Print("Started ", req.Method, " ", req.URL)
	defer r.server.logger.Print("Completed ", http.StatusNotFound, " ", http.StatusText(http.StatusNotFound))

	http.NotFound(resp, req)
}

func (r *AppRoute) methodNotAllowed(resp http.ResponseWriter, req *http.Request) {
	r.server.logger.Print("Started ", req.Method, " ", req.URL)
	defer r.server.logger.Print("Completed ", http.StatusMethodNotAllowed, " ", http.StatusText(http.StatusMethodNotAllowed))

	http.Error(resp, "405 request method not allowed", http.StatusMethodNotAllowed)
}
