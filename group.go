package gogo

import (
	"context"
	"math"
	"net/http"
	"net/http/httputil"
	"path"
	"strings"
	"sync"

	"github.com/dolab/gogo/pkgs/gid"
	"github.com/dolab/httpdispatch"
)

// AppGroup defines a grouped server of gogo.
type AppGroup struct {
	mux sync.Mutex

	server  *AppServer
	prefix  string
	filters []FilterFunc
	handler Handler
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

// NewGroup returns a new *AppGroup which has the same prefix path and filters
func (r *AppGroup) NewGroup(prefix string, filters ...FilterFunc) Grouper {
	r.mux.Lock()
	defer r.mux.Unlock()

	return &AppGroup{
		server:  r.server,
		prefix:  r.buildPrefix(prefix),
		handler: r.handler,
		filters: r.buildFilters(filters...),
	}
}

// SetHandler replaces hanlder of AppGroup
func (r *AppGroup) SetHandler(handler Handler) {
	r.mux.Lock()
	r.handler = handler
	r.mux.Unlock()
}

// Use appends new filters to the end of group
//
// TODO: ignore duplicated filters?
func (r *AppGroup) Use(filters ...FilterFunc) {
	r.mux.Lock()
	defer r.mux.Unlock()

	r.filters = append(r.filters, filters...)
	if len(r.filters) >= math.MaxInt8 {
		panic(ErrTooManyFilters)
	}
}

// Filters returns all filters of AppGroup
func (r *AppGroup) Filters() []FilterFunc {
	return r.filters
}

// CleanFilters removes all registered filters of AppGroup
//
// NOTE: It's useful in testing cases.
func (r *AppGroup) CleanFilters() {
	r.mux.Lock()
	r.filters = []FilterFunc{}
	r.mux.Unlock()
}

// OPTIONS is a shortcut of group.Handle("OPTIONS", path, handler)
func (r *AppGroup) OPTIONS(rpath string, handler FilterFunc) {
	r.Handle("OPTIONS", rpath, handler)
}

// HEAD is a shortcut of group.Handle("HEAD", path, handler)
func (r *AppGroup) HEAD(rpath string, handler FilterFunc) {
	r.Handle("HEAD", rpath, handler)
}

// POST is a shortcut of group.Handle("POST", path, handler)
func (r *AppGroup) POST(rpath string, handler FilterFunc) {
	r.Handle("POST", rpath, handler)
}

// GET is a shortcut of group.Handle("GET", path, handler)
func (r *AppGroup) GET(rpath string, handler FilterFunc) {
	r.Handle("GET", rpath, handler)
}

// PUT is a shortcut of group.Handle("PUT", path, handler)
func (r *AppGroup) PUT(rpath string, handler FilterFunc) {
	r.Handle("PUT", rpath, handler)
}

// PATCH is a shortcut of group.Handle("PATCH", path, handler)
func (r *AppGroup) PATCH(rpath string, handler FilterFunc) {
	r.Handle("PATCH", rpath, handler)
}

// DELETE is a shortcut of group.Handle("DELETE", path, handler)
func (r *AppGroup) DELETE(rpath string, handler FilterFunc) {
	r.Handle("DELETE", rpath, handler)
}

// Any is a shortcut for all request methods
func (r *AppGroup) Any(rpath string, handler FilterFunc) {
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
//
// NOTE: filters is deprecated!
func (r *AppGroup) Proxy(method string, rpath string, proxy *httputil.ReverseProxy) {
	handler := func(ctx *Context) {
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
func (r *AppGroup) Resource(resource string, controller interface{}) Grouper {
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
	filters := r.buildFilters()

	r.handler.Handle(method, uri, NewContextHandle(
		handler.ServeHTTP, filters,
		r.server.RequestRouted, r.server.ResponseReady, r.server.ResponseAlways,
	))
}

// Handle registers a new resource
func (r *AppGroup) Handle(method string, uri string, handler FilterFunc) {
	uri = r.buildPrefix(uri)
	filters := r.buildFilters(handler)

	r.handler.Handle(method, uri, NewContextHandle(
		nil, filters,
		r.server.RequestRouted, r.server.ResponseReady, r.server.ResponseAlways,
	))
}

// MockHandle mocks a new resource with specified response and handler, useful for testing
func (r *AppGroup) MockHandle(method string, rpath string, recorder http.ResponseWriter, handler FilterFunc) {
	uri := r.buildPrefix(rpath)
	filters := r.buildFilters(handler)

	r.handler.Handle(method, uri, NewFakeHandle(
		nil, filters, recorder,
		r.server.RequestRouted, r.server.ResponseReady, r.server.ResponseAlways,
	))
}

// ServeHTTP implements the http.Handler interface
//
// NOTE: ServeHTTP handles request by forwarding to underline Handler
func (r *AppGroup) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	// hijack request id if required
	var logID string
	if r.server.hasRequestID() {
		logID = req.Header.Get(r.server.requestID)
		if logID == "" || len(logID) > DefaultRequestIDMaxLen {
			logID = gid.New().Hex()

			// inject request header with new request id
			req.Header.Set(r.server.requestID, logID)
		}

		resp.Header().Set(r.server.requestID, logID)
	}

	// hijack logger for request
	log := r.server.loggerNew(logID)
	defer r.server.loggerReuse(log)

	req = req.WithContext(context.WithValue(req.Context(), ctxLoggerKey, log))

	// invoke RequestReceived
	if !r.server.RequestReceived.Run(resp, req) {
		return
	}

	r.handler.ServeHTTP(resp, req)
}

func (r *AppGroup) buildPrefix(suffix string) (prefix string) {
	defer func() {
		// assert for internal routes
		if strings.HasPrefix(prefix, GogoHealthz) {
			panic(ErrReservedRoute)
		}
	}()

	if len(suffix) == 0 {
		prefix = r.prefix
		return
	}

	prefix = path.Join(r.prefix, suffix)

	// adjust path.Join side effect
	if suffix[len(suffix)-1] == '/' && prefix[len(prefix)-1] != '/' {
		prefix += "/"
	}

	return prefix
}

func (r *AppGroup) buildFilters(filters ...FilterFunc) []FilterFunc {
	combined := make([]FilterFunc, len(r.filters)+len(filters))
	copy(combined[:len(r.filters)], r.filters)
	copy(combined[len(r.filters):], filters)

	return combined
}

func (r *AppGroup) registerHealthz() {
	handler := NewHealthzHandle(r.server)

	r.handler.Handle(http.MethodGet, GogoHealthz, handler)
	r.handler.Handle(http.MethodPost, GogoHealthz, handler)
}
