package gogo

import (
	"math"
	"net/http"
	"net/http/httputil"
	"path"
	"strings"
	"sync"

	"github.com/dolab/httpdispatch"
)

// AppGroup defines a group component of gogo
type AppGroup struct {
	mux sync.Mutex

	server  *AppServer
	prefix  string
	handler Handler
	filters []Middleware
}

// NewAppGroup creates a new router with specified prefix and server
func NewAppGroup(prefix string, server *AppServer) *AppGroup {
	// init default handler with httpdispatch.Router
	dispatcher := httpdispatch.New()
	dispatcher.RedirectTrailingSlash = false
	dispatcher.HandleMethodNotAllowed = false // strict for RESTful
	dispatcher.NotFound = NewNotFoundHandle(server)
	dispatcher.MethodNotAllowed = NewMethodNotAllowedHandle(server)

	return &AppGroup{
		server:  server,
		prefix:  prefix,
		handler: dispatcher,
	}
}

// NewGroup returns a new *AppGroup which has the same prefix path and middlewares
func (r *AppGroup) NewGroup(prefix string, middlewares ...Middleware) *AppGroup {
	return &AppGroup{
		server:  r.server,
		prefix:  r.buildPrefix(prefix),
		handler: r.handler,
		filters: r.buildFilters(middlewares...),
	}
}

// SetHandler replaces hanlder of AppGroup
func (r *AppGroup) SetHandler(handler Handler) {
	r.mux.Lock()
	r.handler = handler
	r.mux.Unlock()
}

// Use registers new middlewares to the route
// TODO: ignore duplicated middlewares?
func (r *AppGroup) Use(middlewares ...Middleware) {
	r.mux.Lock()
	defer r.mux.Unlock()

	r.filters = append(r.filters, middlewares...)
	if len(r.filters) >= math.MaxInt8 {
		panic(ErrTooManyMiddlewares)
	}
}

// Middlewares returns registered middlewares of AppGroup
func (r *AppGroup) Middlewares() []Middleware {
	return r.filters
}

// CleanMiddlewares removes all registered middlewares of AppGroup
// NOTE: it's useful in testing cases.
func (r *AppGroup) CleanMiddlewares() {
	r.mux.Lock()
	r.filters = []Middleware{}
	r.mux.Unlock()
}

// OPTIONS is a shortcut of group.Handle("OPTIONS", path, handler)
func (r *AppGroup) OPTIONS(rpath string, handler Middleware) {
	r.Handle("OPTIONS", rpath, handler)
}

// HEAD is a shortcut of group.Handle("HEAD", path, handler)
func (r *AppGroup) HEAD(rpath string, handler Middleware) {
	r.Handle("HEAD", rpath, handler)
}

// POST is a shortcut of group.Handle("POST", path, handler)
func (r *AppGroup) POST(rpath string, handler Middleware) {
	r.Handle("POST", rpath, handler)
}

// GET is a shortcut of group.Handle("GET", path, handler)
func (r *AppGroup) GET(rpath string, handler Middleware) {
	r.Handle("GET", rpath, handler)
}

// PUT is a shortcut of group.Handle("PUT", path, handler)
func (r *AppGroup) PUT(rpath string, handler Middleware) {
	r.Handle("PUT", rpath, handler)
}

// PATCH is a shortcut of group.Handle("PATCH", path, handler)
func (r *AppGroup) PATCH(rpath string, handler Middleware) {
	r.Handle("PATCH", rpath, handler)
}

// DELETE is a shortcut of group.Handle("DELETE", path, handler)
func (r *AppGroup) DELETE(rpath string, handler Middleware) {
	r.Handle("DELETE", rpath, handler)
}

// Any is a shortcut for all request methods
func (r *AppGroup) Any(rpath string, handler Middleware) {
	r.Handle("GET", rpath, handler)
	r.Handle("POST", rpath, handler)
	r.Handle("PUT", rpath, handler)
	r.Handle("PATCH", rpath, handler)
	r.Handle("DELETE", rpath, handler)
	r.Handle("HEAD", rpath, handler)
	r.Handle("OPTIONS", rpath, handler)
}

// Static serves files from the given dir
func (r *AppGroup) Static(rpath, root string) {
	if rpath[len(rpath)-1] != '/' {
		rpath += "/"
	}
	rpath += "*filepath"

	r.handler.ServeFiles(rpath, http.Dir(root))
}

// Proxy registers a new resource with a *httputil.ReverseProxy
func (r *AppGroup) Proxy(method string, rpath string, proxy *httputil.ReverseProxy, filters ...ResponseFilter) {
	handler := func(ctx *Context) {
		for _, filter := range filters {
			ctx.Response.Before(filter)
		}

		proxy.ServeHTTP(ctx.Response, ctx.Request)
	}

	switch method {
	case "*":
		r.Any(rpath, handler)

	default:
		r.Handle(method, rpath, handler)

	}
}

// Resource generates routes with controller interfaces, and returns a group routes
// with resource name for nested.
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
func (r *AppGroup) Resource(resource string, controller interface{}) *AppGroup {
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
		names := strings.Split(strings.Trim(resource, "/"), "/")

		idSuffix = strings.ToLower(names[len(names)-1])
	}

	resourceSpec = resource + "/:" + idSuffix

	// for user-defined dispatch route
	dispatch, ok := controller.(ControllerDispatch)
	if ok {
		r.Any(resource, dispatch.DISPATCH)
		r.Any(resourceSpec, dispatch.DISPATCH)

		return r.NewGroup(resourceSpec)
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

	return r.NewGroup(resourceSpec)
}

// HandlerFunc registers a new resource with http.HandlerFunc
func (r *AppGroup) HandlerFunc(method, uri string, handler http.HandlerFunc) {
	r.Handler(method, uri, handler)
}

// Handler registers a new resource with http.Handler
func (r *AppGroup) Handler(method, uri string, handler http.Handler) {
	uri = r.buildPrefix(uri)
	middlewares := r.buildFilters()

	r.handler.Handle(method, uri, NewContextHandle(r.server, handler.ServeHTTP, middlewares))
}

// Handle registers a new resource
func (r *AppGroup) Handle(method string, uri string, handler Middleware) {
	uri = r.buildPrefix(uri)
	middlewares := r.buildFilters(handler)

	r.handler.Handle(method, uri, NewContextHandle(r.server, nil, middlewares))
}

// MockHandle mocks a new resource with specified response and handler, useful for testing
func (r *AppGroup) MockHandle(method string, rpath string, recorder http.ResponseWriter, handler Middleware) {
	uri := r.buildPrefix(rpath)
	middlewares := r.buildFilters(handler)

	r.handler.Handle(method, uri, NewFakeHandle(r.server, nil, middlewares, recorder))
}

// Serve handles request by forwarding to underline Handler
func (r *AppGroup) Serve(resp http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(resp, req)
}

func (r *AppGroup) buildPrefix(suffix string) string {
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

func (r *AppGroup) buildFilters(middlewares ...Middleware) []Middleware {
	combined := make([]Middleware, len(r.filters)+len(middlewares))
	copy(combined[:len(r.filters)], r.filters)
	copy(combined[len(r.filters):], middlewares)

	return combined
}
