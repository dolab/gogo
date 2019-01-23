package gogo

import (
	"net/http"
	"net/http/httputil"

	"github.com/dolab/httpdispatch"
)

// A Configer represents config
type Configer interface {
	RunMode() RunMode
	RunName() string
	SetMode(mode RunMode)
	Section() *SectionConfig
	UnmarshalJSON(v interface{}) error
}

// A Grouper represents routers
type Grouper interface {
	NewGroup(prefix string, middlewares ...Middleware) Grouper
	Resource(uri string, resource interface{}) Grouper
	SetHandler(handler Handler)
	Use(middlewares ...Middleware)
	OPTIONS(uri string, action Middleware)
	HEAD(uri string, action Middleware)
	POST(uri string, action Middleware)
	GET(uri string, action Middleware)
	PUT(uri string, action Middleware)
	PATCH(uri string, action Middleware)
	DELETE(uri string, action Middleware)
	Any(uri string, action Middleware)
	Static(uri, root string)
	Proxy(method, uri string, proxy *httputil.ReverseProxy)
	HandlerFunc(method, uri string, fn http.HandlerFunc)
	Handler(method, uri string, handler http.Handler)
	Handle(method, uri string, action Middleware)
	MockHandle(method, uri string, recorder http.ResponseWriter, action Middleware)
}

// A Servicer represents application
type Servicer interface {
	Init(config Configer, group Grouper)
	Middlewares()
	Resources()
}

// A Handler represents handlers
type Handler interface {
	http.Handler

	Handle(string, string, httpdispatch.Handler)
	ServeFiles(string, http.FileSystem)
}

// Middleware represents request filters and resource handler
//
// NOTE: It is the filter's responsibility to invoke ctx.Next() for chainning.
type Middleware func(ctx *Context)

// Responser represents HTTP response interface
type Responser interface {
	http.ResponseWriter
	http.Flusher

	HeaderFlushed() bool        // whether response header has been sent?
	FlushHeader()               // send response header only if it has not sent
	Status() int                // response status code
	Size() int                  // return the size of response body
	Hijack(http.ResponseWriter) // hijack response with new http.ResponseWriter
}

// StatusCoder represents HTTP response status code
// it is useful for custom response data with response status code
type StatusCoder interface {
	StatusCode() int
}

// Logger defines interface of application log apis.
type Logger interface {
	New(requestID string) Logger
	Reuse(l Logger)
	RequestID() string
	SetLevelByName(level string) error
	SetColor(color bool)

	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Warn(v ...interface{})
	Warnf(format string, v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Panic(v ...interface{})
	Panicf(format string, v ...interface{})
}
