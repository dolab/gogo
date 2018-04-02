package gogo

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"path"
	"strings"
	"sync"

	"github.com/dolab/httpdispatch"
)

// AppRoute defines route component of gogo
type AppRoute struct {
	mux sync.Mutex

	server      *AppServer
	handler     Handler
	prefix      string
	middlewares []Middleware
}

// NewAppRoute creates a new app route with specified prefix and server
func NewAppRoute(prefix string, server *AppServer) *AppRoute {
	route := &AppRoute{
		server: server,
		prefix: prefix,
	}

	// init default handler with httpdispatch.Router
	dispatcher := httpdispatch.New()
	dispatcher.RedirectTrailingSlash = false
	dispatcher.HandleMethodNotAllowed = false // strict for RESTful
	dispatcher.NotFound = http.HandlerFunc(route.notFound)
	dispatcher.MethodNotAllowed = http.HandlerFunc(route.methodNotAllowed)

	route.handler = dispatcher

	return route
}

// Group returns a new *AppRoute which has the same prefix path and middlewares
func (r *AppRoute) Group(prefix string, middlewares ...Middleware) *AppRoute {
	return &AppRoute{
		server:      r.server,
		handler:     r.handler,
		prefix:      r.calculatePrefix(prefix),
		middlewares: r.combineMiddlewares(middlewares...),
	}
}

// SetHandler replaces hanlder of AppRoute
func (r *AppRoute) SetHandler(handler Handler) {
	r.mux.Lock()
	r.handler = handler
	r.mux.Unlock()
}

// Use registers new middlewares to the route
// TODO: ignore duplicated middlewares?
func (r *AppRoute) Use(middlewares ...Middleware) {
	r.mux.Lock()
	r.middlewares = append(r.middlewares, middlewares...)
	r.mux.Unlock()
}

// Middlewares returns registered middlewares of AppRoute
func (r *AppRoute) Middlewares() []Middleware {
	return r.middlewares
}

// Clean removes all registered middlewares of AppRoute
// NOTE: it's useful in testing cases.
func (r *AppRoute) Clean() {
	r.middlewares = []Middleware{}
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

	r.handler.ServeFiles(path, http.Dir(root))
}

// ProxyHandle registers a new resource with a proxy
func (r *AppRoute) ProxyHandle(method string, path string, proxy *httputil.ReverseProxy, filters ...ResponseFilter) {
	handler := func(ctx *Context) {
		for _, filter := range filters {
			ctx.Response.Before(filter)
		}

		proxy.ServeHTTP(ctx.Response, ctx.Request)
	}

	switch method {
	case "*":
		r.Any(path, handler)

	default:
		r.Handle(method, path, handler)

	}
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

// HandlerFunc registers a new resource by http.HandlerFunc
func (r *AppRoute) HandlerFunc(method, uri string, handler http.HandlerFunc) {
	r.Handler(method, uri, handler)
}

// Handler registers a new resource by http.Handler
func (r *AppRoute) Handler(method, uri string, handler http.Handler) {
	uri = r.calculatePrefix(uri)
	middlewares := r.combineMiddlewares()

	r.handler.Handle(method, uri, NewContextHandle(r.server, handler.ServeHTTP, middlewares))
}

// ServeHTTP implements http.Handler by proxy to wrapped Handler
func (r *AppRoute) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(resp, req)
}

// Handle registers a new resource by Middleware
func (r *AppRoute) Handle(method string, uri string, handler Middleware) {
	uri = r.calculatePrefix(uri)
	middlewares := r.combineMiddlewares(handler)

	r.handler.Handle(method, uri, NewContextHandle(r.server, nil, middlewares))
}

// MockHandle mocks a new resource with specified response and handler, useful for testing
func (r *AppRoute) MockHandle(method string, path string, response http.ResponseWriter, handler Middleware) {
	uri := r.calculatePrefix(path)
	middlewares := r.combineMiddlewares(handler)

	r.handler.Handle(method, uri, NewFakeHandle(r.server, nil, middlewares, response))
}

func (r *AppRoute) combineMiddlewares(middlewares ...Middleware) []Middleware {
	combined := make([]Middleware, 0, len(r.middlewares)+len(middlewares))
	combined = append(combined, r.middlewares...)
	combined = append(combined, middlewares...)

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

func (r *AppRoute) notFound(resp http.ResponseWriter, req *http.Request) {
	r.server.logger.Print("Started ", req.Method, " ", req.URL)
	defer r.server.logger.Print("Completed ", http.StatusNotFound, " ", http.StatusText(http.StatusNotFound))

	http.Error(resp, fmt.Sprintf("Route(%s %s) not found", req.Method, req.URL.RequestURI()), http.StatusNotFound)
}

func (r *AppRoute) methodNotAllowed(resp http.ResponseWriter, req *http.Request) {
	r.server.logger.Print("Started ", req.Method, " ", req.URL)
	defer r.server.logger.Print("Completed ", http.StatusMethodNotAllowed, " ", http.StatusText(http.StatusMethodNotAllowed))

	http.Error(resp, fmt.Sprintf("Route(%s %s) not allowed", req.Method, req.URL.RequestURI()), http.StatusMethodNotAllowed)
}
